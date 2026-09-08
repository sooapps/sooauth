package config

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Env                  string
	Port                 string
	DatabaseURL          string
	RedisURL             string
	AppURL               string
	APIURL               string
	CookieDomain         string
	DashboardOrigin      string
	SMTPURL              string
	SMTPHost             string
	SMTPPort             string
	SMTPUser             string
	SMTPPassword         string
	SMTPFrom             string
	AutoMigrate          bool
	CookieSecure         bool
	GoogleClientID       string
	GoogleClientSecret   string
	GithubClientID       string
	GithubClientSecret   string
	FacebookClientID     string
	FacebookClientSecret string
	XClientID            string
	XClientSecret        string
	AdminEmails          []string
	WebAuthnRPID         string
	WebAuthnOrigin       string
	BrandName            string
	BrandLogoURL         string
	BrandAccent          string
	BillingManualUpgrade bool
	MFAEncryptionKey     []byte
}

func Load() (Config, error) {
	loadEnvFiles()

	env := envOr("SOOAUTH_ENV", "development")
	cfg := Config{
		Env:                  env,
		Port:                 envOr("PORT", "8080"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		RedisURL:             os.Getenv("REDIS_URL"),
		AppURL:               envOr("APP_URL", "http://localhost:8080"),
		APIURL:               envOr("API_URL", "http://localhost:8080"),
		CookieDomain:         os.Getenv("COOKIE_DOMAIN"),
		DashboardOrigin:      envOr("DASHBOARD_ORIGIN", envOr("APP_URL", "http://localhost:8080")),
		SMTPURL:              os.Getenv("SMTP_URL"),
		SMTPHost:             os.Getenv("SMTP_HOST"),
		SMTPPort:             envOr("SMTP_PORT", "465"),
		SMTPUser:             os.Getenv("SMTP_USER"),
		SMTPPassword:         os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:             envOr("SMTP_FROM", "sooauth@localhost"),
		AutoMigrate:          env == "development" || os.Getenv("AUTO_MIGRATE") == "true",
		CookieSecure:         env == "production",
		GoogleClientID:       os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		GithubClientID:       os.Getenv("GITHUB_CLIENT_ID"),
		GithubClientSecret:   os.Getenv("GITHUB_CLIENT_SECRET"),
		FacebookClientID:     os.Getenv("FACEBOOK_CLIENT_ID"),
		FacebookClientSecret: os.Getenv("FACEBOOK_CLIENT_SECRET"),
		XClientID:            os.Getenv("X_CLIENT_ID"),
		XClientSecret:        os.Getenv("X_CLIENT_SECRET"),
		AdminEmails:          splitCSV(os.Getenv("ADMIN_EMAILS")),
		WebAuthnRPID:         envOr("WEBAUTHN_RP_ID", "localhost"),
		WebAuthnOrigin:       envOr("WEBAUTHN_ORIGIN", envOr("APP_URL", "http://localhost:8080")),
		BrandName:            envOr("BRAND_NAME", "sooauth"),
		BrandLogoURL:         os.Getenv("BRAND_LOGO_URL"),
		BrandAccent:          envOr("BRAND_ACCENT", "#E8FF3F"),
		BillingManualUpgrade: os.Getenv("BILLING_MANUAL_UPGRADE") == "true",
	}
	if raw := os.Getenv("MFA_ENCRYPTION_KEY"); raw != "" {
		key, err := decodeMFAEncryptionKey(raw)
		if err != nil {
			return cfg, err
		}
		cfg.MFAEncryptionKey = key
	}

	if cfg.Env == "production" {
		if cfg.DatabaseURL == "" {
			return cfg, fmt.Errorf("DATABASE_URL is required in production")
		}
		if cfg.RedisURL == "" {
			return cfg, fmt.Errorf("REDIS_URL is required in production")
		}
		if len(cfg.MFAEncryptionKey) == 0 {
			return cfg, fmt.Errorf("MFA_ENCRYPTION_KEY is required in production")
		}
	}

	return cfg, nil
}

func decodeMFAEncryptionKey(raw string) ([]byte, error) {
	raw = strings.TrimSpace(strings.Trim(raw, `"'`))
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil || len(key) != 32 {
		key, err = base64.RawStdEncoding.DecodeString(raw)
	}
	if err != nil || len(key) != 32 {
		key, err = hex.DecodeString(raw)
	}
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("MFA_ENCRYPTION_KEY must be a base64-encoded or 64-character hex-encoded 32-byte key")
	}
	return key, nil
}

func (c Config) SMTPConfigured() bool {
	if c.SMTPHost != "" && c.SMTPUser != "" {
		return true
	}
	return c.SMTPURL != ""
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, strings.ToLower(part))
		}
	}
	return out
}
