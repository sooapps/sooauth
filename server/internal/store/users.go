package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sooapps/sooauth/server/internal/crypto/password"
)

var ErrEmailTaken = errors.New("email already registered")

type User struct {
	ID              uuid.UUID
	TenantID        *uuid.UUID
	Email           string
	EmailVerifiedAt *time.Time
	DisabledAt      *time.Time
	CreatedAt       time.Time
}

type Users struct {
	db *pgxpool.Pool
}

func NewUsers(db *pgxpool.Pool) *Users {
	return &Users{db: db}
}

func (s *Users) CreatePlatformUser(ctx context.Context, email, plainPassword string) (*User, error) {
	return s.createWithPassword(ctx, nil, email, plainPassword, true)
}

func (s *Users) CreateAppUser(ctx context.Context, tenantID uuid.UUID, email, plainPassword string) (*User, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant required")
	}
	return s.createWithPassword(ctx, &tenantID, email, plainPassword, false)
}

func (s *Users) CreateWithPassword(ctx context.Context, email, plainPassword string, platformOwner bool) (*User, error) {
	if platformOwner {
		return s.CreatePlatformUser(ctx, email, plainPassword)
	}
	return nil, errors.New("app users require tenant scope")
}

func (s *Users) createWithPassword(ctx context.Context, tenantID *uuid.UUID, email, plainPassword string, platformOwner bool) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, errors.New("email required")
	}

	hash, err := password.Hash(plainPassword)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	userID := uuid.New()
	now := time.Now().UTC()

	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, platform_owner, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`, userID, tenantID, email, platformOwner, now)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO credentials (id, user_id, type, password_hash, created_at)
		VALUES ($1, $2, 'password', $3, $4)
	`, uuid.New(), userID, hash, now)
	if err != nil {
		return nil, err
	}

	if tenantID != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO tenant_users (tenant_id, user_id)
			VALUES ($1, $2)
			ON CONFLICT (tenant_id, user_id) DO NOTHING
		`, *tenantID, userID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &User{
		ID:        userID,
		TenantID:  tenantID,
		Email:     email,
		CreatedAt: now,
	}, nil
}

func (s *Users) FindPlatformByEmail(ctx context.Context, email string) (*User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	row := s.db.QueryRow(ctx, `
		SELECT u.id, u.tenant_id, u.email, u.email_verified_at, u.disabled_at, u.created_at, c.password_hash
		FROM users u
		JOIN credentials c ON c.user_id = u.id AND c.type = 'password'
		WHERE u.email = $1 AND u.platform_owner = TRUE
	`, email)

	return scanUserWithHash(row)
}

func (s *Users) FindByEmailInTenant(ctx context.Context, tenantID uuid.UUID, email string) (*User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	row := s.db.QueryRow(ctx, `
		SELECT u.id, u.tenant_id, u.email, u.email_verified_at, u.disabled_at, u.created_at, c.password_hash
		FROM users u
		JOIN credentials c ON c.user_id = u.id AND c.type = 'password'
		WHERE u.tenant_id = $1 AND u.email = $2 AND NOT u.platform_owner
	`, tenantID, email)

	return scanUserWithHash(row)
}

func (s *Users) FindByEmail(ctx context.Context, email string) (*User, string, error) {
	return s.FindPlatformByEmail(ctx, email)
}

func (s *Users) FindByEmailSimpleInTenant(ctx context.Context, tenantID uuid.UUID, email string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	row := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, email, email_verified_at, disabled_at, created_at
		FROM users
		WHERE tenant_id = $1 AND email = $2 AND NOT platform_owner
	`, tenantID, email)
	return scanUser(row)
}

func (s *Users) FindPlatformByEmailSimple(ctx context.Context, email string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	row := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, email, email_verified_at, disabled_at, created_at
		FROM users
		WHERE email = $1 AND platform_owner = TRUE
	`, email)
	return scanUser(row)
}

func (s *Users) FindByEmailSimple(ctx context.Context, email string) (*User, error) {
	return s.FindPlatformByEmailSimple(ctx, email)
}

func scanUserWithHash(row pgx.Row) (*User, string, error) {
	var user User
	var hash string
	if err := row.Scan(&user.ID, &user.TenantID, &user.Email, &user.EmailVerifiedAt, &user.DisabledAt, &user.CreatedAt, &hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", err
	}
	return &user, hash, nil
}

func scanUser(row pgx.Row) (*User, error) {
	var user User
	if err := row.Scan(&user.ID, &user.TenantID, &user.Email, &user.EmailVerifiedAt, &user.DisabledAt, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *Users) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, email, email_verified_at, disabled_at, created_at FROM users WHERE id = $1
	`, id)
	return scanUser(row)
}

func (s *Users) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET email_verified_at = now(), updated_at = now()
		WHERE id = $1 AND email_verified_at IS NULL
	`, userID)
	return err
}

func (s *Users) UpdatePassword(ctx context.Context, userID uuid.UUID, plainPassword string) error {
	hash, err := password.Hash(plainPassword)
	if err != nil {
		return err
	}
	tag, err := s.db.Exec(ctx, `
		UPDATE credentials SET password_hash = $2
		WHERE user_id = $1 AND type = 'password'
	`, userID, hash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO credentials (id, user_id, type, password_hash, created_at)
		VALUES ($1, $2, 'password', $3, now())
	`, uuid.New(), userID, hash)
	return err
}

func (s *Users) VerifyPassword(ctx context.Context, userID uuid.UUID, plainPassword string) (bool, error) {
	var hash string
	err := s.db.QueryRow(ctx, `
		SELECT password_hash FROM credentials WHERE user_id = $1 AND type = 'password'
	`, userID).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return password.Verify(plainPassword, hash)
}

func (s *Users) HasPasswordCredential(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM credentials WHERE user_id = $1 AND type = 'password'
		)
	`, userID).Scan(&exists)
	return exists, err
}

func (s *Users) DeleteUnverified(ctx context.Context, userID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		DELETE FROM users WHERE id = $1 AND email_verified_at IS NULL
	`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	_, err = tx.Exec(ctx, `DELETE FROM credentials WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Users) IsPlatformOwner(ctx context.Context, userID uuid.UUID) (bool, error) {
	var ok bool
	err := s.db.QueryRow(ctx, `SELECT platform_owner FROM users WHERE id = $1`, userID).Scan(&ok)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return ok, nil
}

func (s *Users) PromoteToPlatformOwner(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET platform_owner = TRUE, tenant_id = NULL, updated_at = now() WHERE id = $1
	`, userID)
	return err
}

func (s *Users) DeleteOrphanAppUser(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM users
		WHERE id = $1
		  AND NOT platform_owner
		  AND NOT EXISTS (SELECT 1 FROM tenant_users WHERE user_id = $1)
		  AND NOT EXISTS (SELECT 1 FROM tenants WHERE owner_user_id = $1)
	`, userID)
	return err
}

func (s *Users) SetDisabledForTenant(ctx context.Context, tenantID, userID uuid.UUID, disabled bool) (bool, error) {
	var query string
	if disabled {
		query = `UPDATE users u SET disabled_at = COALESCE(u.disabled_at, now()), updated_at = now()
			WHERE u.id = $1 AND (u.tenant_id = $2 OR EXISTS (SELECT 1 FROM tenant_users tu WHERE tu.tenant_id = $2 AND tu.user_id = u.id)) AND NOT u.platform_owner`
	} else {
		query = `UPDATE users u SET disabled_at = NULL, updated_at = now()
			WHERE u.id = $1 AND (u.tenant_id = $2 OR EXISTS (SELECT 1 FROM tenant_users tu WHERE tu.tenant_id = $2 AND tu.user_id = u.id)) AND NOT u.platform_owner`
	}
	tag, err := s.db.Exec(ctx, query, userID, tenantID)
	return tag.RowsAffected() > 0, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
