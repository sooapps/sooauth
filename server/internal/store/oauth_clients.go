package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OAuthClient struct {
	ID               uuid.UUID
	TenantID         *uuid.UUID
	ClientID         string
	ClientSecretHash *string
	Name             string
	RedirectURIs     []string
	AllowedScopes    []string
	Public           bool
}

type OAuthClientInput struct {
	TenantID     uuid.UUID
	ClientID     string
	Name         string
	RedirectURIs []string
	Public       bool
}

type OAuthClients struct {
	db *pgxpool.Pool
}

var ErrInvalidClient = errors.New("invalid client")

func NewOAuthClients(db *pgxpool.Pool) *OAuthClients {
	return &OAuthClients{db: db}
}

func (s *OAuthClients) scanClient(row pgx.Row) (*OAuthClient, error) {
	var c OAuthClient
	if err := row.Scan(
		&c.ID,
		&c.TenantID,
		&c.ClientID,
		&c.ClientSecretHash,
		&c.Name,
		&c.RedirectURIs,
		&c.AllowedScopes,
		&c.Public,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *OAuthClients) FindByClientID(ctx context.Context, clientID string) (*OAuthClient, error) {
	return s.scanClient(s.db.QueryRow(ctx, `
		SELECT id, tenant_id, client_id, client_secret_hash, name, redirect_uris, allowed_scopes, public
		FROM oauth_clients WHERE client_id = $1
	`, clientID))
}

func (s *OAuthClients) FindPrimaryByTenant(ctx context.Context, tenantID uuid.UUID) (*OAuthClient, error) {
	return s.scanClient(s.db.QueryRow(ctx, `
		SELECT id, tenant_id, client_id, client_secret_hash, name, redirect_uris, allowed_scopes, public
		FROM oauth_clients WHERE tenant_id = $1
		ORDER BY created_at ASC LIMIT 1
	`, tenantID))
}

func (s *OAuthClients) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]OAuthClient, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, client_id, client_secret_hash, name, redirect_uris, allowed_scopes, public
		FROM oauth_clients WHERE tenant_id = $1
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []OAuthClient
	for rows.Next() {
		var c OAuthClient
		if err := rows.Scan(&c.ID, &c.TenantID, &c.ClientID, &c.ClientSecretHash, &c.Name, &c.RedirectURIs, &c.AllowedScopes, &c.Public); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *OAuthClients) Create(ctx context.Context, in OAuthClientInput) (*OAuthClient, error) {
	id := uuid.New()
	row := s.db.QueryRow(ctx, `
		INSERT INTO oauth_clients (id, tenant_id, client_id, name, redirect_uris, public)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, tenant_id, client_id, client_secret_hash, name, redirect_uris, allowed_scopes, public
	`, id, in.TenantID, in.ClientID, in.Name, in.RedirectURIs, in.Public)
	return s.scanClient(row)
}

func (s *OAuthClients) Update(ctx context.Context, tenantID uuid.UUID, clientID string, name string, redirectURIs []string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE oauth_clients SET name = $3, redirect_uris = $4
		WHERE tenant_id = $1 AND client_id = $2
	`, tenantID, clientID, name, redirectURIs)
	return err
}

func (c *OAuthClient) AllowsRedirectURI(uri string) bool {
	for _, allowed := range c.RedirectURIs {
		if allowed == uri {
			return true
		}
	}
	return false
}

func (c *OAuthClient) AllowsScope(requested string) bool {
	if requested == "" {
		return false
	}
	allowed := make(map[string]struct{}, len(c.AllowedScopes))
	for _, scope := range c.AllowedScopes {
		allowed[scope] = struct{}{}
	}
	for _, part := range splitScopes(requested) {
		if _, ok := allowed[part]; !ok {
			return false
		}
	}
	return true
}

func splitScopes(scope string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(scope); i++ {
		if i == len(scope) || scope[i] == ' ' {
			if i > start {
				out = append(out, scope[start:i])
			}
			start = i + 1
		}
	}
	return out
}
