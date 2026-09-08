-- +goose Up
ALTER TABLE oauth_providers
    ADD COLUMN use_platform BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE oauth_providers DROP COLUMN IF EXISTS use_platform;
