package httpserver

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/i18n"
	"github.com/sooapps/sooauth/server/internal/ratelimit"
	"github.com/sooapps/sooauth/server/internal/store"
)

const (
	otpTTL         = 5 * time.Minute
	otpMaxAttempts = 5
)

var (
	errOTPDisabled         = errors.New("otp_disabled")
	errOTPIdentifier       = errors.New("identifier_not_allowed")
	errOTPPhoneUnavailable = errors.New("phone_otp_unavailable")
)

func otpCacheKey(tenantID uuid.UUID, identifier string) string {
	return fmt.Sprintf("otp:%s:%s", tenantID.String(), strings.ToLower(strings.TrimSpace(identifier)))
}

type otpPayload struct {
	Code       string    `json:"code"`
	Identifier string    `json:"identifier"`
	ClientID   string    `json:"client_id"`
	Attempts   int       `json:"attempts"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (s *Server) limiter() *ratelimit.Limiter {
	return ratelimit.New(s.rdb)
}

// generateOTP returns a uniformly random 6-digit code.
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// otpClient resolves the project for a public OTP request.
func (s *Server) otpClient(ctx context.Context, clientID string) (*store.OAuthClient, *store.Tenant, error) {
	client, err := s.oauthClients.FindByClientID(ctx, clientID)
	if err != nil || client == nil || client.TenantID == nil {
		return nil, nil, store.ErrInvalidClient
	}
	tenant, err := s.tenants.FindByID(ctx, *client.TenantID)
	if err != nil || tenant == nil {
		return nil, nil, store.ErrInvalidClient
	}
	return client, tenant, nil
}

// issueOTP stores a fresh code for key and returns it.
func (s *Server) issueOTP(ctx context.Context, key, identifier, clientID string) (string, error) {
	if s.ephemeral == nil {
		return "", errors.New("ephemeral store not configured")
	}
	code, err := generateOTP()
	if err != nil {
		return "", err
	}
	payload := otpPayload{
		Code:       code,
		Identifier: identifier,
		ClientID:   clientID,
		ExpiresAt:  time.Now().UTC().Add(otpTTL),
	}
	if err := s.ephemeral.Set(ctx, key, payload, otpTTL); err != nil {
		return "", err
	}
	return code, nil
}

type otpCheck int

const (
	otpValid otpCheck = iota
	otpInvalid
	otpLocked
)

// checkOTP compares code against the stored one without consuming it. Wrong
// guesses count towards otpMaxAttempts, after which the code is discarded.
func (s *Server) checkOTP(ctx context.Context, key, code string) otpCheck {
	if s.ephemeral == nil {
		return otpInvalid
	}
	var payload otpPayload
	ok, err := s.ephemeral.Get(ctx, key, &payload)
	if err != nil || !ok || payload.Code == "" {
		return otpInvalid
	}
	if subtle.ConstantTimeCompare([]byte(payload.Code), []byte(strings.TrimSpace(code))) == 1 {
		return otpValid
	}
	payload.Attempts++
	if payload.Attempts >= otpMaxAttempts {
		_ = s.ephemeral.Delete(ctx, key)
		return otpLocked
	}
	ttl := time.Until(payload.ExpiresAt)
	if ttl <= 0 {
		_ = s.ephemeral.Delete(ctx, key)
		return otpInvalid
	}
	_ = s.ephemeral.Set(ctx, key, payload, ttl)
	return otpInvalid
}

// deliverOTP sends a code to an email address or, when the project enabled it,
// hands a phone code to the project's otp.sent webhook.
func (s *Server) deliverOTP(r *http.Request, tenant *store.Tenant, client *store.OAuthClient, kind, destination, code string, mfa bool) error {
	ctx := r.Context()
	switch kind {
	case "email":
		theme, _ := s.theme.GetByTenant(ctx, tenant.ID)
		brand := appDisplayBrand(tenant, theme, client)
		lang := i18n.Resolve(r, tenant.DefaultLocale)
		if err := s.auth.SendSignInCode(ctx, tenant.ID, destination, brand, code, lang, mfa); err != nil {
			return err
		}
	case "phone":
		if !tenant.AuthConfig.PhoneOTPViaWebhook || s.webhooks == nil {
			return errOTPPhoneUnavailable
		}
	default:
		return errOTPIdentifier
	}
	if s.cfg.Env == "development" {
		slog.Info("otp issued (development only)", "tenant", tenant.ID, "to", destination, "code", code)
	}
	if s.webhooks != nil {
		data := map[string]any{"identifier": destination, "channel": kind, "ip": clientIP(r), "purpose": "sign_in"}
		if mfa {
			data["purpose"] = "mfa"
		}
		if kind == "phone" {
			// The project sends the SMS itself, so it needs the code.
			data["code"] = code
			data["expires_in"] = int(otpTTL.Seconds())
		}
		s.webhooks.Emit(ctx, &tenant.ID, "otp.sent", nil, data)
	}
	return nil
}

// otpDestination maps a sign-in identifier to where the code is delivered. A
// username has no inbox, so the code goes to that user's email or phone.
func (s *Server) otpDestination(ctx context.Context, tenant *store.Tenant, identifier string) (kind, destination string, err error) {
	kind = store.IdentifierKind(identifier)
	if !tenant.AuthConfig.AllowsIdentifier(kind) {
		return "", "", errOTPIdentifier
	}
	if kind != "username" {
		return kind, identifier, nil
	}
	user, _ := s.users.FindAppUserByIdentifierSimple(ctx, tenant.ID, identifier)
	if user == nil {
		return "", "", nil
	}
	if user.Email != "" {
		return "email", user.Email, nil
	}
	if user.Phone != nil && *user.Phone != "" {
		return "phone", *user.Phone, nil
	}
	return "", "", nil
}

func (s *Server) handleSendOTP(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	var body struct {
		ClientID   string `json:"client_id"`
		Identifier string `json:"identifier"`
		Channel    string `json:"channel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	clientID := strings.TrimSpace(body.ClientID)
	identifier := strings.TrimSpace(body.Identifier)
	if clientID == "" || identifier == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "client_id_and_identifier_required"})
		return
	}
	client, tenant, err := s.otpClient(r.Context(), clientID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "client_not_found"})
		return
	}
	if !tenant.AuthConfig.AllowsOTP() {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": errOTPDisabled.Error(), "message": "Sign-in codes are not enabled for this project."})
		return
	}

	limit := s.limiter()
	if ok, _ := limit.Allow(r.Context(), "otp_send_ip", clientIP(r), 20, time.Minute); !ok {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if ok, _ := limit.Allow(r.Context(), "otp_send_id", tenant.ID.String()+":"+strings.ToLower(identifier), 5, 10*time.Minute); !ok {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}

	kind, destination, err := s.otpDestination(r.Context(), tenant, identifier)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if kind == "phone" && !tenant.AuthConfig.PhoneOTPViaWebhook {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errOTPPhoneUnavailable.Error(), "message": "Phone codes are not set up for this project."})
		return
	}
	sent := map[string]any{"message": "If the account can receive codes, a code was sent.", "expires_in": int(otpTTL.Seconds())}
	if destination == "" {
		// Unknown username: answer the same way so usernames can't be probed.
		writeJSON(w, http.StatusOK, sent)
		return
	}

	code, err := s.issueOTP(r.Context(), otpCacheKey(tenant.ID, identifier), identifier, clientID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_store_otp"})
		return
	}
	if err := s.deliverOTP(r, tenant, client, kind, destination, code, false); err != nil {
		_ = s.ephemeral.Delete(r.Context(), otpCacheKey(tenant.ID, identifier))
		if errors.Is(err, errOTPPhoneUnavailable) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		slog.Error("otp delivery", "tenant", tenant.ID, "err", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "otp_delivery_failed"})
		return
	}
	writeJSON(w, http.StatusOK, sent)
}

func (s *Server) handleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	var body struct {
		ClientID   string         `json:"client_id"`
		Identifier string         `json:"identifier"`
		Code       string         `json:"code"`
		Metadata   map[string]any `json:"metadata"`
		RememberMe *bool          `json:"remember_me"`
		Hosted     bool           `json:"hosted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	clientID := strings.TrimSpace(body.ClientID)
	identifier := strings.TrimSpace(body.Identifier)
	code := strings.TrimSpace(body.Code)
	if clientID == "" || identifier == "" || code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_parameters"})
		return
	}
	_, tenant, err := s.otpClient(r.Context(), clientID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "client_not_found"})
		return
	}
	if !tenant.AuthConfig.AllowsOTP() {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": errOTPDisabled.Error()})
		return
	}
	kind := store.IdentifierKind(identifier)
	if !tenant.AuthConfig.AllowsIdentifier(kind) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errOTPIdentifier.Error()})
		return
	}

	key := otpCacheKey(tenant.ID, identifier)
	switch s.checkOTP(r.Context(), key, code) {
	case otpLocked:
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too_many_attempts", "message": "Too many wrong codes. Request a new code."})
		return
	case otpInvalid:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_code"})
		return
	}

	user, _ := s.users.FindAppUserByIdentifierSimple(r.Context(), tenant.ID, identifier)
	if user == nil {
		// First sign-in with a code creates the account. Required registration
		// fields must be supplied; the code stays valid so the client can retry
		// with metadata.
		if missing := missingRequiredFields(tenant.RegistrationSchema, body.Metadata); len(missing) > 0 {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "registration_required", "fields": missing})
			return
		}
		var email, username, phone string
		switch kind {
		case "email":
			email = strings.ToLower(identifier)
		case "phone":
			phone = identifier
		default:
			// A username cannot receive a code, so it must already exist.
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_code"})
			return
		}
		user, err = s.users.CreateAppUserFlexible(r.Context(), tenant.ID, email, username, phone, "", body.Metadata)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_create_user"})
			return
		}
		_ = s.tenantUsers.LinkWithMethod(r.Context(), tenant.ID, user.ID, "otp")
		if s.webhooks != nil {
			s.webhooks.Emit(r.Context(), &tenant.ID, "user.created", user, map[string]any{"method": "otp", "ip": clientIP(r)})
		}
	}
	_ = s.ephemeral.Delete(r.Context(), key)

	if kind == "phone" {
		_ = s.users.MarkPhoneVerified(r.Context(), user.ID)
	} else if kind == "email" {
		_ = s.users.MarkEmailVerified(r.Context(), user.ID)
	}

	rememberMe := true
	if body.RememberMe != nil {
		rememberMe = *body.RememberMe
	}
	bundle, err := s.auth.CompleteSignIn(r.Context(), user, clientIP(r), r.UserAgent(), "otp_signin", rememberMe)
	if errors.Is(err, auth.ErrUserDisabled) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "user_disabled"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "signin_failed"})
		return
	}
	if body.Hosted {
		setAppSessionCookie(w, s.cfg.CookieSecure, s.cfg.CookieDomain, bundle)
	}
	writeJSON(w, http.StatusOK, appTokenResponse(bundle))
}

func missingRequiredFields(schema []store.RegistrationField, metadata map[string]any) []string {
	var missing []string
	for _, field := range schema {
		if !field.Required {
			continue
		}
		val, ok := metadata[field.ID]
		if !ok || val == nil || fmt.Sprintf("%v", val) == "" {
			missing = append(missing, field.ID)
		}
	}
	return missing
}

// appTokenResponse is the JSON body returned to custom UIs and the widget after
// a project sign-in.
func appTokenResponse(bundle *auth.TokenBundle) map[string]any {
	return map[string]any{
		"message":       "Signed in successfully.",
		"id":            bundle.User.ID,
		"email":         bundle.User.Email,
		"access_token":  bundle.AccessToken,
		"refresh_token": bundle.RefreshToken,
		"expires_in":    bundle.ExpiresIn,
		"token_type":    "Bearer",
	}
}
