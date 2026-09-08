-- +goose Up
ALTER TABLE tenants
    ADD COLUMN password_min_length INT NOT NULL DEFAULT 8,
    ADD COLUMN password_require_uppercase BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN password_require_number BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN password_require_special BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE tenant_users
    ADD COLUMN signup_method TEXT NULL;

ALTER TABLE tenant_users
    ADD CONSTRAINT tenant_users_signup_method_check
    CHECK (signup_method IS NULL OR signup_method IN ('email', 'google', 'github', 'facebook', 'x'));

-- +goose Down
ALTER TABLE tenant_users DROP CONSTRAINT IF EXISTS tenant_users_signup_method_check;
ALTER TABLE tenant_users DROP COLUMN IF EXISTS signup_method;
ALTER TABLE tenants
    DROP COLUMN IF EXISTS password_require_special,
    DROP COLUMN IF EXISTS password_require_number,
    DROP COLUMN IF EXISTS password_require_uppercase,
    DROP COLUMN IF EXISTS password_min_length;
