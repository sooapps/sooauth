-- +goose Up
-- Dashboard setup progress per project (wizard state and manual checklist marks).
ALTER TABLE tenants
    ADD COLUMN IF NOT EXISTS onboarding JSONB NOT NULL DEFAULT '{}'::jsonb;

-- Existing projects already know their way around: start them with the
-- wizard dismissed (the checklist is still shown).
UPDATE tenants SET onboarding = '{"dismissed": true}'::jsonb WHERE onboarding = '{}'::jsonb;

-- +goose Down
ALTER TABLE tenants DROP COLUMN IF EXISTS onboarding;
