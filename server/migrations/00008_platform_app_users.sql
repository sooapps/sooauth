-- +goose Up
ALTER TABLE users ADD COLUMN platform_owner BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE users u
SET platform_owner = TRUE
FROM accounts a
WHERE a.owner_user_id = u.id;

CREATE TABLE tenant_users (
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, user_id)
);

CREATE INDEX tenant_users_user_id_idx ON tenant_users (user_id);

-- +goose Down
DROP TABLE IF EXISTS tenant_users;
ALTER TABLE users DROP COLUMN IF EXISTS platform_owner;
