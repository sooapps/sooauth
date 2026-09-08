-- +goose Up
CREATE TABLE platform_mfa (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    secret TEXT NOT NULL,
    enabled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE platform_mfa_backup_codes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL UNIQUE,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE platform_mfa_challenges (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX platform_mfa_challenges_expiry_idx ON platform_mfa_challenges (expires_at);

-- +goose Down
DROP TABLE IF EXISTS platform_mfa_challenges;
DROP TABLE IF EXISTS platform_mfa_backup_codes;
DROP TABLE IF EXISTS platform_mfa;
