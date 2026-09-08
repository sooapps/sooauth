package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebAuthnCredentials struct {
	db *pgxpool.Pool
}

type WebAuthnCredential struct {
	ID           uuid.UUID `json:"id"`
	CredentialID string    `json:"credential_id"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewWebAuthnCredentials(db *pgxpool.Pool) *WebAuthnCredentials {
	return &WebAuthnCredentials{db: db}
}

type webAuthnUser struct {
	id          uuid.UUID
	email       string
	credentials []webauthn.Credential
}

func (u webAuthnUser) WebAuthnID() []byte                         { return []byte(u.id.String()) }
func (u webAuthnUser) WebAuthnName() string                       { return u.email }
func (u webAuthnUser) WebAuthnDisplayName() string                { return u.email }
func (u webAuthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

func (s *WebAuthnCredentials) LoadUser(ctx context.Context, userID uuid.UUID) (webauthn.User, error) {
	user, err := NewUsers(s.db).FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, err
	}
	creds, err := s.listForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return webAuthnUser{id: userID, email: user.Email, credentials: creds}, nil
}

func (s *WebAuthnCredentials) SaveCredential(ctx context.Context, userID uuid.UUID, cred webauthn.Credential) error {
	raw, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO credentials (id, user_id, type, webauthn_credential, created_at)
		VALUES ($1, $2, 'webauthn', $3::jsonb, now())
	`, uuid.New(), userID, string(raw))
	return err
}

func (s *WebAuthnCredentials) listForUser(ctx context.Context, userID uuid.UUID) ([]webauthn.Credential, error) {
	rows, err := s.db.Query(ctx, `
		SELECT webauthn_credential
		FROM credentials
		WHERE user_id = $1 AND type = 'webauthn'
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []webauthn.Credential
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var cred webauthn.Credential
		if err := json.Unmarshal(raw, &cred); err != nil {
			return nil, err
		}
		out = append(out, cred)
	}
	return out, rows.Err()
}

func (s *WebAuthnCredentials) HasCredentials(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM credentials WHERE user_id = $1 AND type = 'webauthn'
	`, userID).Scan(&count)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return count > 0, err
}

func (s *WebAuthnCredentials) ListForUser(ctx context.Context, userID uuid.UUID) ([]WebAuthnCredential, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, webauthn_credential->>'id', created_at
		FROM credentials
		WHERE user_id = $1 AND type = 'webauthn'
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []WebAuthnCredential
	for rows.Next() {
		var credential WebAuthnCredential
		if err := rows.Scan(&credential.ID, &credential.CredentialID, &credential.CreatedAt); err != nil {
			return nil, err
		}
		credentials = append(credentials, credential)
	}
	return credentials, rows.Err()
}

func (s *WebAuthnCredentials) DeleteForUser(ctx context.Context, userID, credentialID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM credentials WHERE id = $1 AND user_id = $2 AND type = 'webauthn'
	`, credentialID, userID)
	return err
}
