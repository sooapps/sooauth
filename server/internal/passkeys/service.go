package passkeys

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/config"
	"github.com/sooapps/sooauth/server/internal/ephemeral"
	"github.com/sooapps/sooauth/server/internal/store"
)

type Service struct {
	cfg   config.Config
	wa    *webauthn.WebAuthn
	store *store.WebAuthnCredentials
	users *store.Users
	auth  *auth.Service
	audit *store.Audit
	ephem *ephemeral.Store
}

func NewService(cfg config.Config, dbStore *store.WebAuthnCredentials, users *store.Users, authSvc *auth.Service, audit *store.Audit, ephem *ephemeral.Store) (*Service, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: cfg.BrandName,
		RPID:          cfg.WebAuthnRPID,
		RPOrigins:     []string{cfg.WebAuthnOrigin},
	})
	if err != nil {
		return nil, err
	}
	return &Service{cfg: cfg, wa: wa, store: dbStore, users: users, auth: authSvc, audit: audit, ephem: ephem}, nil
}

type sessionPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Raw    []byte    `json:"raw"`
}

func (s *Service) BeginRegistration(ctx context.Context, userID uuid.UUID) (*protocol.CredentialCreation, error) {
	user, err := s.store.LoadUser(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	options, session, err := s.wa.BeginRegistration(user)
	if err != nil {
		return nil, err
	}
	if err := s.saveSession(ctx, regKey(userID), userID, session); err != nil {
		return nil, err
	}
	return options, nil
}

func (s *Service) FinishRegistration(ctx context.Context, userID uuid.UUID, body io.Reader) error {
	user, err := s.store.LoadUser(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}
	session, err := s.loadSession(ctx, regKey(userID), userID)
	if err != nil {
		return err
	}
	parsed, err := protocol.ParseCredentialCreationResponseBody(body)
	if err != nil {
		return err
	}
	credential, err := s.wa.CreateCredential(user, *session, parsed)
	if err != nil {
		return err
	}
	if err := s.store.SaveCredential(ctx, userID, *credential); err != nil {
		return err
	}
	_ = s.ephem.Delete(ctx, regKey(userID))
	uid := userID
	s.audit.Log(ctx, &uid, "passkey_registered", nil, "")
	return nil
}

func (s *Service) BeginLogin(ctx context.Context, email string) (*protocol.CredentialAssertion, string, error) {
	found, err := s.users.FindByEmailSimple(ctx, email)
	if err != nil || found == nil {
		return nil, "", auth.ErrInvalidCredentials
	}
	user, err := s.store.LoadUser(ctx, found.ID)
	if err != nil || user == nil {
		return nil, "", auth.ErrInvalidCredentials
	}
	options, session, err := s.wa.BeginLogin(user)
	if err != nil {
		return nil, "", err
	}
	state := base64.RawURLEncoding.EncodeToString([]byte(found.ID.String()))
	if err := s.saveSession(ctx, loginKey(state), found.ID, session); err != nil {
		return nil, "", err
	}
	return options, state, nil
}

func (s *Service) FinishLogin(ctx context.Context, state string, body io.Reader, ip, userAgent string) (*auth.TokenBundle, error) {
	var saved sessionPayload
	ok, err := s.ephem.Get(ctx, loginKey(state), &saved)
	if err != nil || !ok {
		return nil, auth.ErrInvalidToken
	}
	var session webauthn.SessionData
	if err := json.Unmarshal(saved.Raw, &session); err != nil {
		return nil, err
	}

	user, err := s.store.LoadUser(ctx, saved.UserID)
	if err != nil || user == nil {
		return nil, auth.ErrInvalidCredentials
	}
	parsed, err := protocol.ParseCredentialRequestResponseBody(body)
	if err != nil {
		return nil, err
	}
	if _, err = s.wa.ValidateLogin(user, session, parsed); err != nil {
		return nil, err
	}
	_ = s.ephem.Delete(ctx, loginKey(state))

	found, err := s.users.FindByID(ctx, saved.UserID)
	if err != nil || found == nil {
		return nil, auth.ErrInvalidCredentials
	}
	bundle, err := s.auth.SignInSocial(ctx, found, ip, userAgent)
	if err != nil {
		return nil, err
	}
	uid := found.ID
	s.audit.Log(ctx, &uid, "passkey_signin", nil, ip)
	return bundle, nil
}

func (s *Service) saveSession(ctx context.Context, key string, userID uuid.UUID, session *webauthn.SessionData) error {
	raw, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.ephem.Set(ctx, key, sessionPayload{UserID: userID, Raw: raw}, 5*time.Minute)
}

func (s *Service) loadSession(ctx context.Context, key string, userID uuid.UUID) (*webauthn.SessionData, error) {
	var saved sessionPayload
	ok, err := s.ephem.Get(ctx, key, &saved)
	if err != nil || !ok || saved.UserID != userID {
		return nil, fmt.Errorf("session expired")
	}
	var session webauthn.SessionData
	if err := json.Unmarshal(saved.Raw, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func regKey(userID uuid.UUID) string  { return "passkey:reg:" + userID.String() }
func loginKey(state string) string    { return "passkey:login:" + state }
