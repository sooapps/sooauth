package billing

import (
	"context"
	"errors"
)

var ErrProviderUnavailable = errors.New("billing provider unavailable")

type CheckoutRequest struct {
	CheckoutID string
	AccountID  string
	Plan       string
	AmountCents int
	Currency   string
	ReturnURL  string
	CustomerEmail string
}

type CheckoutResult struct {
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url,omitempty"`
	Message     string `json:"message,omitempty"`
}

type Provider interface {
	Name() string
	Available() bool
	StartCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResult, error)
}
