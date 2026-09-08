-- +goose Up
CREATE TABLE social_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider TEXT NOT NULL CHECK (provider IN ('google', 'github')),
    provider_sub TEXT NOT NULL,
    email TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT social_identities_provider_sub_unique UNIQUE (provider, provider_sub)
);

CREATE INDEX social_identities_user_id_idx ON social_identities (user_id);

CREATE TABLE theme_config (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    brand_name TEXT NOT NULL DEFAULT 'sooauth',
    logo_url TEXT NULL,
    accent_color TEXT NOT NULL DEFAULT '#7B72E9',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO theme_config (id) VALUES (1);

-- +goose Down
DROP TABLE IF EXISTS theme_config;
DROP TABLE IF EXISTS social_identities;
