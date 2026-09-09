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
	ErrTokenInvalid = errors.New("token invalid or expired")
)

type VerificationTokens struct {
	db *pgxpool.Pool
}

func NewVerificationTokens(db *pgxpool.Pool) *VerificationTokens {
	return &VerificationTokens{db: db}
}

func (s *VerificationTokens) Create(
	ctx context.Context,
	userID uuid.UUID,
	email, tokenType, tokenHash string,
	expiresAt time.Time,
) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO verification_tokens (id, user_id, email, type, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
	`, uuid.New(), userID, email, tokenType, tokenHash, expiresAt)
	return err
}

func (s *VerificationTokens) Consume(ctx context.Context, tokenType, tokenHash string) (uuid.UUID, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	var id uuid.UUID
	var userID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id, user_id
		FROM verification_tokens
		WHERE type = $1 AND token_hash = $2 AND used_at IS NULL AND expires_at > now()
		FOR UPDATE
	`, tokenType, tokenHash).Scan(&id, &userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrTokenInvalid
		}
		return uuid.Nil, err
	}

	_, err = tx.Exec(ctx, `UPDATE verification_tokens SET used_at = now() WHERE id = $1`, id)
	if err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (s *VerificationTokens) ConsumeCode(ctx context.Context, email, tokenType, tokenHash string) (uuid.UUID, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	email = strings.ToLower(strings.TrimSpace(email))
	var id uuid.UUID
	var userID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id, user_id
		FROM verification_tokens
		WHERE type = $1 AND lower(email) = $2 AND token_hash = $3 AND used_at IS NULL AND expires_at > now()
		FOR UPDATE
	`, tokenType, email, tokenHash).Scan(&id, &userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrTokenInvalid
		}
		return uuid.Nil, err
	}

	_, err = tx.Exec(ctx, `UPDATE verification_tokens SET used_at = now() WHERE id = $1`, id)
	if err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (s *VerificationTokens) DeletePending(ctx context.Context, userID uuid.UUID, tokenTypes ...string) error {
	if len(tokenTypes) == 0 {
		return nil
	}
	_, err := s.db.Exec(ctx, `
		DELETE FROM verification_tokens
		WHERE user_id = $1 AND used_at IS NULL AND type = ANY($2)
	`, userID, tokenTypes)
	return err
}

func (s *VerificationTokens) UserFromUsed(ctx context.Context, tokenType, tokenHash string) (uuid.UUID, bool, error) {
	var userID uuid.UUID
	err := s.db.QueryRow(ctx, `
		SELECT user_id
		FROM verification_tokens
		WHERE type = $1 AND token_hash = $2 AND used_at IS NOT NULL
		ORDER BY used_at DESC
		LIMIT 1
	`, tokenType, tokenHash).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}
	return userID, true, nil
}
