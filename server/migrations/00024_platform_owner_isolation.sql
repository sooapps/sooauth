-- +goose Up
-- Dashboard (platform) accounts must never be scoped to a project. An older
-- OIDC code path wrote the project id into the owner's row when they signed in
-- to their own app, which mixed the admin account with project end users.
UPDATE users SET tenant_id = NULL, updated_at = now()
WHERE platform_owner = TRUE AND tenant_id IS NOT NULL;

-- +goose Down
SELECT 1;
