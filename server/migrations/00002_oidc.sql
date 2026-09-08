-- +goose Up
CREATE TABLE oauth_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id TEXT NOT NULL,
    client_secret_hash TEXT NULL,
    name TEXT NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    allowed_scopes TEXT[] NOT NULL DEFAULT ARRAY['openid', 'email', 'profile'],
    public BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT oauth_clients_client_id_unique UNIQUE (client_id)
);

CREATE TABLE authorization_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code_hash TEXT NOT NULL,
    client_id TEXT NOT NULL REFERENCES oauth_clients (client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    redirect_uri TEXT NOT NULL,
    scope TEXT NOT NULL,
    nonce TEXT NULL,
    code_challenge TEXT NOT NULL,
    code_challenge_method TEXT NOT NULL DEFAULT 'S256',
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT authorization_codes_code_hash_unique UNIQUE (code_hash)
);

CREATE INDEX authorization_codes_client_id_idx ON authorization_codes (client_id);

INSERT INTO oauth_clients (client_id, name, redirect_uris, public)
VALUES (
    'dev',
    'Development client',
    ARRAY[
        'http://localhost:3000/callback',
        'http://127.0.0.1:3000/callback',
        'http://localhost:5556/callback',
        'http://127.0.0.1:5556/callback'
    ],
    TRUE
);

-- +goose Down
DROP TABLE IF EXISTS authorization_codes;
DROP TABLE IF EXISTS oauth_clients;
