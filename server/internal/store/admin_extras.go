package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Users) List(ctx context.Context, limit int) ([]User, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, email, email_verified_at, disabled_at, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.EmailVerifiedAt, &u.DisabledAt, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Users) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]User, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.email, u.email_verified_at, u.disabled_at, u.created_at
		FROM tenant_users tu
		JOIN users u ON u.id = tu.user_id
		WHERE tu.tenant_id = $1
		ORDER BY u.created_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.EmailVerifiedAt, &u.DisabledAt, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

type TenantUserAuth struct {
	User
	HasPassword  bool
	Providers    []string
	SignupMethod *string
}

func (s *Users) ListAppEndUsersByTenant(ctx context.Context, tenantID, ownerUserID uuid.UUID, limit int) ([]TenantUserAuth, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(ctx, `
		SELECT
			u.id,
			u.email,
			u.email_verified_at,
			u.disabled_at,
			u.created_at,
			EXISTS(
				SELECT 1 FROM credentials c
				WHERE c.user_id = u.id AND c.type = 'password'
			) AS has_password,
			COALESCE(
				array_agg(DISTINCT si.provider) FILTER (WHERE si.provider IS NOT NULL),
				'{}'
			) AS providers,
			tu.signup_method
		FROM tenant_users tu
		JOIN users u ON u.id = tu.user_id
		LEFT JOIN social_identities si ON si.user_id = u.id
		WHERE tu.tenant_id = $1
		  AND u.id != $3
		  AND NOT u.platform_owner
		GROUP BY u.id, tu.signup_method
		ORDER BY u.created_at DESC
		LIMIT $2
	`, tenantID, limit, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TenantUserAuth
	for rows.Next() {
		var row TenantUserAuth
		if err := rows.Scan(
			&row.ID,
			&row.Email,
			&row.EmailVerifiedAt,
			&row.DisabledAt,
			&row.CreatedAt,
			&row.HasPassword,
			&row.Providers,
			&row.SignupMethod,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *Users) EnsureTenant(ctx context.Context, userID, tenantID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET tenant_id = $2, updated_at = now()
		WHERE id = $1 AND (tenant_id IS NULL OR tenant_id = $2)
	`, userID, tenantID)
	return err
}

type SessionRow struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Email     string
	IP        *string
	UserAgent *string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (s *Sessions) ListActive(ctx context.Context, limit int) ([]SessionRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.user_id, u.email, host(s.ip)::text, s.user_agent, s.created_at, s.expires_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.revoked_at IS NULL AND s.expires_at > now()
		ORDER BY s.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SessionRow
	for rows.Next() {
		var row SessionRow
		if err := rows.Scan(&row.ID, &row.UserID, &row.Email, &row.IP, &row.UserAgent, &row.CreatedAt, &row.ExpiresAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *Sessions) ListActiveByTenant(ctx context.Context, tenantID, ownerID uuid.UUID, limit int) ([]SessionRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.user_id, u.email, host(s.ip)::text, s.user_agent, s.created_at, s.expires_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.revoked_at IS NULL AND s.expires_at > now()
		  AND (u.tenant_id = $1 OR u.id = $2)
		ORDER BY s.created_at DESC
		LIMIT $3
	`, tenantID, ownerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SessionRow
	for rows.Next() {
		var row SessionRow
		if err := rows.Scan(&row.ID, &row.UserID, &row.Email, &row.IP, &row.UserAgent, &row.CreatedAt, &row.ExpiresAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *Sessions) RevokeByIDForTenant(ctx context.Context, id, tenantID, ownerID uuid.UUID) (bool, error) {
	tag, err := s.db.Exec(ctx, `
		UPDATE sessions s SET revoked_at = now()
		FROM users u
		WHERE s.id = $1 AND s.user_id = u.id AND s.revoked_at IS NULL
		  AND (u.tenant_id = $2 OR u.id = $3)
	`, id, tenantID, ownerID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

type AuditEntry struct {
	ID        int64
	UserID    *uuid.UUID
	Action    string
	Meta      map[string]any
	IP        *string
	CreatedAt time.Time
}

func (a *Audit) ListByTenant(ctx context.Context, tenantID, ownerID uuid.UUID, limit int) ([]AuditEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := a.db.Query(ctx, `
		SELECT a.id, a.user_id, a.action, a.meta, host(a.ip)::text, a.created_at
		FROM audit_log a
		LEFT JOIN users u ON u.id = a.user_id
		WHERE a.user_id = $2 OR u.tenant_id = $1
		ORDER BY a.created_at DESC
		LIMIT $3
	`, tenantID, ownerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		var metaRaw []byte
		if err := rows.Scan(&e.ID, &e.UserID, &e.Action, &metaRaw, &e.IP, &e.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(metaRaw, &e.Meta)
		out = append(out, e)
	}
	return out, rows.Err()
}

type Theme struct {
	BrandName    string
	LogoURL      string
	AccentColor  string
	HideBranding bool
}

type ThemeStore struct {
	db *pgxpool.Pool
}

func NewThemeStore(db *pgxpool.Pool) *ThemeStore {
	return &ThemeStore{db: db}
}

func (t *ThemeStore) Get(ctx context.Context) (Theme, error) {
	row := t.db.QueryRow(ctx, `
		SELECT brand_name, COALESCE(logo_url, ''), accent_color FROM theme_config WHERE id = 1
	`)
	var theme Theme
	if err := row.Scan(&theme.BrandName, &theme.LogoURL, &theme.AccentColor); err != nil {
		return Theme{BrandName: "sooauth", AccentColor: "#FF3B3B"}, err
	}
	return theme, nil
}

func (t *ThemeStore) Update(ctx context.Context, brandName, logoURL, accent string) error {
	_, err := t.db.Exec(ctx, `
		UPDATE theme_config
		SET brand_name = $1, logo_url = NULLIF($2, ''), accent_color = $3, updated_at = now()
		WHERE id = 1
	`, brandName, logoURL, accent)
	return err
}

func (t *ThemeStore) GetByTenant(ctx context.Context, tenantID uuid.UUID) (Theme, error) {
	row := t.db.QueryRow(ctx, `
		SELECT brand_name, COALESCE(logo_url, ''), accent_color, hide_branding
		FROM tenant_themes WHERE tenant_id = $1
	`, tenantID)
	var theme Theme
	if err := row.Scan(&theme.BrandName, &theme.LogoURL, &theme.AccentColor, &theme.HideBranding); err != nil {
		return Theme{BrandName: "sooauth", AccentColor: "#FF3B3B"}, err
	}
	return theme, nil
}

func (t *ThemeStore) EnsureDefault(ctx context.Context, tenantID uuid.UUID) error {
	_, err := t.db.Exec(ctx, `
		INSERT INTO tenant_themes (tenant_id)
		VALUES ($1)
		ON CONFLICT (tenant_id) DO NOTHING
	`, tenantID)
	return err
}

func (t *ThemeStore) UpdateByTenant(ctx context.Context, tenantID uuid.UUID, brandName, logoURL, accent string, hideBranding bool) error {
	_, err := t.db.Exec(ctx, `
		INSERT INTO tenant_themes (tenant_id, brand_name, logo_url, accent_color, hide_branding)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5)
		ON CONFLICT (tenant_id) DO UPDATE SET
			brand_name = EXCLUDED.brand_name,
			logo_url = EXCLUDED.logo_url,
			accent_color = EXCLUDED.accent_color,
			hide_branding = EXCLUDED.hide_branding,
			updated_at = now()
	`, tenantID, brandName, logoURL, accent, hideBranding)
	return err
}
