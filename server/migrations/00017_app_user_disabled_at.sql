-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS disabled_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS users_tenant_disabled_at_idx
    ON users (tenant_id, disabled_at)
    WHERE platform_owner = FALSE;

-- +goose Down
DROP INDEX IF EXISTS users_tenant_disabled_at_idx;
ALTER TABLE users DROP COLUMN IF EXISTS disabled_at;
