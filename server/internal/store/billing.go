package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Subscription struct {
	ID                     uuid.UUID
	AccountID              uuid.UUID
	Plan                   string
	Status                 string
	Provider               string
	ProviderSubscriptionID *string
	ProviderCustomerID     *string
	CurrentPeriodStart     *time.Time
	CurrentPeriodEnd       *time.Time
	CancelAtPeriodEnd      bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type Subscriptions struct {
	db *pgxpool.Pool
}

func NewSubscriptions(db *pgxpool.Pool) *Subscriptions {
	return &Subscriptions{db: db}
}

func (s *Subscriptions) FindByAccount(ctx context.Context, accountID uuid.UUID) (*Subscription, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, account_id, plan, status, provider,
			provider_subscription_id, provider_customer_id,
			current_period_start, current_period_end, cancel_at_period_end,
			created_at, updated_at
		FROM subscriptions WHERE account_id = $1
	`, accountID)
	var sub Subscription
	var providerSubID, providerCustID *string
	if err := row.Scan(
		&sub.ID, &sub.AccountID, &sub.Plan, &sub.Status, &sub.Provider,
		&providerSubID, &providerCustID,
		&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd,
		&sub.CreatedAt, &sub.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	sub.ProviderSubscriptionID = providerSubID
	sub.ProviderCustomerID = providerCustID
	return &sub, nil
}

func (s *Subscriptions) EnsureForAccount(ctx context.Context, accountID uuid.UUID, plan string) (*Subscription, error) {
	if existing, err := s.FindByAccount(ctx, accountID); err != nil || existing != nil {
		return existing, err
	}
	id := uuid.New()
	now := time.Now().UTC()
	_, err := s.db.Exec(ctx, `
		INSERT INTO subscriptions (id, account_id, plan, status, provider, created_at, updated_at)
		VALUES ($1, $2, $3, 'active', 'manual', $4, $4)
	`, id, accountID, plan, now)
	if err != nil {
		return s.FindByAccount(ctx, accountID)
	}
	return &Subscription{
		ID:        id,
		AccountID: accountID,
		Plan:      plan,
		Status:    "active",
		Provider:  "manual",
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Subscriptions) ApplyPlan(
	ctx context.Context,
	accountID uuid.UUID,
	plan, provider, providerSubID string,
) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	periodEnd := now.AddDate(0, 1, 0)

	_, err = tx.Exec(ctx, `
		INSERT INTO subscriptions (
			id, account_id, plan, status, provider, provider_subscription_id,
			current_period_start, current_period_end, created_at, updated_at
		)
		VALUES ($1, $2, $3, 'active', $4, NULLIF($5, ''), $6, $7, $6, $6)
		ON CONFLICT (account_id) DO UPDATE SET
			plan = EXCLUDED.plan,
			status = 'active',
			provider = EXCLUDED.provider,
			provider_subscription_id = COALESCE(EXCLUDED.provider_subscription_id, subscriptions.provider_subscription_id),
			current_period_start = EXCLUDED.current_period_start,
			current_period_end = EXCLUDED.current_period_end,
			cancel_at_period_end = FALSE,
			updated_at = EXCLUDED.updated_at
	`, uuid.New(), accountID, plan, provider, providerSubID, now, periodEnd)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `UPDATE accounts SET plan = $2 WHERE id = $1`, accountID, plan)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type BillingCheckout struct {
	ID                 uuid.UUID
	AccountID          uuid.UUID
	Plan               string
	Provider           string
	AmountCents        int
	Currency           string
	Status             string
	ProviderCheckoutID *string
	ReturnURL          *string
	ExpiresAt          *time.Time
	CompletedAt        *time.Time
	CreatedAt          time.Time
}

type BillingCheckouts struct {
	db *pgxpool.Pool
}

func NewBillingCheckouts(db *pgxpool.Pool) *BillingCheckouts {
	return &BillingCheckouts{db: db}
}

func (b *BillingCheckouts) Create(
	ctx context.Context,
	accountID uuid.UUID,
	plan, provider string,
	amountCents int,
	currency, returnURL string,
) (*BillingCheckout, error) {
	id := uuid.New()
	expires := time.Now().UTC().Add(30 * time.Minute)
	now := time.Now().UTC()
	_, err := b.db.Exec(ctx, `
		INSERT INTO billing_checkouts (
			id, account_id, plan, provider, amount_cents, currency, status, return_url, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, 'pending', NULLIF($7, ''), $8, $9)
	`, id, accountID, plan, provider, amountCents, currency, returnURL, expires, now)
	if err != nil {
		return nil, err
	}
	return &BillingCheckout{
		ID:          id,
		AccountID:   accountID,
		Plan:        plan,
		Provider:    provider,
		AmountCents: amountCents,
		Currency:    currency,
		Status:      "pending",
		ExpiresAt:   &expires,
		CreatedAt:   now,
	}, nil
}

func (b *BillingCheckouts) MarkCompleted(ctx context.Context, checkoutID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := b.db.Exec(ctx, `
		UPDATE billing_checkouts SET status = 'completed', completed_at = $2 WHERE id = $1
	`, checkoutID, now)
	return err
}

type BillingEvents struct {
	db *pgxpool.Pool
}

func NewBillingEvents(db *pgxpool.Pool) *BillingEvents {
	return &BillingEvents{db: db}
}

func (b *BillingEvents) Log(ctx context.Context, accountID uuid.UUID, eventType string, payload map[string]any) error {
	raw, _ := json.Marshal(payload)
	_, err := b.db.Exec(ctx, `
		INSERT INTO billing_events (id, account_id, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, now())
	`, uuid.New(), accountID, eventType, raw)
	return err
}
