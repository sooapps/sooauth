-- +goose Up
INSERT INTO oauth_clients (client_id, name, redirect_uris, public)
VALUES (
    'soobserver',
    'SoObserver Dashboard',
    ARRAY[
        'http://localhost:3001/api/auth/callback/oidc',
        'http://127.0.0.1:3001/api/auth/callback/oidc'
    ],
    TRUE
)
ON CONFLICT (client_id) DO UPDATE
SET redirect_uris = EXCLUDED.redirect_uris,
    name = EXCLUDED.name;

-- +goose Down
DELETE FROM oauth_clients WHERE client_id = 'soobserver';
