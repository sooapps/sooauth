-- +goose Up
ALTER TABLE tenants ADD COLUMN social_callback_origin TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE tenants DROP COLUMN IF EXISTS social_callback_origin;
