-- +goose Up
CREATE TABLE IF NOT EXISTS tenant_email_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    provider TEXT NOT NULL DEFAULT 'smtp',
    from_name TEXT NOT NULL DEFAULT '',
    from_email TEXT NOT NULL DEFAULT '',
    reply_to TEXT NOT NULL DEFAULT '',

    -- SMTP settings
    smtp_host TEXT NOT NULL DEFAULT '',
    smtp_port INT NOT NULL DEFAULT 587,
    smtp_user TEXT NOT NULL DEFAULT '',
    smtp_password_encrypted TEXT NOT NULL DEFAULT '',
    smtp_tls_mode TEXT NOT NULL DEFAULT 'starttls',

    -- API key providers (Resend, Postmark)
    api_key_encrypted TEXT NOT NULL DEFAULT '',

    -- AWS SES settings
    aws_access_key_id TEXT NOT NULL DEFAULT '',
    aws_secret_key_encrypted TEXT NOT NULL DEFAULT '',
    aws_region TEXT NOT NULL DEFAULT 'us-east-1',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tenant_email_settings_tenant_id_unique UNIQUE (tenant_id)
);

CREATE INDEX IF NOT EXISTS tenant_email_settings_tenant_id_idx ON tenant_email_settings (tenant_id);

-- +goose Down
DROP INDEX IF EXISTS tenant_email_settings_tenant_id_idx;
DROP TABLE IF EXISTS tenant_email_settings;
