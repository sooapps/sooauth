package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sooapps/sooauth/server/internal/crypto/encryption"
)

type MFA struct {
	db     *pgxpool.Pool
	cipher *encryption.Cipher
}

func NewMFA(db *pgxpool.Pool, key []byte) *MFA {
	cipher, _ := encryption.New(key)
	return &MFA{db: db, cipher: cipher}
}

func (s *MFA) PendingSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	if s.cipher == nil {
		return encryption.ErrInvalidKey
	}
	protected, err := s.cipher.Encrypt(secret)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO platform_mfa (user_id, secret) VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET secret = EXCLUDED.secret, enabled_at = NULL, updated_at = now()
	`, userID, protected)
	return err
}

func (s *MFA) Secret(ctx context.Context, userID uuid.UUID) (secret string, enabled bool, err error) {
	var stored string
	err = s.db.QueryRow(ctx, `SELECT secret, enabled_at IS NOT NULL FROM platform_mfa WHERE user_id = $1`, userID).Scan(&stored, &enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if encryption.IsEncrypted(stored) {
		if s.cipher == nil {
			return "", false, encryption.ErrInvalidKey
		}
		secret, err = s.cipher.Decrypt(stored)
		return secret, enabled, err
	}
	// Read legacy plaintext rows and upgrade them without blocking authentication.
	secret = stored
	if s.cipher != nil {
		if protected, encryptErr := s.cipher.Encrypt(secret); encryptErr == nil {
			_, _ = s.db.Exec(ctx, `UPDATE platform_mfa SET secret = $2, updated_at = now() WHERE user_id = $1 AND secret = $3`, userID, protected, stored)
		}
	}
	return secret, enabled, nil
}

func (s *MFA) Enable(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `UPDATE platform_mfa SET enabled_at = now(), updated_at = now() WHERE user_id = $1`, userID)
	return err
}

func (s *MFA) Disable(ctx context.Context, userID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM platform_mfa_backup_codes WHERE user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM platform_mfa WHERE user_id = $1`, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *MFA) ReplaceBackupCodes(ctx context.Context, userID uuid.UUID, hashes []string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM platform_mfa_backup_codes WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, hash := range hashes {
		if _, err = tx.Exec(ctx, `INSERT INTO platform_mfa_backup_codes (id, user_id, code_hash) VALUES ($1, $2, $3)`, uuid.New(), userID, hash); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *MFA) ConsumeBackupCode(ctx context.Context, userID uuid.UUID, hash string) (bool, error) {
	tag, err := s.db.Exec(ctx, `UPDATE platform_mfa_backup_codes SET used_at = now() WHERE user_id = $1 AND code_hash = $2 AND used_at IS NULL`, userID, hash)
	return tag.RowsAffected() == 1, err
}

func (s *MFA) CreateChallenge(ctx context.Context, userID uuid.UUID, hash string, expires time.Time) error {
	_, err := s.db.Exec(ctx, `INSERT INTO platform_mfa_challenges (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)`, uuid.New(), userID, hash, expires)
	return err
}

func (s *MFA) ConsumeChallenge(ctx context.Context, hash string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := s.db.QueryRow(ctx, `
		UPDATE platform_mfa_challenges SET used_at = now()
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()
		RETURNING user_id
	`, hash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrTokenInvalid
	}
	return userID, err
}

func (s *MFA) ChallengeUser(ctx context.Context, hash string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := s.db.QueryRow(ctx, `
		SELECT user_id FROM platform_mfa_challenges
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()
	`, hash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrTokenInvalid
	}
	return userID, err
}
