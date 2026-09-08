package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantUsers struct {
	db *pgxpool.Pool
}

func NewTenantUsers(db *pgxpool.Pool) *TenantUsers {
	return &TenantUsers{db: db}
}

func (s *TenantUsers) Link(ctx context.Context, tenantID, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO tenant_users (tenant_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, tenantID, userID)
	return err
}

func (s *TenantUsers) LinkWithMethod(ctx context.Context, tenantID, userID uuid.UUID, method string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO tenant_users (tenant_id, user_id, signup_method)
		VALUES ($1, $2, NULLIF($3, ''))
		ON CONFLICT (tenant_id, user_id) DO NOTHING
	`, tenantID, userID, method)
	return err
}

func (s *TenantUsers) HasAccess(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM tenant_users WHERE tenant_id = $1 AND user_id = $2
		)
	`, tenantID, userID).Scan(&exists)
	return exists, err
}

func (s *TenantUsers) Unlink(ctx context.Context, tenantID, userID uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM tenant_users WHERE tenant_id = $1 AND user_id = $2
	`, tenantID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *TenantUsers) CountForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM tenant_users WHERE user_id = $1
	`, userID).Scan(&n)
	return n, err
}

func (s *TenantUsers) IsLinked(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	return s.HasAccess(ctx, tenantID, userID)
}
