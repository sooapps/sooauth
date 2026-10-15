package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/crypto/token"
	"github.com/sooapps/sooauth/server/internal/mfa"
	"github.com/sooapps/sooauth/server/internal/store"
)

// Two-step verification for project end users. After the password matched,
// appMFAGate decides whether a second step is needed and returns
// *auth.MFARequiredError with a challenge. The client finishes the sign-in
// with POST /auth/mfa/verify {challenge, code}.

func mfaEmailKey(challenge string) string {
	return "mfa_email:" + token.Hash(challenge)
}

func (s *Server) appMFAGate(r *http.Request, tenant *store.Tenant, client *store.OAuthClient) func(context.Context, *store.User) error {
	cfg := tenant.AuthConfig
	if !cfg.MFAEnabled() || s.mfa == nil {
		return nil
	}
	return func(ctx context.Context, user *store.User) error {
		enrolled := false
		if cfg.AllowsMFAMethod(store.MFAMethodTOTP) {
			enrolled, _ = s.mfa.Enabled(ctx, user.ID)
		}
		canEmail := cfg.AllowsMFAMethod(store.MFAMethodEmailOTP) && user.Email != ""

		var method string
		var methods []string
		switch {
		case enrolled:
			method = store.MFAMethodTOTP
			methods = append(methods, store.MFAMethodTOTP)
			if canEmail {
				methods = append(methods, store.MFAMethodEmailOTP)
			}
		case cfg.MFAPolicy != store.MFAPolicyRequired:
			// Optional 2FA only applies to users who set up an authenticator.
			return nil
		case canEmail:
			method = store.MFAMethodEmailOTP
			methods = []string{store.MFAMethodEmailOTP}
		case cfg.AllowsMFAMethod(store.MFAMethodTOTP):
			method = "totp_setup"
			methods = []string{"totp_setup"}
		default:
			return nil
		}

		challenge, err := s.mfa.Challenge(ctx, user.ID)
		if err != nil {
			return err
		}
		if method == store.MFAMethodEmailOTP {
			if err := s.sendMFAEmailCode(r, tenant, client, user, challenge); err != nil {
				return err
			}
		}
		return &auth.MFARequiredError{Challenge: challenge, Method: method, Methods: methods}
	}
}

func (s *Server) sendMFAEmailCode(r *http.Request, tenant *store.Tenant, client *store.OAuthClient, user *store.User, challenge string) error {
	code, err := s.issueOTP(r.Context(), mfaEmailKey(challenge), user.Email, "")
	if err != nil {
		return err
	}
	return s.deliverOTP(r, tenant, client, "email", user.Email, code, true)
}

// appChallengeUser loads the end user and project behind an open challenge.
func (s *Server) appChallengeUser(ctx context.Context, challenge string) (*store.User, *store.Tenant, *store.OAuthClient, error) {
	if s.mfa == nil || challenge == "" {
		return nil, nil, nil, mfa.ErrInvalidCode
	}
	userID, err := s.mfa.ChallengeOwner(ctx, challenge)
	if err != nil {
		return nil, nil, nil, err
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, nil, nil, mfa.ErrInvalidCode
	}
	if user.TenantID == nil {
		return user, nil, nil, nil
	}
	tenant, err := s.tenants.FindByID(ctx, *user.TenantID)
	if err != nil || tenant == nil {
		return nil, nil, nil, mfa.ErrInvalidCode
	}
	client, _ := s.oauthClients.FindPrimaryByTenant(ctx, tenant.ID)
	return user, tenant, client, nil
}

type appMFAVerifyBody struct {
	Challenge  string `json:"challenge"`
	Code       string `json:"code"`
	Method     string `json:"method"`
	RememberMe *bool  `json:"remember_me"`
	Hosted     bool   `json:"hosted"`
}

func (s *Server) handleAppMFAVerify(w http.ResponseWriter, r *http.Request, body appMFAVerifyBody, user *store.User, tenant *store.Tenant) {
	ctx := r.Context()
	if ok, _ := s.limiter().Allow(ctx, "mfa_verify", token.Hash(body.Challenge), 10, 5*time.Minute); !ok {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}

	var backupCodes []string
	verified := false
	enrolled, _ := s.mfa.Enabled(ctx, user.ID)
	switch {
	case body.Method == store.MFAMethodEmailOTP || (!enrolled && body.Method == ""):
		if pending, _ := s.mfa.HasPending(ctx, user.ID); pending && body.Method == "" {
			codes, err := s.mfa.ConfirmFor(ctx, user.ID, body.Code)
			verified, backupCodes = err == nil, codes
			break
		}
		switch s.checkOTP(ctx, mfaEmailKey(body.Challenge), body.Code) {
		case otpValid:
			verified = true
			_ = s.ephemeral.Delete(ctx, mfaEmailKey(body.Challenge))
		case otpLocked:
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too_many_attempts"})
			return
		}
	case body.Method == "totp_setup":
		codes, err := s.mfa.ConfirmFor(ctx, user.ID, body.Code)
		verified, backupCodes = err == nil, codes
	default:
		verified = tenant.AuthConfig.AllowsMFAMethod(store.MFAMethodTOTP) && s.mfa.CheckCode(ctx, user.ID, body.Code)
	}
	if !verified {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_mfa_code"})
		return
	}
	if err := s.mfa.UseChallenge(ctx, body.Challenge, user.ID); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_mfa_code"})
		return
	}

	rememberMe := true
	if body.RememberMe != nil {
		rememberMe = *body.RememberMe
	}
	bundle, err := s.auth.CompleteSignIn(ctx, user, clientIP(r), r.UserAgent(), "mfa_signin", rememberMe)
	if errors.Is(err, auth.ErrUserDisabled) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "user_disabled"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "mfa_verification_failed"})
		return
	}
	if body.Hosted {
		setAppSessionCookie(w, s.cfg.CookieSecure, s.cfg.CookieDomain, bundle)
	}
	resp := appTokenResponse(bundle)
	if len(backupCodes) > 0 {
		resp["backup_codes"] = backupCodes
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleAppMFASetup starts authenticator enrollment during a "totp_setup"
// challenge (2FA required but the user has no authenticator yet).
func (s *Server) handleAppMFASetup(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}
	var body struct {
		Challenge string `json:"challenge"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	user, tenant, client, err := s.appChallengeUser(r.Context(), body.Challenge)
	if err != nil || tenant == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_challenge"})
		return
	}
	if !tenant.AuthConfig.AllowsMFAMethod(store.MFAMethodTOTP) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "totp_disabled"})
		return
	}
	s.writeTOTPSetup(w, r, user, tenant, client)
}

func (s *Server) writeTOTPSetup(w http.ResponseWriter, r *http.Request, user *store.User, tenant *store.Tenant, client *store.OAuthClient) {
	theme, _ := s.theme.GetByTenant(r.Context(), tenant.ID)
	issuer := appDisplayBrand(tenant, theme, client)
	account := user.Email
	if account == "" && user.Username != nil {
		account = *user.Username
	}
	if account == "" && user.Phone != nil {
		account = *user.Phone
	}
	setup, err := s.mfa.StartFor(r.Context(), user.ID, issuer, account)
	if errors.Is(err, mfa.ErrAlreadyEnable) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "mfa_already_enabled"})
		return
	}
	if err != nil {
		slog.Error("totp setup", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "mfa_setup_failed"})
		return
	}
	resp := map[string]any{"secret": setup.Secret, "otpauth_url": setup.URL}
	if png, err := qrcode.Encode(setup.URL, qrcode.Medium, 220); err == nil {
		resp["qr_code"] = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleAppMFAResend(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}
	var body struct {
		Challenge string `json:"challenge"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	user, tenant, client, err := s.appChallengeUser(r.Context(), body.Challenge)
	if err != nil || tenant == nil || user.Email == "" || !tenant.AuthConfig.AllowsMFAMethod(store.MFAMethodEmailOTP) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_challenge"})
		return
	}
	if ok, _ := s.limiter().Allow(r.Context(), "mfa_resend", token.Hash(body.Challenge), 3, 5*time.Minute); !ok {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if err := s.sendMFAEmailCode(r, tenant, client, user, body.Challenge); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "otp_delivery_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "Code sent.", "method": store.MFAMethodEmailOTP})
}

// Self-service authenticator management for a signed-in end user
// (Authorization: Bearer <access_token>).

func (s *Server) requireAppBearer(w http.ResponseWriter, r *http.Request) (*store.User, *store.Tenant, *store.OAuthClient, bool) {
	bearer := bearerToken(r)
	user, err := s.auth.UserFromAccess(r.Context(), bearer)
	if bearer == "" || err != nil || user == nil || user.TenantID == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return nil, nil, nil, false
	}
	tenant, err := s.tenants.FindByID(r.Context(), *user.TenantID)
	if err != nil || tenant == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return nil, nil, nil, false
	}
	client, _ := s.oauthClients.FindPrimaryByTenant(r.Context(), tenant.ID)
	return user, tenant, client, true
}

func (s *Server) handleAppMeMFA(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}
	user, tenant, _, ok := s.requireAppBearer(w, r)
	if !ok {
		return
	}
	enabled, _ := s.mfa.Enabled(r.Context(), user.ID)
	writeJSON(w, http.StatusOK, map[string]any{
		"policy":       tenant.AuthConfig.MFAPolicy,
		"methods":      tenant.AuthConfig.MFAMethods,
		"totp_enabled": enabled,
	})
}

func (s *Server) handleAppMeTOTPStart(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}
	user, tenant, client, ok := s.requireAppBearer(w, r)
	if !ok {
		return
	}
	if !tenant.AuthConfig.MFAEnabled() || !tenant.AuthConfig.AllowsMFAMethod(store.MFAMethodTOTP) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "totp_disabled"})
		return
	}
	s.writeTOTPSetup(w, r, user, tenant, client)
}

func (s *Server) handleAppMeTOTPConfirm(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}
	user, _, _, ok := s.requireAppBearer(w, r)
	if !ok {
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	codes, err := s.mfa.ConfirmFor(r.Context(), user.ID, body.Code)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_mfa_code"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "backup_codes": codes})
}

func (s *Server) handleAppMeTOTPDisable(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}
	user, tenant, _, ok := s.requireAppBearer(w, r)
	if !ok {
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	if tenant.AuthConfig.MFAPolicy == store.MFAPolicyRequired && !tenant.AuthConfig.AllowsMFAMethod(store.MFAMethodEmailOTP) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "mfa_required_by_project"})
		return
	}
	if !s.mfa.CheckCode(r.Context(), user.ID, body.Code) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_mfa_code"})
		return
	}
	if err := s.mfa.Reset(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "mfa_disable_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": false})
}

// handleDashboardResetUserMFA lets the project owner remove a user's
// authenticator (lost phone). The user enrolls again on the next sign-in.
func (s *Server) handleDashboardResetUserMFA(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok || !accountCSRFValid(w, r) {
		return
	}
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_id"})
		return
	}
	user, err := s.users.FindByID(r.Context(), userID)
	if err != nil || user == nil || user.TenantID == nil || *user.TenantID != dash.tenant.ID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	if s.mfa == nil || s.mfa.Reset(r.Context(), userID) != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "mfa_reset_failed"})
		return
	}
	uid := dash.user.ID
	s.audit.Log(r.Context(), &uid, "app_user_mfa_reset", map[string]any{"user_id": userID.String(), "tenant_id": dash.tenant.ID.String()}, clientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "mfa_reset"})
}

func mfaRequiredBody(e *auth.MFARequiredError) map[string]any {
	body := map[string]any{"error": "mfa_required", "challenge": e.Challenge}
	if e.Method != "" {
		body["method"] = e.Method
		body["methods"] = e.Methods
	}
	return body
}
