-- +goose Up
ALTER TABLE verification_tokens DROP CONSTRAINT IF EXISTS verification_tokens_type_check;
ALTER TABLE verification_tokens
    ADD CONSTRAINT verification_tokens_type_check
    CHECK (type IN ('email_verify', 'email_verify_code', 'password_reset', 'password_reset_code'));

-- +goose Down
ALTER TABLE verification_tokens DROP CONSTRAINT IF EXISTS verification_tokens_type_check;
ALTER TABLE verification_tokens
    ADD CONSTRAINT verification_tokens_type_check
    CHECK (type IN ('email_verify', 'email_verify_code', 'password_reset'));
