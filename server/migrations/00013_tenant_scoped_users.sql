-- +goose Up
-- App users become tenant-scoped: same email may exist in different projects.

-- 1. Assign single-tenant app users to their project.
UPDATE users u
SET tenant_id = sub.tenant_id
FROM (
    SELECT tu.user_id, (MIN(tu.tenant_id::text))::uuid AS tenant_id
    FROM tenant_users tu
    JOIN users usr ON usr.id = tu.user_id
    WHERE NOT usr.platform_owner
    GROUP BY tu.user_id
    HAVING COUNT(*) = 1
) sub
WHERE u.id = sub.user_id
  AND NOT u.platform_owner
  AND u.tenant_id IS NULL;

-- 2. Drop old global uniqueness before splitting rows.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;

ALTER TABLE social_identities ADD COLUMN IF NOT EXISTS tenant_id UUID NULL REFERENCES tenants (id) ON DELETE CASCADE;

ALTER TABLE social_identities DROP CONSTRAINT IF EXISTS social_identities_provider_sub_unique;

-- 3. Tag existing social rows with the user's primary tenant when known.
UPDATE social_identities si
SET tenant_id = u.tenant_id
FROM users u
WHERE si.user_id = u.id
  AND si.tenant_id IS NULL
  AND u.tenant_id IS NOT NULL;

-- 4. Duplicate shared app users into one row per additional project membership.
-- +goose StatementBegin
DO $$
DECLARE
    rec RECORD;
    new_user_id UUID;
BEGIN
    FOR rec IN
        SELECT
            tu.tenant_id,
            tu.user_id,
            u.email,
            u.email_verified_at,
            u.created_at
        FROM tenant_users tu
        JOIN users u ON u.id = tu.user_id
        WHERE NOT u.platform_owner
          AND (u.tenant_id IS NULL OR tu.tenant_id <> u.tenant_id)
        ORDER BY tu.user_id, tu.tenant_id
    LOOP
        new_user_id := gen_random_uuid();

        INSERT INTO users (id, tenant_id, email, email_verified_at, platform_owner, created_at, updated_at)
        VALUES (new_user_id, rec.tenant_id, rec.email, rec.email_verified_at, FALSE, rec.created_at, now());

        INSERT INTO credentials (id, user_id, type, password_hash, created_at)
        SELECT gen_random_uuid(), new_user_id, c.type, c.password_hash, now()
        FROM credentials c
        WHERE c.user_id = rec.user_id AND c.type = 'password';

        INSERT INTO social_identities (id, user_id, tenant_id, provider, provider_sub, email, created_at)
        SELECT gen_random_uuid(), new_user_id, rec.tenant_id, si.provider, si.provider_sub, si.email, si.created_at
        FROM social_identities si
        WHERE si.user_id = rec.user_id;

        UPDATE tenant_users
        SET user_id = new_user_id
        WHERE tenant_id = rec.tenant_id AND user_id = rec.user_id;
    END LOOP;
END $$;
-- +goose StatementEnd

-- 5. New scoped uniqueness.
CREATE UNIQUE INDEX IF NOT EXISTS users_platform_email_unique
    ON users (lower(email))
    WHERE platform_owner = TRUE;

CREATE UNIQUE INDEX IF NOT EXISTS users_tenant_email_unique
    ON users (tenant_id, lower(email))
    WHERE platform_owner = FALSE AND tenant_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS social_identities_tenant_provider_sub_unique
    ON social_identities (tenant_id, provider, provider_sub)
    WHERE tenant_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS social_identities_platform_provider_sub_unique
    ON social_identities (provider, provider_sub)
    WHERE tenant_id IS NULL;

CREATE INDEX IF NOT EXISTS social_identities_tenant_id_idx ON social_identities (tenant_id);

-- +goose Down
DROP INDEX IF EXISTS social_identities_tenant_id_idx;
DROP INDEX IF EXISTS social_identities_platform_provider_sub_unique;
DROP INDEX IF EXISTS social_identities_tenant_provider_sub_unique;
ALTER TABLE social_identities DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE social_identities
    ADD CONSTRAINT social_identities_provider_sub_unique UNIQUE (provider, provider_sub);

DROP INDEX IF EXISTS users_tenant_email_unique;
DROP INDEX IF EXISTS users_platform_email_unique;
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
