-- +goose Up
-- 1. Add auth_config and registration_schema to tenants
ALTER TABLE tenants
    ADD COLUMN IF NOT EXISTS auth_config JSONB NOT NULL DEFAULT '{"allowed_identifiers":["email"],"primary_auth_mode":"password"}'::jsonb,
    ADD COLUMN IF NOT EXISTS registration_schema JSONB NOT NULL DEFAULT '[]'::jsonb;

-- 2. Allow users without email (e.g. username-only or phone-only)
ALTER TABLE users
    ALTER COLUMN email DROP NOT NULL;

-- 3. Add username, phone, phone_verified_at, and metadata to users
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS username TEXT NULL,
    ADD COLUMN IF NOT EXISTS phone TEXT NULL,
    ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;

-- 4. Unique indexes for username and phone within tenant
CREATE UNIQUE INDEX IF NOT EXISTS users_tenant_username_unique
    ON users (tenant_id, lower(username))
    WHERE platform_owner = FALSE AND tenant_id IS NOT NULL AND username IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS users_tenant_phone_unique
    ON users (tenant_id, phone)
    WHERE platform_owner = FALSE AND tenant_id IS NOT NULL AND phone IS NOT NULL;

-- 5. Check constraint: at least one identifier must be non-null
ALTER TABLE users
    ADD CONSTRAINT users_has_identifier_check
    CHECK (email IS NOT NULL OR username IS NOT NULL OR phone IS NOT NULL);

-- 6. Update tenant_users signup_method check constraint
ALTER TABLE tenant_users
    DROP CONSTRAINT IF EXISTS tenant_users_signup_method_check;

ALTER TABLE tenant_users
    ADD CONSTRAINT tenant_users_signup_method_check
    CHECK (signup_method IS NULL OR signup_method IN ('email', 'username', 'phone', 'otp', 'google', 'github', 'facebook', 'x'));

-- +goose Down
ALTER TABLE tenant_users
    DROP CONSTRAINT IF EXISTS tenant_users_signup_method_check;

ALTER TABLE tenant_users
    ADD CONSTRAINT tenant_users_signup_method_check
    CHECK (signup_method IS NULL OR signup_method IN ('email', 'google', 'github', 'facebook', 'x'));

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_has_identifier_check;

DROP INDEX IF EXISTS users_tenant_phone_unique;
DROP INDEX IF EXISTS users_tenant_username_unique;

ALTER TABLE users
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS phone_verified_at,
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS username;

UPDATE users SET email = id || '@placeholder.sooauth.internal' WHERE email IS NULL;
ALTER TABLE users
    ALTER COLUMN email SET NOT NULL;

ALTER TABLE tenants
    DROP COLUMN IF EXISTS registration_schema,
    DROP COLUMN IF EXISTS auth_config;
