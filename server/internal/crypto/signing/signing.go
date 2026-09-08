package signing

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Key struct {
	ID           uuid.UUID
	KID          string
	Algorithm    string
	PrivateKeyPEM string
	PublicKeyPEM  string
	Active       bool
	CreatedAt    time.Time
}

func EnsureActiveKey(ctx context.Context, db *pgxpool.Pool) (*Key, error) {
	var count int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM signing_keys WHERE active = TRUE`).Scan(&count); err != nil {
		return nil, err
	}
	if count > 0 {
		return loadActive(ctx, db)
	}
	return createKey(ctx, db)
}

func loadActive(ctx context.Context, db *pgxpool.Pool) (*Key, error) {
	row := db.QueryRow(ctx, `
		SELECT id, kid, algorithm, private_key_pem, public_key_pem, active, created_at
		FROM signing_keys
		WHERE active = TRUE
		ORDER BY created_at DESC
		LIMIT 1
	`)
	var key Key
	if err := row.Scan(
		&key.ID,
		&key.KID,
		&key.Algorithm,
		&key.PrivateKeyPEM,
		&key.PublicKeyPEM,
		&key.Active,
		&key.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &key, nil
}

func createKey(ctx context.Context, db *pgxpool.Pool) (*Key, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateDER})
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})

	kid := uuid.NewString()
	id := uuid.New()
	now := time.Now().UTC()

	_, err = db.Exec(ctx, `
		INSERT INTO signing_keys (id, kid, algorithm, private_key_pem, public_key_pem, active, created_at)
		VALUES ($1, $2, 'RS256', $3, $4, TRUE, $5)
	`, id, kid, string(privatePEM), string(publicPEM), now)
	if err != nil {
		return nil, err
	}

	return &Key{
		ID:            id,
		KID:           kid,
		Algorithm:     "RS256",
		PrivateKeyPEM: string(privatePEM),
		PublicKeyPEM:  string(publicPEM),
		Active:        true,
		CreatedAt:     now,
	}, nil
}

func ParseRSAPrivate(pemText string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, errors.New("invalid private key pem")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return key, nil
}
