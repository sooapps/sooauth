package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sooapps/sooauth/server/internal/store"
)

type Dispatcher struct {
	store  *store.Webhooks
	client *http.Client
}

func NewDispatcher(s *store.Webhooks) *Dispatcher {
	return &Dispatcher{store: s, client: &http.Client{Timeout: 10 * time.Second}}
}

type Event struct {
	ID        uuid.UUID      `json:"id"`
	Type      string         `json:"type"`
	CreatedAt time.Time      `json:"created_at"`
	TenantID  *uuid.UUID     `json:"tenant_id,omitempty"`
	User      *store.User    `json:"user,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
}

func (d *Dispatcher) Emit(ctx context.Context, tenantID *uuid.UUID, eventType string, user *store.User, data map[string]any) {
	if d == nil || d.store == nil || tenantID == nil || *tenantID == uuid.Nil {
		return
	}
	endpoints, err := d.store.ListByTenant(ctx, *tenantID)
	if err != nil {
		return
	}
	payload, err := json.Marshal(Event{ID: uuid.New(), Type: eventType, CreatedAt: time.Now().UTC(), TenantID: tenantID, User: user, Data: data})
	if err != nil {
		return
	}
	for _, endpoint := range endpoints {
		if !endpoint.Active || !subscribed(endpoint.Events, eventType) {
			continue
		}
		d.deliver(ctx, endpoint, eventType, payload)
	}
}

func subscribed(events []string, eventType string) bool {
	if len(events) == 0 {
		return true
	}
	for _, event := range events {
		if event == "*" || event == eventType {
			return true
		}
	}
	return false
}

func (d *Dispatcher) DeliverTest(ctx context.Context, endpoint *store.WebhookEndpoint, tenantID uuid.UUID) error {
	payload, err := json.Marshal(Event{ID: uuid.New(), Type: "webhook.test", CreatedAt: time.Now().UTC(), TenantID: &tenantID, Data: map[string]any{"message": "Webhook endpoint test"}})
	if err != nil {
		return err
	}
	return d.deliver(ctx, *endpoint, "webhook.test", payload)
}

func (d *Dispatcher) deliver(ctx context.Context, endpoint store.WebhookEndpoint, eventType string, payload []byte) error {
	deliveryID := uuid.New()
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if attempt > 1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt-1) * 250 * time.Millisecond):
			}
		}
		timestamp := time.Now().Unix()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.URL, bytes.NewReader(payload))
		if err != nil {
			lastErr = err
		} else {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "sooauth-webhooks/1")
			req.Header.Set("X-Sooauth-Event", eventType)
			req.Header.Set("X-Sooauth-Delivery", deliveryID.String())
			req.Header.Set("X-Sooauth-Signature", Signature(endpoint.Secret, timestamp, payload))
			resp, requestErr := d.client.Do(req)
			status, responseBody := 0, ""
			if resp != nil {
				status = resp.StatusCode
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
				_ = resp.Body.Close()
				responseBody = string(body)
			}
			if requestErr == nil && status >= 200 && status < 300 {
				now := time.Now().UTC()
				_ = d.store.RecordAttempt(ctx, endpoint.ID, deliveryID, eventType, payload, attempt, status, responseBody, "", &now)
				return nil
			}
			if requestErr != nil {
				lastErr = requestErr
			} else {
				lastErr = fmt.Errorf("endpoint returned %d", status)
			}
			_ = d.store.RecordAttempt(ctx, endpoint.ID, deliveryID, eventType, payload, attempt, status, responseBody, lastErr.Error(), nil)
		}
	}
	return lastErr
}

func ValidateURL(raw string) error {
	if !strings.HasPrefix(raw, "https://") && !strings.HasPrefix(raw, "http://") {
		return fmt.Errorf("webhook URL must use http or https")
	}
	return nil
}
