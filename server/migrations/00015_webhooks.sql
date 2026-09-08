-- +goose Up
CREATE TABLE webhook_endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    secret TEXT NOT NULL,
    events TEXT[] NOT NULL DEFAULT '{}',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX webhook_endpoints_tenant_id_idx ON webhook_endpoints (tenant_id);

CREATE TABLE webhook_delivery_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint_id UUID NOT NULL REFERENCES webhook_endpoints (id) ON DELETE CASCADE,
    delivery_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    attempt INTEGER NOT NULL,
    status_code INTEGER NULL,
    response TEXT NULL,
    error TEXT NULL,
    delivered_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX webhook_delivery_attempts_endpoint_idx ON webhook_delivery_attempts (endpoint_id, created_at DESC);
CREATE INDEX webhook_delivery_attempts_delivery_idx ON webhook_delivery_attempts (delivery_id);

-- +goose Down
DROP TABLE IF EXISTS webhook_delivery_attempts;
DROP INDEX IF EXISTS webhook_endpoints_tenant_id_idx;
DROP TABLE IF EXISTS webhook_endpoints;
