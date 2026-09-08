package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/crypto/signing"
)

const accessTTL = 15 * time.Minute

type Claims struct {
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Nonce         string `json:"nonce,omitempty"`
	jwtv5.RegisteredClaims
}

type Issuer struct {
	key    *rsa.PrivateKey
	kid    string
	issuer string
}

func NewIssuer(signingKey *signing.Key, issuerURL string) (*Issuer, error) {
	key, err := signing.ParseRSAPrivate(signingKey.PrivateKeyPEM)
	if err != nil {
		return nil, err
	}
	return &Issuer{key: key, kid: signingKey.KID, issuer: issuerURL}, nil
}

func (i *Issuer) AccessToken(userID uuid.UUID, email string) (string, time.Time, error) {
	expires := time.Now().UTC().Add(accessTTL)
	claims := Claims{
		Email: email,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    i.issuer,
			ExpiresAt: jwtv5.NewNumericDate(expires),
			IssuedAt:  jwtv5.NewNumericDate(time.Now().UTC()),
		},
	}
	return i.sign(claims, expires)
}

func (i *Issuer) IDToken(userID uuid.UUID, email string, clientID, nonce string, emailVerified bool) (string, time.Time, error) {
	expires := time.Now().UTC().Add(accessTTL)
	claims := Claims{
		Email:         email,
		EmailVerified: emailVerified,
		Nonce:         nonce,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject:   userID.String(),
			Audience:  jwtv5.ClaimStrings{clientID},
			Issuer:    i.issuer,
			ExpiresAt: jwtv5.NewNumericDate(expires),
			IssuedAt:  jwtv5.NewNumericDate(time.Now().UTC()),
		},
	}
	return i.sign(claims, expires)
}

func (i *Issuer) sign(claims Claims, expires time.Time) (string, time.Time, error) {
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, claims)
	token.Header["kid"] = i.kid
	signed, err := token.SignedString(i.key)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expires, nil
}

func (i *Issuer) ParseAccess(tokenString string) (*Claims, error) {
	parsed, err := jwtv5.ParseWithClaims(tokenString, &Claims{}, func(t *jwtv5.Token) (any, error) {
		if t.Method != jwtv5.SigningMethodRS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return &i.key.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
