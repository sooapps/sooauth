package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const refreshTTL = 30 * 24 * time.Hour

type RefreshToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	FamilyID   uuid.UUID
	ReplacedBy *uuid.UUID
	ExpiresAt  time.Time
	RevokedAt  *time.Time
}

type RefreshTokens struct {
	db *pgxpool.Pool
}

func NewRefreshTokens(db *pgxpool.Pool) *RefreshTokens {
	return &RefreshTokens{db: db}
}

func (s *RefreshTokens) Issue(ctx context.Context, userID, familyID uuid.UUID, tokenHash string) (uuid.UUID, error) {
	return s.IssueWithTTL(ctx, userID, familyID, tokenHash, refreshTTL)
}

func (s *RefreshTokens) IssueWithTTL(ctx context.Context, userID, familyID uuid.UUID, tokenHash string, ttl time.Duration) (uuid.UUID, error) {
	if ttl <= 0 {
		ttl = refreshTTL
	}
	id := uuid.New()
	expires := time.Now().UTC().Add(ttl)
	_, err := s.db.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, family_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, now())
	`, id, userID, familyID, tokenHash, expires)
	return id, err
}

func (s *RefreshTokens) FindByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, user_id, family_id, replaced_by, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash)
	var rt RefreshToken
	var replacedBy *uuid.UUID
	if err := row.Scan(&rt.ID, &rt.UserID, &rt.FamilyID, &replacedBy, &rt.ExpiresAt, &rt.RevokedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	rt.ReplacedBy = replacedBy
	return &rt, nil
}

func (s *RefreshTokens) MarkReplaced(ctx context.Context, id, replacedBy uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE refresh_tokens
		SET replaced_by = $2, revoked_at = now()
		WHERE id = $1 AND revoked_at IS NULL
	`, id, replacedBy)
	return err
}

func (s *RefreshTokens) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE family_id = $1 AND revoked_at IS NULL
	`, familyID)
	return err
}

func (s *RefreshTokens) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	return err
}
