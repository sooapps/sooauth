-- +goose Up
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    email_verify_required BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tenants_owner_user_id_unique UNIQUE (owner_user_id)
);

ALTER TABLE oauth_clients
    ADD COLUMN tenant_id UUID NULL REFERENCES tenants (id) ON DELETE CASCADE;

CREATE INDEX oauth_clients_tenant_id_idx ON oauth_clients (tenant_id);
CREATE INDEX users_tenant_id_idx ON users (tenant_id);

-- +goose Down
DROP INDEX IF EXISTS users_tenant_id_idx;
DROP INDEX IF EXISTS oauth_clients_tenant_id_idx;
ALTER TABLE oauth_clients DROP COLUMN IF EXISTS tenant_id;
DROP TABLE IF EXISTS tenants;
