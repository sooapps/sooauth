# SQL migrations (goose)

M2 adds goose migrations for users, credentials, sessions, refresh_tokens, providers, verification_tokens, audit_log.

Reserved column on user-facing tables: `tenant_id` (nullable, unused in v1).

Run (M2+):

```bash
goose -dir migrations postgres "$DATABASE_URL" up
```
