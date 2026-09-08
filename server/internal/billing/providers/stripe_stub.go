package providers

import (
	"context"

	"github.com/sooapps/sooauth/server/internal/billing"
)

// Stripe will handle international card payments (USD/EUR). Wire STRIPE_* env when ready.
type Stripe struct{}

func NewStripe() *Stripe { return &Stripe{} }

func (p *Stripe) Name() string { return "stripe" }

func (p *Stripe) Available() bool { return false }

func (p *Stripe) StartCheckout(_ context.Context, _ billing.CheckoutRequest) (*billing.CheckoutResult, error) {
	return &billing.CheckoutResult{
		Status:  "coming_soon",
		Message: "Stripe checkout is not enabled yet.",
	}, nil
}
