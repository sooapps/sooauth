-- +goose Up
ALTER TABLE tenants
    ADD COLUMN IF NOT EXISTS default_locale TEXT NOT NULL DEFAULT 'en';

ALTER TABLE tenants
    DROP CONSTRAINT IF EXISTS tenants_default_locale_check;

ALTER TABLE tenants
    ADD CONSTRAINT tenants_default_locale_check
    CHECK (default_locale IN ('en', 'tr'));

-- +goose Down
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_default_locale_check;
ALTER TABLE tenants DROP COLUMN IF EXISTS default_locale;
