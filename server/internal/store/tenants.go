package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sooapps/sooauth/server/internal/emailverify"
	"github.com/sooapps/sooauth/server/internal/passwordpolicy"
)

type Tenant struct {
	ID                     uuid.UUID
	AccountID              *uuid.UUID
	Name                   string
	OwnerUserID            uuid.UUID
	EmailVerifyRequired    bool
	EmailVerifyDelivery    string
	PasswordMinLength      int
	PasswordRequireUpper   bool
	PasswordRequireNumber  bool
	PasswordRequireSpecial bool
	SocialCallbackOrigin   string
	CreatedAt              time.Time
}

func (t Tenant) VerifyDelivery() emailverify.Delivery {
	return emailverify.ParseDelivery(t.EmailVerifyDelivery)
}

func (t Tenant) PasswordPolicy() passwordpolicy.Policy {
	return passwordpolicy.Policy{
		MinLength:        t.PasswordMinLength,
		RequireUppercase: t.PasswordRequireUpper,
		RequireNumber:    t.PasswordRequireNumber,
		RequireSpecial:   t.PasswordRequireSpecial,
	}.Normalize()
}

type Tenants struct {
	db *pgxpool.Pool
}

func NewTenants(db *pgxpool.Pool) *Tenants {
	return &Tenants{db: db}
}

func (t *Tenants) scanTenant(row pgx.Row) (*Tenant, error) {
	var tenant Tenant
	if err := row.Scan(
		&tenant.ID,
		&tenant.AccountID,
		&tenant.Name,
		&tenant.OwnerUserID,
		&tenant.EmailVerifyRequired,
		&tenant.EmailVerifyDelivery,
		&tenant.PasswordMinLength,
		&tenant.PasswordRequireUpper,
		&tenant.PasswordRequireNumber,
		&tenant.PasswordRequireSpecial,
		&tenant.SocialCallbackOrigin,
		&tenant.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (t *Tenants) FindByOwner(ctx context.Context, ownerID uuid.UUID) (*Tenant, error) {
	return t.scanTenant(t.db.QueryRow(ctx, `
		SELECT id, account_id, name, owner_user_id, email_verify_required, email_verify_delivery,
		       password_min_length, password_require_uppercase, password_require_number, password_require_special,
		       social_callback_origin, created_at
		FROM tenants WHERE owner_user_id = $1
		ORDER BY created_at ASC LIMIT 1
	`, ownerID))
}

func (t *Tenants) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]Tenant, error) {
	rows, err := t.db.Query(ctx, `
		SELECT id, account_id, name, owner_user_id, email_verify_required, email_verify_delivery,
		       password_min_length, password_require_uppercase, password_require_number, password_require_special,
		       social_callback_origin, created_at
		FROM tenants WHERE account_id = $1
		ORDER BY created_at ASC
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Tenant
	for rows.Next() {
		var tenant Tenant
		if err := rows.Scan(
			&tenant.ID,
			&tenant.AccountID,
			&tenant.Name,
			&tenant.OwnerUserID,
			&tenant.EmailVerifyRequired,
			&tenant.EmailVerifyDelivery,
			&tenant.PasswordMinLength,
			&tenant.PasswordRequireUpper,
			&tenant.PasswordRequireNumber,
			&tenant.PasswordRequireSpecial,
			&tenant.SocialCallbackOrigin,
			&tenant.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, tenant)
	}
	return out, rows.Err()
}

func (t *Tenants) CountByAccount(ctx context.Context, accountID uuid.UUID) (int, error) {
	var n int
	err := t.db.QueryRow(ctx, `SELECT COUNT(*) FROM tenants WHERE account_id = $1`, accountID).Scan(&n)
	return n, err
}

func (t *Tenants) FindByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	return t.scanTenant(t.db.QueryRow(ctx, `
		SELECT id, account_id, name, owner_user_id, email_verify_required, email_verify_delivery,
		       password_min_length, password_require_uppercase, password_require_number, password_require_special,
		       social_callback_origin, created_at
		FROM tenants WHERE id = $1
	`, id))
}

func (t *Tenants) FindByIDForAccount(ctx context.Context, id, accountID uuid.UUID) (*Tenant, error) {
	return t.scanTenant(t.db.QueryRow(ctx, `
		SELECT id, account_id, name, owner_user_id, email_verify_required, email_verify_delivery,
		       password_min_length, password_require_uppercase, password_require_number, password_require_special,
		       social_callback_origin, created_at
		FROM tenants WHERE id = $1 AND account_id = $2
	`, id, accountID))
}

func (t *Tenants) Create(ctx context.Context, accountID, ownerID uuid.UUID, name string) (*Tenant, error) {
	id := uuid.New()
	now := time.Now().UTC()
	_, err := t.db.Exec(ctx, `
		INSERT INTO tenants (id, account_id, name, owner_user_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, id, accountID, name, ownerID, now)
	if err != nil {
		return nil, err
	}
	return &Tenant{
		ID:                     id,
		AccountID:              &accountID,
		Name:                   name,
		OwnerUserID:            ownerID,
		EmailVerifyRequired:    true,
		EmailVerifyDelivery:    string(emailverify.DeliveryLink),
		PasswordMinLength:      8,
		PasswordRequireUpper:   false,
		PasswordRequireNumber:  false,
		PasswordRequireSpecial: false,
		CreatedAt:              now,
	}, nil
}

func (t *Tenants) UpdateSettings(ctx context.Context, id uuid.UUID, name string, emailVerifyRequired bool, delivery emailverify.Delivery, policy passwordpolicy.Policy) error {
	if !delivery.Valid() {
		delivery = emailverify.DeliveryLink
	}
	policy = policy.Normalize()
	_, err := t.db.Exec(ctx, `
		UPDATE tenants SET
			name = $2,
			email_verify_required = $3,
			email_verify_delivery = $4,
			password_min_length = $5,
			password_require_uppercase = $6,
			password_require_number = $7,
			password_require_special = $8
		WHERE id = $1
	`, id, name, emailVerifyRequired, delivery.String(), policy.MinLength, policy.RequireUppercase, policy.RequireNumber, policy.RequireSpecial)
	return err
}

func (t *Tenants) UpdateSocialCallbackOrigin(ctx context.Context, id uuid.UUID, origin string) error {
	_, err := t.db.Exec(ctx, `
		UPDATE tenants SET social_callback_origin = $2 WHERE id = $1
	`, id, strings.TrimSpace(origin))
	return err
}
