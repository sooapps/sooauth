-- +goose Up
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL UNIQUE REFERENCES accounts (id) ON DELETE CASCADE,
    plan TEXT NOT NULL DEFAULT 'free' CHECK (plan IN ('free', 'pro', 'business')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (
        status IN ('active', 'trialing', 'past_due', 'canceled', 'pending')
    ),
    provider TEXT NOT NULL DEFAULT 'manual' CHECK (
        provider IN ('manual', 'paytr', 'stripe', 'lemonsqueezy')
    ),
    provider_subscription_id TEXT,
    provider_customer_id TEXT,
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX subscriptions_status_idx ON subscriptions (status);

CREATE TABLE billing_checkouts (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    plan TEXT NOT NULL CHECK (plan IN ('free', 'pro', 'business')),
    provider TEXT NOT NULL CHECK (
        provider IN ('manual', 'paytr', 'stripe', 'lemonsqueezy')
    ),
    amount_cents INT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'TRY',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (
        status IN ('pending', 'completed', 'failed', 'expired')
    ),
    provider_checkout_id TEXT,
    provider_payload JSONB,
    return_url TEXT,
    expires_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX billing_checkouts_account_id_idx ON billing_checkouts (account_id);
CREATE INDEX billing_checkouts_status_idx ON billing_checkouts (status);

CREATE TABLE billing_events (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX billing_events_account_id_idx ON billing_events (account_id);

INSERT INTO subscriptions (id, account_id, plan, status, provider, created_at, updated_at)
SELECT gen_random_uuid(), a.id, a.plan, 'active', 'manual', a.created_at, a.created_at
FROM accounts a
ON CONFLICT (account_id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS billing_events;
DROP TABLE IF EXISTS billing_checkouts;
DROP TABLE IF EXISTS subscriptions;
