-- +goose Up
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    plan TEXT NOT NULL DEFAULT 'free' CHECK (plan IN ('free', 'pro', 'business')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT accounts_owner_user_id_unique UNIQUE (owner_user_id)
);

INSERT INTO accounts (owner_user_id, plan)
SELECT owner_user_id, 'free' FROM tenants
ON CONFLICT (owner_user_id) DO NOTHING;

ALTER TABLE tenants ADD COLUMN account_id UUID NULL REFERENCES accounts (id) ON DELETE CASCADE;

UPDATE tenants t
SET account_id = a.id
FROM accounts a
WHERE a.owner_user_id = t.owner_user_id AND t.account_id IS NULL;

ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_owner_user_id_unique;

CREATE INDEX tenants_account_id_idx ON tenants (account_id);

CREATE TABLE tenant_themes (
    tenant_id UUID PRIMARY KEY REFERENCES tenants (id) ON DELETE CASCADE,
    brand_name TEXT NOT NULL DEFAULT 'sooauth',
    logo_url TEXT NULL,
    accent_color TEXT NOT NULL DEFAULT '#7B72E9',
    hide_branding BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE oauth_providers ADD COLUMN tenant_id UUID NULL REFERENCES tenants (id) ON DELETE CASCADE;
ALTER TABLE oauth_providers DROP CONSTRAINT IF EXISTS oauth_providers_provider_unique;
ALTER TABLE oauth_providers DROP CONSTRAINT IF EXISTS oauth_providers_provider_check;
ALTER TABLE oauth_providers ADD CONSTRAINT oauth_providers_provider_check
    CHECK (provider IN ('google', 'github', 'facebook', 'x'));
CREATE UNIQUE INDEX oauth_providers_tenant_provider_unique
    ON oauth_providers (tenant_id, provider) WHERE tenant_id IS NOT NULL;
CREATE UNIQUE INDEX oauth_providers_global_provider_unique
    ON oauth_providers (provider) WHERE tenant_id IS NULL;

-- +goose Down
DROP INDEX IF EXISTS oauth_providers_global_provider_unique;
DROP INDEX IF EXISTS oauth_providers_tenant_provider_unique;
ALTER TABLE oauth_providers DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE oauth_providers DROP CONSTRAINT IF EXISTS oauth_providers_provider_check;
ALTER TABLE oauth_providers ADD CONSTRAINT oauth_providers_provider_check
    CHECK (provider IN ('google', 'github'));
ALTER TABLE oauth_providers ADD CONSTRAINT oauth_providers_provider_unique UNIQUE (provider);

DROP TABLE IF EXISTS tenant_themes;

DROP INDEX IF EXISTS tenants_account_id_idx;
ALTER TABLE tenants DROP COLUMN IF EXISTS account_id;
ALTER TABLE tenants ADD CONSTRAINT tenants_owner_user_id_unique UNIQUE (owner_user_id);

DROP TABLE IF EXISTS accounts;
