package signing

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func JWKSFromDB(ctx context.Context, db *pgxpool.Pool) (JWKS, error) {
	rows, err := db.Query(ctx, `
		SELECT kid, algorithm, public_key_pem
		FROM signing_keys
		WHERE active = TRUE
		ORDER BY created_at DESC
	`)
	if err != nil {
		return JWKS{}, err
	}
	defer rows.Close()

	var keys []JWK
	for rows.Next() {
		var kid, alg, publicPEM string
		if err := rows.Scan(&kid, &alg, &publicPEM); err != nil {
			return JWKS{}, err
		}
		jwk, err := jwkFromPEM(kid, alg, publicPEM)
		if err != nil {
			return JWKS{}, err
		}
		keys = append(keys, jwk)
	}
	return JWKS{Keys: keys}, rows.Err()
}

func jwkFromPEM(kid, alg, publicPEM string) (JWK, error) {
	block, _ := pem.Decode([]byte(publicPEM))
	if block == nil {
		return JWK{}, pemDecodeError("public key")
	}
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return JWK{}, err
	}
	pub, ok := pubAny.(*rsa.PublicKey)
	if !ok {
		return JWK{}, pemDecodeError("rsa public key")
	}
	nBytes := make([]byte, (pub.N.BitLen()+7)/8)
	pub.N.FillBytes(nBytes)
	return JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: alg,
		N:   base64.RawURLEncoding.EncodeToString(nBytes),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}, nil
}

func pemDecodeError(msg string) error {
	return fmt.Errorf("jwks: %s", msg)
}
