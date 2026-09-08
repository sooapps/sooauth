package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrIdentityTaken  = errors.New("social identity already linked")
	ErrLastCredential = errors.New("at least one other credential is required")
	ErrIdentityScope  = errors.New("social identity scope does not match user")
)

type SocialIdentity struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TenantID    *uuid.UUID
	Provider    string
	ProviderSub string
	Email       string
}

type SocialIdentities struct {
	db *pgxpool.Pool
}

type LinkedIdentity struct {
	ID          uuid.UUID `json:"id"`
	Provider    string    `json:"provider"`
	ProviderSub string    `json:"-"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewSocialIdentities(db *pgxpool.Pool) *SocialIdentities {
	return &SocialIdentities{db: db}
}

func (s *SocialIdentities) FindByProviderSub(ctx context.Context, tenantID *uuid.UUID, provider, sub string) (*SocialIdentity, error) {
	var row pgx.Row
	if tenantID != nil {
		row = s.db.QueryRow(ctx, `
			SELECT id, user_id, tenant_id, provider, provider_sub, COALESCE(email, '')
			FROM social_identities
			WHERE tenant_id = $1 AND provider = $2 AND provider_sub = $3
		`, *tenantID, provider, sub)
	} else {
		row = s.db.QueryRow(ctx, `
			SELECT id, user_id, tenant_id, provider, provider_sub, COALESCE(email, '')
			FROM social_identities
			WHERE tenant_id IS NULL AND provider = $1 AND provider_sub = $2
		`, provider, sub)
	}
	var si SocialIdentity
	if err := row.Scan(&si.ID, &si.UserID, &si.TenantID, &si.Provider, &si.ProviderSub, &si.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &si, nil
}

func (s *SocialIdentities) Link(ctx context.Context, tenantID *uuid.UUID, userID uuid.UUID, provider, sub, email string) error {
	if err := s.validateScope(ctx, tenantID, userID); err != nil {
		return err
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO social_identities (id, user_id, tenant_id, provider, provider_sub, email, created_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), now())
	`, uuid.New(), userID, tenantID, provider, sub, email)
	if isUniqueViolation(err) {
		return ErrIdentityTaken
	}
	return err
}

func (s *SocialIdentities) ListForUser(ctx context.Context, tenantID *uuid.UUID, userID uuid.UUID) ([]LinkedIdentity, error) {
	if err := s.validateScope(ctx, tenantID, userID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, provider, provider_sub, COALESCE(email, ''), created_at
		FROM social_identities
		WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2
		ORDER BY created_at DESC
	`, userID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LinkedIdentity
	for rows.Next() {
		var identity LinkedIdentity
		if err := rows.Scan(&identity.ID, &identity.Provider, &identity.ProviderSub, &identity.Email, &identity.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, identity)
	}
	return out, rows.Err()
}

func (s *SocialIdentities) UnlinkForUser(ctx context.Context, tenantID *uuid.UUID, userID, identityID uuid.UUID) error {
	if err := s.validateScope(ctx, tenantID, userID); err != nil {
		return err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := tx.QueryRow(ctx, `SELECT id FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&userID); err != nil {
		return err
	}
	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM social_identities WHERE id = $1 AND user_id = $2 AND tenant_id IS NOT DISTINCT FROM $3)
	`, identityID, userID, tenantID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return pgx.ErrNoRows
	}
	var usable int
	if err := tx.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM credentials WHERE user_id = $1 AND type IN ('password', 'webauthn')) +
			(SELECT COUNT(*) FROM social_identities WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2)
	`, userID, tenantID).Scan(&usable); err != nil {
		return err
	}
	if usable <= 1 {
		return ErrLastCredential
	}
	if _, err := tx.Exec(ctx, `DELETE FROM social_identities WHERE id = $1 AND user_id = $2`, identityID, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *SocialIdentities) validateScope(ctx context.Context, tenantID *uuid.UUID, userID uuid.UUID) error {
	var platformOwner bool
	var rawTenant *uuid.UUID
	if err := s.db.QueryRow(ctx, `SELECT platform_owner, tenant_id FROM users WHERE id = $1`, userID).Scan(&platformOwner, &rawTenant); err != nil {
		return err
	}
	if tenantID == nil {
		if !platformOwner || rawTenant != nil {
			return ErrIdentityScope
		}
		return nil
	}
	if platformOwner || rawTenant == nil || *rawTenant != *tenantID {
		return ErrIdentityScope
	}
	return nil
}

func (s *Users) FindOrCreateSocialUser(ctx context.Context, tenantID *uuid.UUID, email, provider, sub string) (*User, bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	social := NewSocialIdentities(s.db)

	existing, err := social.FindByProviderSub(ctx, tenantID, provider, sub)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		user, err := s.FindByID(ctx, existing.UserID)
		return user, false, err
	}

	if email != "" {
		var user *User
		if tenantID != nil {
			user, _, err = s.FindByEmailInTenant(ctx, *tenantID, email)
		} else {
			user, err = s.FindPlatformByEmailSimple(ctx, email)
		}
		if err != nil {
			return nil, false, err
		}
		if user != nil {
			if err := social.Link(ctx, tenantID, user.ID, provider, sub, email); err != nil {
				return nil, false, err
			}
			if user.EmailVerifiedAt == nil {
				_ = s.MarkEmailVerified(ctx, user.ID)
				user, _ = s.FindByID(ctx, user.ID)
			}
			return user, false, nil
		}
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback(ctx)

	userID := uuid.New()
	now := time.Now().UTC()
	if email == "" {
		email = provider + "+" + sub + "@users.sooauth.local"
	}

	platformOwner := tenantID == nil
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, email_verified_at, platform_owner, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $4, $4)
	`, userID, tenantID, email, now, platformOwner)
	if err != nil {
		return nil, false, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO social_identities (id, user_id, tenant_id, provider, provider_sub, email, created_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7)
	`, uuid.New(), userID, tenantID, provider, sub, email, now)
	if err != nil {
		return nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}

	user, err := s.FindByID(ctx, userID)
	return user, true, err
}
