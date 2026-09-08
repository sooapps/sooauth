-- +goose Up
ALTER TABLE tenants
    ADD COLUMN email_verify_delivery TEXT NOT NULL DEFAULT 'link';

ALTER TABLE tenants
    ADD CONSTRAINT tenants_email_verify_delivery_check
    CHECK (email_verify_delivery IN ('link', 'code', 'both'));

ALTER TABLE verification_tokens DROP CONSTRAINT IF EXISTS verification_tokens_type_check;
ALTER TABLE verification_tokens
    ADD CONSTRAINT verification_tokens_type_check
    CHECK (type IN ('email_verify', 'email_verify_code', 'password_reset'));

-- +goose Down
ALTER TABLE verification_tokens DROP CONSTRAINT IF EXISTS verification_tokens_type_check;
ALTER TABLE verification_tokens
    ADD CONSTRAINT verification_tokens_type_check
    CHECK (type IN ('email_verify', 'password_reset'));

ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_email_verify_delivery_check;
ALTER TABLE tenants DROP COLUMN IF EXISTS email_verify_delivery;
