package mfa

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/sooapps/sooauth/server/internal/crypto/token"
	"github.com/sooapps/sooauth/server/internal/store"
)

var (
	ErrInvalidCode    = errors.New("invalid mfa code")
	ErrAlreadyEnable  = errors.New("mfa already enabled")
	ErrNoPendingSetup = errors.New("no pending mfa setup")
)

type Setup struct {
	Secret string
	URL    string
}

type Service struct{ records *store.MFA }

func NewService(records *store.MFA) *Service { return &Service{records: records} }

func (s *Service) Start(ctx context.Context, user *store.User) (*Setup, error) {
	return s.StartFor(ctx, user.ID, "sooauth", user.Email)
}

// StartFor creates a pending TOTP secret for userID. issuer is the name shown in
// the authenticator app (the project's brand for end users).
func (s *Service) StartFor(ctx context.Context, userID uuid.UUID, issuer, account string) (*Setup, error) {
	if _, enabled, err := s.records.Secret(ctx, userID); err != nil {
		return nil, err
	} else if enabled {
		return nil, ErrAlreadyEnable
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: issuer, AccountName: account, Period: 30, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		return nil, err
	}
	if err := s.records.PendingSecret(ctx, userID, key.Secret()); err != nil {
		return nil, err
	}
	return &Setup{Secret: key.Secret(), URL: key.URL()}, nil
}

// PendingURL rebuilds the local setup URI from the encrypted pending secret.
// The URI is only returned to the authenticated account and never sent to a
// third-party QR service.
func (s *Service) PendingURL(ctx context.Context, user *store.User) (string, error) {
	secret, enabled, err := s.records.Secret(ctx, user.ID)
	if err != nil {
		return "", err
	}
	if secret == "" || enabled {
		return "", ErrNoPendingSetup
	}
	values := url.Values{}
	values.Set("secret", secret)
	values.Set("issuer", "sooauth")
	return fmt.Sprintf("otpauth://totp/%s:%s?%s", url.PathEscape("sooauth"), url.PathEscape(user.Email), values.Encode()), nil
}

func (s *Service) Confirm(ctx context.Context, user *store.User, code string) ([]string, error) {
	return s.ConfirmFor(ctx, user.ID, code)
}

// ConfirmFor enables a pending TOTP secret once code matches it and returns
// fresh backup codes.
func (s *Service) ConfirmFor(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	secret, enabled, err := s.records.Secret(ctx, userID)
	if err != nil || secret == "" || enabled || !validCode(secret, code) {
		return nil, ErrInvalidCode
	}
	codes, hashes, err := backupCodes()
	if err != nil {
		return nil, err
	}
	if err := s.records.ReplaceBackupCodes(ctx, userID, hashes); err != nil {
		return nil, err
	}
	if err := s.records.Enable(ctx, userID); err != nil {
		return nil, err
	}
	return codes, nil
}

func (s *Service) Disable(ctx context.Context, user *store.User, code string) error {
	secret, enabled, err := s.records.Secret(ctx, user.ID)
	if err != nil || !enabled || !validCode(secret, code) {
		return ErrInvalidCode
	}
	return s.records.Disable(ctx, user.ID)
}

func (s *Service) Enabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	_, enabled, err := s.records.Secret(ctx, userID)
	return enabled, err
}

func (s *Service) Challenge(ctx context.Context, userID uuid.UUID) (string, error) {
	plain, hash, err := token.Generate()
	if err != nil {
		return "", err
	}
	if err := s.records.CreateChallenge(ctx, userID, hash, time.Now().UTC().Add(5*time.Minute)); err != nil {
		return "", err
	}
	return plain, nil
}

func (s *Service) VerifyChallenge(ctx context.Context, challenge, code string) (uuid.UUID, error) {
	challengeHash := token.Hash(challenge)
	userID, err := s.records.ChallengeUser(ctx, challengeHash)
	if err != nil {
		return uuid.Nil, ErrInvalidCode
	}
	secret, enabled, err := s.records.Secret(ctx, userID)
	if err != nil || !enabled {
		return uuid.Nil, ErrInvalidCode
	}
	if validCode(secret, code) {
		return s.consumeVerifiedChallenge(ctx, challengeHash, userID)
	}
	used, err := s.records.ConsumeBackupCode(ctx, userID, token.Hash(strings.ToUpper(strings.TrimSpace(code))))
	if err != nil || !used {
		return uuid.Nil, ErrInvalidCode
	}
	return s.consumeVerifiedChallenge(ctx, challengeHash, userID)
}

func (s *Service) consumeVerifiedChallenge(ctx context.Context, hash string, userID uuid.UUID) (uuid.UUID, error) {
	consumed, err := s.records.ConsumeChallenge(ctx, hash)
	if err != nil || consumed != userID {
		return uuid.Nil, ErrInvalidCode
	}
	return userID, nil
}

func validCode(secret, code string) bool {
	return totp.Validate(strings.TrimSpace(code), secret)
}

func backupCodes() ([]string, []string, error) {
	codes := make([]string, 10)
	hashes := make([]string, 10)
	for i := range codes {
		raw := make([]byte, 5)
		if _, err := rand.Read(raw); err != nil {
			return nil, nil, err
		}
		codes[i] = strings.ToUpper(hex.EncodeToString(raw))
		hashes[i] = token.Hash(codes[i])
	}
	return codes, hashes, nil
}

// HasPending reports whether userID started TOTP setup without confirming it.
func (s *Service) HasPending(ctx context.Context, userID uuid.UUID) (bool, error) {
	secret, enabled, err := s.records.Secret(ctx, userID)
	return secret != "" && !enabled, err
}

// CheckCode accepts a current TOTP code or an unused backup code for an
// enrolled user. A backup code is consumed.
func (s *Service) CheckCode(ctx context.Context, userID uuid.UUID, code string) bool {
	secret, enabled, err := s.records.Secret(ctx, userID)
	if err != nil || !enabled {
		return false
	}
	if validCode(secret, code) {
		return true
	}
	used, err := s.records.ConsumeBackupCode(ctx, userID, token.Hash(strings.ToUpper(strings.TrimSpace(code))))
	return err == nil && used
}

// ChallengeOwner returns the user behind an open challenge without using it.
func (s *Service) ChallengeOwner(ctx context.Context, challenge string) (uuid.UUID, error) {
	userID, err := s.records.ChallengeUser(ctx, token.Hash(challenge))
	if err != nil {
		return uuid.Nil, ErrInvalidCode
	}
	return userID, nil
}

// UseChallenge marks a challenge as used after the second factor succeeded.
func (s *Service) UseChallenge(ctx context.Context, challenge string, userID uuid.UUID) error {
	_, err := s.consumeVerifiedChallenge(ctx, token.Hash(challenge), userID)
	return err
}

// Reset removes a user's TOTP secret and backup codes (admin reset).
func (s *Service) Reset(ctx context.Context, userID uuid.UUID) error {
	return s.records.Disable(ctx, userID)
}

// EnabledAmong returns which of userIDs have TOTP enabled.
func (s *Service) EnabledAmong(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	return s.records.EnabledAmong(ctx, userIDs)
}
