package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const sessionTTL = 7 * 24 * time.Hour

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	IP        string
	UserAgent string
	CreatedAt time.Time
}

type Sessions struct {
	db *pgxpool.Pool
}

func NewSessions(db *pgxpool.Pool) *Sessions {
	return &Sessions{db: db}
}

func (s *Sessions) Create(ctx context.Context, userID uuid.UUID, tokenHash, ip, userAgent string) (*Session, error) {
	id := uuid.New()
	expires := time.Now().UTC().Add(sessionTTL)
	_, err := s.db.Exec(ctx, `
		INSERT INTO sessions (id, user_id, token_hash, expires_at, ip, user_agent, created_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::inet, NULLIF($6, ''), now())
	`, id, userID, tokenHash, expires, ip, userAgent)
	if err != nil {
		return nil, err
	}
	return &Session{ID: id, UserID: userID, ExpiresAt: expires}, nil
}

func (s *Sessions) FindByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, user_id, expires_at, COALESCE(host(ip), ''), COALESCE(user_agent, ''), created_at
		FROM sessions
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
	`, tokenHash)
	var sess Session
	if err := row.Scan(&sess.ID, &sess.UserID, &sess.ExpiresAt, &sess.IP, &sess.UserAgent, &sess.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sess, nil
}

func (s *Sessions) ListActiveForUser(ctx context.Context, userID uuid.UUID) ([]Session, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, expires_at, COALESCE(host(ip), ''), COALESCE(user_agent, ''), created_at
		FROM sessions
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > now()
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		if err := rows.Scan(&sess.ID, &sess.UserID, &sess.ExpiresAt, &sess.IP, &sess.UserAgent, &sess.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}

func (s *Sessions) RevokeForUser(ctx context.Context, userID, sessionID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE sessions SET revoked_at = now()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, sessionID, userID)
	return err
}

func (s *Sessions) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	return err
}

func (s *Sessions) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	return err
}
