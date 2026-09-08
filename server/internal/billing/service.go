package billing

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/store"
)

var (
	ErrInvalidPlan      = errors.New("invalid plan")
	ErrDowngradeBlocked = errors.New("downgrade not supported yet")
	ErrSamePlan         = errors.New("already on this plan")
)

type Service struct {
	subs      *store.Subscriptions
	checkouts *store.BillingCheckouts
	events    *store.BillingEvents
	accounts  *store.Accounts
	providers map[string]Provider
}

func NewService(
	subs *store.Subscriptions,
	checkouts *store.BillingCheckouts,
	events *store.BillingEvents,
	accounts *store.Accounts,
	providers []Provider,
) *Service {
	m := make(map[string]Provider, len(providers))
	for _, p := range providers {
		m[p.Name()] = p
	}
	return &Service{
		subs:      subs,
		checkouts: checkouts,
		events:    events,
		accounts:  accounts,
		providers: m,
	}
}

type Overview struct {
	Plan              string      `json:"plan"`
	PlanLabel         string      `json:"plan_label"`
	MaxProjects       int         `json:"max_projects"`
	Status            string      `json:"status"`
	Provider          string      `json:"provider"`
	PeriodEnd         *string     `json:"period_end,omitempty"`
	Plans             []PlanOffer `json:"plans"`
	Providers         []ProviderInfo `json:"providers"`
	ManualUpgrade     bool        `json:"manual_upgrade_enabled"`
}

type ProviderInfo struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Available bool   `json:"available"`
	Region    string `json:"region"`
}

type UpgradeInput struct {
	Plan      string
	Provider  string
	ReturnURL string
	Email     string
}

type UpgradeResult struct {
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url,omitempty"`
	Message     string `json:"message,omitempty"`
	Plan        string `json:"plan,omitempty"`
	PlanLabel   string `json:"plan_label,omitempty"`
}

func (s *Service) Overview(ctx context.Context, account *store.Account, manualUpgrade bool) (*Overview, error) {
	if account == nil {
		return nil, errors.New("account not found")
	}

	sub, err := s.subs.EnsureForAccount(ctx, account.ID, account.Plan)
	if err != nil {
		return nil, err
	}

	status := "active"
	provider := "manual"
	var periodEnd *string
	if sub != nil {
		status = sub.Status
		provider = sub.Provider
		if sub.CurrentPeriodEnd != nil {
			raw := sub.CurrentPeriodEnd.UTC().Format("2006-01-02")
			periodEnd = &raw
		}
	}

	return &Overview{
		Plan:          account.Plan,
		PlanLabel:     PlanLabel(account.Plan),
		MaxProjects:   MaxProjects(account.Plan),
		Status:        status,
		Provider:      provider,
		PeriodEnd:     periodEnd,
		Plans:         AllPlans(),
		Providers:     s.providerList(),
		ManualUpgrade: manualUpgrade,
	}, nil
}

func (s *Service) providerList() []ProviderInfo {
	return []ProviderInfo{
		{ID: "paytr", Label: "PayTR", Available: s.providers["paytr"] != nil && s.providers["paytr"].Available(), Region: "TR"},
		{ID: "stripe", Label: "Stripe", Available: s.providers["stripe"] != nil && s.providers["stripe"].Available(), Region: "global"},
		{ID: "manual", Label: "Manual", Available: s.providers["manual"] != nil && s.providers["manual"].Available(), Region: "self-hosted"},
	}
}

func (s *Service) Upgrade(ctx context.Context, ownerID uuid.UUID, in UpgradeInput) (*UpgradeResult, error) {
	if !ValidPlan(in.Plan) {
		return nil, ErrInvalidPlan
	}

	account, err := s.accounts.FindOrCreateByOwner(ctx, ownerID)
	if err != nil || account == nil {
		return nil, errors.New("account not found")
	}

	if account.Plan == in.Plan {
		return nil, ErrSamePlan
	}
	if PlanRank(in.Plan) < PlanRank(account.Plan) {
		return nil, ErrDowngradeBlocked
	}

	providerName := in.Provider
	if providerName == "" {
		providerName = "paytr"
	}

	if in.Plan == PlanFree {
		if err := s.applyPlan(ctx, account.ID, in.Plan, "manual", ""); err != nil {
			return nil, err
		}
		return &UpgradeResult{
			Status:    "completed",
			Plan:      in.Plan,
			PlanLabel: PlanLabel(in.Plan),
		}, nil
	}

	provider, ok := s.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("unknown billing provider: %s", providerName)
	}

	quote := priceForPlan(in.Plan, providerName)
	checkout, err := s.checkouts.Create(
		ctx, account.ID, in.Plan, providerName,
		quote.AmountCents, quote.Currency, in.ReturnURL,
	)
	if err != nil {
		return nil, err
	}

	result, err := provider.StartCheckout(ctx, CheckoutRequest{
		CheckoutID:    checkout.ID.String(),
		AccountID:     account.ID.String(),
		Plan:          in.Plan,
		AmountCents:   quote.AmountCents,
		Currency:      quote.Currency,
		ReturnURL:     in.ReturnURL,
		CustomerEmail: in.Email,
	})
	if err != nil {
		return nil, err
	}

	if result.Status == "manual" {
		if err := s.applyPlan(ctx, account.ID, in.Plan, "manual", checkout.ID.String()); err != nil {
			return nil, err
		}
		_ = s.checkouts.MarkCompleted(ctx, checkout.ID)
		return &UpgradeResult{
			Status:    "completed",
			Message:   result.Message,
			Plan:      in.Plan,
			PlanLabel: PlanLabel(in.Plan),
		}, nil
	}

	return &UpgradeResult{
		Status:      result.Status,
		CheckoutURL: result.CheckoutURL,
		Message:     result.Message,
	}, nil
}

func (s *Service) applyPlan(ctx context.Context, accountID uuid.UUID, plan, provider, providerRef string) error {
	if err := s.subs.ApplyPlan(ctx, accountID, plan, provider, providerRef); err != nil {
		return err
	}
	return s.events.Log(ctx, accountID, "plan_changed", map[string]any{
		"plan":     plan,
		"provider": provider,
	})
}

func priceForPlan(plan, provider string) PriceQuote {
	for _, offer := range AllPlans() {
		if offer.ID != plan {
			continue
		}
		if provider == "stripe" && offer.PriceUSD != nil {
			return *offer.PriceUSD
		}
		if offer.PriceTRY != nil {
			return *offer.PriceTRY
		}
		break
	}
	return PriceQuote{AmountCents: 0, Currency: "TRY", Interval: "month"}
}
