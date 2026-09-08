package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookEndpoint struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	URL       string
	Secret    string
	Events    []string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WebhookDeliveryAttempt struct {
	ID          uuid.UUID
	EndpointID  uuid.UUID
	DeliveryID  uuid.UUID
	EventType   string
	Payload     string
	Attempt     int
	StatusCode  *int
	Response    *string
	Error       *string
	DeliveredAt *time.Time
	CreatedAt   time.Time
}

func (s *Webhooks) ListAttemptsByEndpoint(ctx context.Context, endpointID, tenantID uuid.UUID) ([]WebhookDeliveryAttempt, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.endpoint_id, a.delivery_id, a.event_type, a.payload::text,
		       a.attempt, a.status_code, a.response, a.error, a.delivered_at, a.created_at
		FROM webhook_delivery_attempts a
		JOIN webhook_endpoints e ON e.id = a.endpoint_id
		WHERE a.endpoint_id = $1 AND e.tenant_id = $2
		ORDER BY a.created_at DESC LIMIT 25
	`, endpointID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WebhookDeliveryAttempt
	for rows.Next() {
		var attempt WebhookDeliveryAttempt
		if err := rows.Scan(&attempt.ID, &attempt.EndpointID, &attempt.DeliveryID, &attempt.EventType, &attempt.Payload, &attempt.Attempt, &attempt.StatusCode, &attempt.Response, &attempt.Error, &attempt.DeliveredAt, &attempt.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, attempt)
	}
	return out, rows.Err()
}

type Webhooks struct{ db *pgxpool.Pool }

func NewWebhooks(db *pgxpool.Pool) *Webhooks { return &Webhooks{db: db} }

func (s *Webhooks) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]WebhookEndpoint, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, url, events, active, created_at, updated_at
		FROM webhook_endpoints WHERE tenant_id = $1 ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WebhookEndpoint
	for rows.Next() {
		var e WebhookEndpoint
		if err := rows.Scan(&e.ID, &e.TenantID, &e.URL, &e.Events, &e.Active, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Webhooks) FindByIDForTenant(ctx context.Context, tenantID, id uuid.UUID) (*WebhookEndpoint, error) {
	var e WebhookEndpoint
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, url, secret, events, active, created_at, updated_at
		FROM webhook_endpoints WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&e.ID, &e.TenantID, &e.URL, &e.Secret, &e.Events, &e.Active, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Webhooks) Create(ctx context.Context, tenantID uuid.UUID, url, secret string, events []string) (*WebhookEndpoint, error) {
	id := uuid.New()
	_, err := s.db.Exec(ctx, `
		INSERT INTO webhook_endpoints (id, tenant_id, url, secret, events)
		VALUES ($1, $2, $3, $4, $5)
	`, id, tenantID, url, secret, events)
	if err != nil {
		return nil, err
	}
	return s.FindByIDForTenant(ctx, tenantID, id)
}

func (s *Webhooks) Update(ctx context.Context, tenantID, id uuid.UUID, url string, events []string, active bool) error {
	_, err := s.db.Exec(ctx, `
		UPDATE webhook_endpoints SET url = $3, events = $4, active = $5, updated_at = now()
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID, url, events, active)
	return err
}

func (s *Webhooks) Delete(ctx context.Context, tenantID, id uuid.UUID) (bool, error) {
	tag, err := s.db.Exec(ctx, `DELETE FROM webhook_endpoints WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return tag.RowsAffected() > 0, err
}

func (s *Webhooks) RecordAttempt(ctx context.Context, endpointID, deliveryID uuid.UUID, eventType string, payload []byte, attempt, statusCode int, response, deliveryError string, deliveredAt *time.Time) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO webhook_delivery_attempts
		(id, endpoint_id, delivery_id, event_type, payload, attempt, status_code, response, error, delivered_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, NULLIF($7, 0), NULLIF($8, ''), NULLIF($9, ''), $10)
	`, uuid.New(), endpointID, deliveryID, eventType, string(payload), attempt, statusCode, response, deliveryError, deliveredAt)
	return err
}
