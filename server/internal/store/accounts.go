package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Account struct {
	ID          uuid.UUID
	OwnerUserID uuid.UUID
	Plan        string
	CreatedAt   time.Time
}

type Accounts struct {
	db *pgxpool.Pool
}

func NewAccounts(db *pgxpool.Pool) *Accounts {
	return &Accounts{db: db}
}

func (a *Accounts) FindByOwner(ctx context.Context, ownerID uuid.UUID) (*Account, error) {
	row := a.db.QueryRow(ctx, `
		SELECT id, owner_user_id, plan, created_at FROM accounts WHERE owner_user_id = $1
	`, ownerID)
	var ac Account
	if err := row.Scan(&ac.ID, &ac.OwnerUserID, &ac.Plan, &ac.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ac, nil
}

func (a *Accounts) FindOrCreateByOwner(ctx context.Context, ownerID uuid.UUID) (*Account, error) {
	if existing, err := a.FindByOwner(ctx, ownerID); err != nil || existing != nil {
		return existing, err
	}
	id := uuid.New()
	now := time.Now().UTC()
	_, err := a.db.Exec(ctx, `
		INSERT INTO accounts (id, owner_user_id, plan, created_at)
		VALUES ($1, $2, 'free', $3)
	`, id, ownerID, now)
	if err != nil {
		return a.FindByOwner(ctx, ownerID)
	}
	return &Account{ID: id, OwnerUserID: ownerID, Plan: "free", CreatedAt: now}, nil
}

func (a *Accounts) FindByTenantID(ctx context.Context, tenantID uuid.UUID) (*Account, error) {
	row := a.db.QueryRow(ctx, `
		SELECT a.id, a.owner_user_id, a.plan, a.created_at
		FROM accounts a
		INNER JOIN tenants t ON t.account_id = a.id
		WHERE t.id = $1
	`, tenantID)
	var ac Account
	if err := row.Scan(&ac.ID, &ac.OwnerUserID, &ac.Plan, &ac.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ac, nil
}
