package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var AllSocialProviders = []string{"google", "github", "facebook", "x"}

type OAuthProviderConfig struct {
	Provider              string
	ClientID              string
	ClientSecretEncrypted string
	Enabled               bool
	UsePlatform           bool
}

type OAuthProviders struct {
	db *pgxpool.Pool
}

func NewOAuthProviders(db *pgxpool.Pool) *OAuthProviders {
	return &OAuthProviders{db: db}
}

func (s *OAuthProviders) scanConfig(row pgx.Row) (*OAuthProviderConfig, error) {
	var cfg OAuthProviderConfig
	if err := row.Scan(&cfg.Provider, &cfg.ClientID, &cfg.ClientSecretEncrypted, &cfg.Enabled, &cfg.UsePlatform); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (s *OAuthProviders) FindByProvider(ctx context.Context, provider string) (*OAuthProviderConfig, error) {
	return s.findGlobal(ctx, provider)
}

func (s *OAuthProviders) FindByTenantProvider(ctx context.Context, tenantID uuid.UUID, provider string) (*OAuthProviderConfig, error) {
	return s.scanConfig(s.db.QueryRow(ctx, `
		SELECT provider, client_id, client_secret_encrypted, enabled, use_platform
		FROM oauth_providers
		WHERE tenant_id = $1 AND provider = $2
	`, tenantID, provider))
}

func (s *OAuthProviders) findGlobal(ctx context.Context, provider string) (*OAuthProviderConfig, error) {
	return s.scanConfig(s.db.QueryRow(ctx, `
		SELECT provider, client_id, client_secret_encrypted, enabled, use_platform
		FROM oauth_providers
		WHERE provider = $1 AND tenant_id IS NULL
	`, provider))
}

func (s *OAuthProviders) FindEnabled(ctx context.Context, provider string) (*OAuthProviderConfig, error) {
	return s.scanConfig(s.db.QueryRow(ctx, `
		SELECT provider, client_id, client_secret_encrypted, enabled, use_platform
		FROM oauth_providers
		WHERE provider = $1 AND tenant_id IS NULL AND enabled = TRUE
	`, provider))
}

func (s *OAuthProviders) List(ctx context.Context) ([]OAuthProviderConfig, error) {
	return s.ListByTenant(ctx, uuid.Nil)
}

func (s *OAuthProviders) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]OAuthProviderConfig, error) {
	byName := make(map[string]OAuthProviderConfig, len(AllSocialProviders))
	if tenantID != uuid.Nil {
		rows, err := s.db.Query(ctx, `
			SELECT provider, client_id, client_secret_encrypted, enabled, use_platform
			FROM oauth_providers WHERE tenant_id = $1
		`, tenantID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var cfg OAuthProviderConfig
			if err := rows.Scan(&cfg.Provider, &cfg.ClientID, &cfg.ClientSecretEncrypted, &cfg.Enabled, &cfg.UsePlatform); err != nil {
				return nil, err
			}
			byName[cfg.Provider] = cfg
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	out := make([]OAuthProviderConfig, 0, len(AllSocialProviders))
	for _, name := range AllSocialProviders {
		if cfg, ok := byName[name]; ok {
			out = append(out, cfg)
			continue
		}
		out = append(out, OAuthProviderConfig{Provider: name, UsePlatform: true})
	}
	return out, nil
}

func (s *OAuthProviders) SetEnabled(ctx context.Context, tenantID uuid.UUID, provider string, enabled bool) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE oauth_providers SET enabled = $3
		WHERE tenant_id = $1 AND provider = $2
	`, tenantID, provider, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO oauth_providers (id, tenant_id, provider, client_id, client_secret_encrypted, enabled, use_platform, created_at)
		VALUES (gen_random_uuid(), $1, $2, '', '', $3, TRUE, now())
	`, tenantID, provider, enabled)
	return err
}

func (s *OAuthProviders) RevertToPlatform(ctx context.Context, tenantID uuid.UUID, provider string, enabled bool) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE oauth_providers
		SET use_platform = TRUE, client_id = '', client_secret_encrypted = '', enabled = $3
		WHERE tenant_id = $1 AND provider = $2
	`, tenantID, provider, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO oauth_providers (id, tenant_id, provider, client_id, client_secret_encrypted, enabled, use_platform, created_at)
		VALUES (gen_random_uuid(), $1, $2, '', '', $3, TRUE, now())
	`, tenantID, provider, enabled)
	return err
}

func (s *OAuthProviders) TogglePlatform(ctx context.Context, tenantID uuid.UUID, provider string, enabled bool) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE oauth_providers
		SET enabled = $3, use_platform = TRUE
		WHERE tenant_id = $1 AND provider = $2
	`, tenantID, provider, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO oauth_providers (id, tenant_id, provider, client_id, client_secret_encrypted, enabled, use_platform, created_at)
		VALUES (gen_random_uuid(), $1, $2, '', '', $3, TRUE, now())
	`, tenantID, provider, enabled)
	return err
}

func (s *OAuthProviders) UpsertCustom(ctx context.Context, tenantID uuid.UUID, provider, clientID, clientSecret string, enabled bool) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE oauth_providers
		SET client_id = $3, client_secret_encrypted = $4, enabled = $5, use_platform = FALSE
		WHERE tenant_id = $1 AND provider = $2
	`, tenantID, provider, clientID, clientSecret, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO oauth_providers (id, tenant_id, provider, client_id, client_secret_encrypted, enabled, use_platform, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, FALSE, now())
	`, tenantID, provider, clientID, clientSecret, enabled)
	return err
}

func (s *OAuthProviders) Upsert(ctx context.Context, provider, clientID, clientSecret string, enabled bool) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE oauth_providers
		SET client_id = $2, client_secret_encrypted = $3, enabled = $4, use_platform = FALSE
		WHERE provider = $1 AND tenant_id IS NULL
	`, provider, clientID, clientSecret, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO oauth_providers (id, tenant_id, provider, client_id, client_secret_encrypted, enabled, use_platform, created_at)
		VALUES (gen_random_uuid(), NULL, $1, $2, $3, $4, FALSE, now())
	`, provider, clientID, clientSecret, enabled)
	return err
}

func (s *OAuthProviders) ListEnabledForTenant(ctx context.Context, tenantID uuid.UUID) ([]OAuthProviderConfig, error) {
	all, err := s.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	var out []OAuthProviderConfig
	for _, p := range all {
		if p.Enabled {
			out = append(out, p)
		}
	}
	return out, nil
}
