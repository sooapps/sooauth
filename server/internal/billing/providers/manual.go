package providers

import (
	"context"

	"github.com/sooapps/sooauth/server/internal/billing"
)

type Manual struct {
	enabled bool
}

func NewManual(enabled bool) *Manual {
	return &Manual{enabled: enabled}
}

func (m *Manual) Name() string { return "manual" }

func (m *Manual) Available() bool { return m.enabled }

func (m *Manual) StartCheckout(_ context.Context, _ billing.CheckoutRequest) (*billing.CheckoutResult, error) {
	if !m.enabled {
		return nil, billing.ErrProviderUnavailable
	}
	return &billing.CheckoutResult{
		Status:  "manual",
		Message: "Manual billing enabled — plan will apply immediately.",
	}, nil
}
