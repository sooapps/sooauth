-- +goose Up
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS is_beta_user BOOLEAN NOT NULL DEFAULT FALSE;

-- Backfill existing early accounts to be recognized as beta adopters with Pro tier access
UPDATE accounts SET is_beta_user = TRUE, plan = 'pro' WHERE plan = 'free';
UPDATE accounts SET is_beta_user = TRUE WHERE plan IN ('pro', 'business');

-- +goose Down
ALTER TABLE accounts DROP COLUMN IF EXISTS is_beta_user;
