package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const authCodeTTL = 10 * time.Minute

type AuthorizationCode struct {
	ID                    uuid.UUID
	ClientID              string
	UserID                uuid.UUID
	RedirectURI           string
	Scope                 string
	Nonce                 *string
	CodeChallenge         string
	CodeChallengeMethod   string
	ExpiresAt             time.Time
	UsedAt                *time.Time
}

type AuthorizationCodes struct {
	db *pgxpool.Pool
}

func NewAuthorizationCodes(db *pgxpool.Pool) *AuthorizationCodes {
	return &AuthorizationCodes{db: db}
}

func (s *AuthorizationCodes) Create(
	ctx context.Context,
	clientID string,
	userID uuid.UUID,
	redirectURI, scope, nonce, codeChallenge, codeChallengeMethod, codeHash string,
) error {
	expires := time.Now().UTC().Add(authCodeTTL)
	var nonceVal *string
	if nonce != "" {
		nonceVal = &nonce
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO authorization_codes (
			id, code_hash, client_id, user_id, redirect_uri, scope, nonce,
			code_challenge, code_challenge_method, expires_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
	`, uuid.New(), codeHash, clientID, userID, redirectURI, scope, nonceVal, codeChallenge, codeChallengeMethod, expires)
	return err
}

func (s *AuthorizationCodes) Consume(ctx context.Context, codeHash string) (*AuthorizationCode, error) {
	row := s.db.QueryRow(ctx, `
		UPDATE authorization_codes
		SET used_at = now()
		WHERE code_hash = $1
		  AND used_at IS NULL
		  AND expires_at > now()
		RETURNING id, client_id, user_id, redirect_uri, scope, nonce,
		          code_challenge, code_challenge_method, expires_at, used_at
	`, codeHash)
	var ac AuthorizationCode
	var nonce *string
	var usedAt *time.Time
	if err := row.Scan(
		&ac.ID,
		&ac.ClientID,
		&ac.UserID,
		&ac.RedirectURI,
		&ac.Scope,
		&nonce,
		&ac.CodeChallenge,
		&ac.CodeChallengeMethod,
		&ac.ExpiresAt,
		&usedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	ac.Nonce = nonce
	ac.UsedAt = usedAt
	return &ac, nil
}
