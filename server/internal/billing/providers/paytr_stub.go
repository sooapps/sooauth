package providers

import (
	"context"

	"github.com/sooapps/sooauth/server/internal/billing"
)

// PayTR will handle domestic card payments (TRY). Wire API keys via PAYTR_* env when ready.
type PayTR struct{}

func NewPayTR() *PayTR { return &PayTR{} }

func (p *PayTR) Name() string { return "paytr" }

func (p *PayTR) Available() bool { return false }

func (p *PayTR) StartCheckout(_ context.Context, _ billing.CheckoutRequest) (*billing.CheckoutResult, error) {
	return &billing.CheckoutResult{
		Status:  "coming_soon",
		Message: "PayTR checkout is not enabled yet.",
	}, nil
}
