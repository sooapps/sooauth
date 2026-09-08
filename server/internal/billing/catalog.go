package billing

type PriceQuote struct {
	AmountCents int    `json:"amount_cents"`
	Currency    string `json:"currency"`
	Interval    string `json:"interval"`
}

type PlanOffer struct {
	ID          string       `json:"id"`
	Label       string       `json:"label"`
	MaxProjects int          `json:"max_projects"`
	Features    []string     `json:"features"`
	PriceTRY    *PriceQuote  `json:"price_try,omitempty"`
	PriceUSD    *PriceQuote  `json:"price_usd,omitempty"`
	PaymentLive bool         `json:"payment_live"`
}

func AllPlans() []PlanOffer {
	return []PlanOffer{
		{
			ID:          PlanFree,
			Label:       PlanLabel(PlanFree),
			MaxProjects: MaxProjects(PlanFree),
			Features: []string{
				"1 project",
				"Email + social login",
				"Embed widget",
				"Powered by sooauth branding",
			},
			PaymentLive: true,
		},
		{
			ID:          PlanPro,
			Label:       PlanLabel(PlanPro),
			MaxProjects: MaxProjects(PlanPro),
			Features: []string{
				"Up to 3 projects",
				"Custom login branding",
				"Hide powered-by badge",
				"Custom OAuth credentials",
			},
			PriceTRY: &PriceQuote{
				AmountCents: 49900,
				Currency:    "TRY",
				Interval:    "month",
			},
			PriceUSD: &PriceQuote{
				AmountCents: 1900,
				Currency:    "USD",
				Interval:    "month",
			},
		},
		{
			ID:          PlanBusiness,
			Label:       PlanLabel(PlanBusiness),
			MaxProjects: MaxProjects(PlanBusiness),
			Features: []string{
				"Up to 10 projects",
				"Everything in Pro",
				"Priority support (coming soon)",
			},
			PriceTRY: &PriceQuote{
				AmountCents: 149900,
				Currency:    "TRY",
				Interval:    "month",
			},
			PriceUSD: &PriceQuote{
				AmountCents: 4900,
				Currency:    "USD",
				Interval:    "month",
			},
		},
	}
}

func ValidPlan(plan string) bool {
	switch plan {
	case PlanFree, PlanPro, PlanBusiness:
		return true
	default:
		return false
	}
}

func PlanRank(plan string) int {
	switch plan {
	case PlanBusiness:
		return 3
	case PlanPro:
		return 2
	case PlanFree:
		return 1
	default:
		return 0
	}
}
