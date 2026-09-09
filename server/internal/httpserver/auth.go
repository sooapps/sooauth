package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/accounts"
	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/emailverify"
	"github.com/sooapps/sooauth/server/internal/mfa"
	"github.com/sooapps/sooauth/server/internal/passwordpolicy"
	"github.com/sooapps/sooauth/server/internal/store"
)

const (
	sessionCookie = "sooauth_session"
	refreshCookie = "sooauth_refresh"
	csrfCookie    = "sooauth_csrf"
	csrfHeader    = "X-CSRF-Token"
)

func (s *Server) handleSignUp(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		ClientID string `json:"client_id"`
		ReturnTo string `json:"return_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	var err error
	if strings.TrimSpace(body.ClientID) != "" {
		err = s.signUpAppUser(r.Context(), body.ClientID, body.Email, body.Password, body.ReturnTo, clientIP(r))
	} else {
		err = s.auth.SignUp(r.Context(), body.Email, body.Password, clientIP(r))
	}
	if err == auth.ErrRateLimited {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if errors.Is(err, auth.ErrEmailDelivery) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error":   "email_delivery_failed",
			"message": "Verification email could not be sent. Check SMTP settings, then try signing up again with the same email.",
		})
		return
	}
	if errors.Is(err, store.ErrEmailTaken) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "email_already_registered",
			"message": "This email is already registered. Sign in instead.",
		})
		return
	}
	if errors.Is(err, passwordpolicy.ErrWeak) {
		msg := "Password does not meet this project's requirements."
		if cid := strings.TrimSpace(body.ClientID); cid != "" {
			if client, _ := s.oauthClients.FindByClientID(r.Context(), cid); client != nil && client.TenantID != nil {
				if tenant, _ := s.tenants.FindByID(r.Context(), *client.TenantID); tenant != nil {
					msg = tenant.PasswordPolicy().Describe()
				}
			}
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "weak_password",
			"message": msg,
		})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "signup_failed"})
		return
	}

	if body.ClientID == "" {
		if user, _ := s.users.FindPlatformByEmailSimple(r.Context(), body.Email); user != nil {
			_, _, _, _ = accounts.ProvisionOwner(r.Context(), s.accounts, s.tenants, s.oauthClients, s.theme, user.ID, user.Email)
		}
	}

	msg := "Check your email to verify your account."
	if strings.TrimSpace(body.ClientID) != "" {
		client, _ := s.oauthClients.FindByClientID(r.Context(), body.ClientID)
		if client != nil && client.TenantID != nil {
			if tenant, _ := s.tenants.FindByID(r.Context(), *client.TenantID); tenant != nil && !tenant.EmailVerifyRequired {
				msg = "Account created. You can sign in now."
			}
		}
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": msg,
	})
}

func (s *Server) handleResendVerification(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	var body struct {
		Email    string `json:"email"`
		ClientID string `json:"client_id"`
		ReturnTo string `json:"return_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	if strings.TrimSpace(body.ClientID) != "" {
		err := s.resendAppVerification(r, body.ClientID, body.Email, body.ReturnTo)
		if err == auth.ErrRateLimited {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
			return
		}
		if errors.Is(err, auth.ErrEmailDelivery) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"error":   "email_delivery_failed",
				"message": "Verification email could not be sent. SMTP may be misconfigured on the server.",
			})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "If your account is unverified, we sent a new verification message.",
		})
		return
	}

	err := s.auth.ResendVerification(r.Context(), body.Email, clientIP(r))
	if err == auth.ErrRateLimited {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if errors.Is(err, auth.ErrEmailDelivery) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error":   "email_delivery_failed",
			"message": "Verification email could not be sent. SMTP may be misconfigured on the server.",
		})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request_failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "If your account is unverified, we sent a new verification link.",
	})
}

func (s *Server) handleSignIn(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		ClientID string `json:"client_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	var bundle *auth.TokenBundle
	var err error
	if strings.TrimSpace(body.ClientID) != "" {
		bundle, err = s.signInAppUser(r.Context(), body.ClientID, body.Email, body.Password, clientIP(r), r.UserAgent())
	} else {
		bundle, err = s.auth.SignIn(r.Context(), body.Email, body.Password, clientIP(r), r.UserAgent())
	}
	if err == auth.ErrRateLimited {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if err == auth.ErrEmailNotVerified {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "email_not_verified"})
		return
	}
	if err == auth.ErrUserDisabled {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "user_disabled"})
		return
	}
	if err == auth.ErrInvalidCredentials {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_credentials"})
		return
	}
	var mfaErr *auth.MFARequiredError
	if errors.As(err, &mfaErr) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error":     "mfa_required",
			"challenge": mfaErr.Challenge,
		})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "signin_failed"})
		return
	}

	if body.ClientID == "" {
		setAuthCookies(w, s.cfg.CookieSecure, s.cfg.CookieDomain, bundle)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":            bundle.User.ID,
		"email":         bundle.User.Email,
		"access_token":  bundle.AccessToken,
		"refresh_token": bundle.RefreshToken,
		"expires_in":    bundle.ExpiresIn,
		"csrf_token":    bundle.CSRFToken,
		"token_type":    "Bearer",
	})
}

func (s *Server) handleMFAStart(w http.ResponseWriter, r *http.Request) {
	user, ok := s.accountUser(w, r)
	if !ok {
		return
	}
	setup, err := s.mfa.Start(r.Context(), user)
	if errors.Is(err, mfa.ErrAlreadyEnable) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "mfa_already_enabled"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "mfa_setup_failed"})
		return
	}
	writeJSON(w, http.StatusOK, setup)
}

func (s *Server) handleMFAConfirm(w http.ResponseWriter, r *http.Request) {
	user, ok := s.accountUser(w, r)
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
	codes, err := s.mfa.Confirm(r.Context(), user, body.Code)
	if errors.Is(err, mfa.ErrInvalidCode) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_mfa_code"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "mfa_confirm_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "backup_codes": codes})
}

func (s *Server) handleMFADisable(w http.ResponseWriter, r *http.Request) {
	user, ok := s.accountUser(w, r)
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
	if err := s.mfa.Disable(r.Context(), user, body.Code); errors.Is(err, mfa.ErrInvalidCode) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_mfa_code"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "mfa_disable_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": false})
}

func (s *Server) handleMFAVerify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Challenge string `json:"challenge"`
		Code      string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	bundle, err := s.auth.VerifyMFA(r.Context(), body.Challenge, body.Code, clientIP(r), r.UserAgent())
	if errors.Is(err, mfa.ErrInvalidCode) || errors.Is(err, auth.ErrInvalidToken) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_mfa_code"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "mfa_verification_failed"})
		return
	}
	setAuthCookies(w, s.cfg.CookieSecure, s.cfg.CookieDomain, bundle)
	writeJSON(w, http.StatusOK, map[string]any{
		"id": bundle.User.ID, "email": bundle.User.Email, "access_token": bundle.AccessToken,
		"refresh_token": bundle.RefreshToken, "expires_in": bundle.ExpiresIn, "csrf_token": bundle.CSRFToken, "token_type": "Bearer",
	})
}

func (s *Server) requirePlatformUser(w http.ResponseWriter, r *http.Request) (*store.User, bool) {
	user, err := s.currentUser(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return nil, false
	}
	platform, err := s.users.IsPlatformOwner(r.Context(), user.ID)
	if err != nil || !platform {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "platform_access_denied"})
		return nil, false
	}
	return user, true
}

func (s *Server) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	plain := r.URL.Query().Get("token")
	if plain == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token_required"})
		return
	}
	already, err := s.auth.VerifyEmail(r.Context(), plain, clientIP(r))
	if err == auth.ErrRateLimited {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_token"})
		return
	}
	msg := "Email verified. You can sign in."
	if already {
		msg = "Email already verified. You can sign in."
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": msg})
}

func (s *Server) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	var body struct {
		Email    string `json:"email"`
		ClientID string `json:"client_id"`
		TenantID string `json:"tenant_id"`
		ReturnTo string `json:"return_to"`
		Delivery string `json:"delivery"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	clientID := strings.TrimSpace(body.ClientID)
	var tenantID *uuid.UUID
	var returnTo string
	brandName := s.cfg.BrandName
	appScoped := false
	delivery := strings.ToLower(strings.TrimSpace(body.Delivery))

	if clientID != "" {
		client, err := s.oauthClients.FindByClientID(r.Context(), clientID)
		if err == nil && client != nil && client.TenantID != nil {
			tenantID = client.TenantID
			appScoped = true
			if tenant, err := s.tenants.FindByID(r.Context(), *client.TenantID); err == nil && tenant != nil {
				theme, _ := s.theme.GetByTenant(r.Context(), *client.TenantID)
				brandName = appDisplayBrand(tenant, theme, client)
				if delivery == "" {
					if tenant.VerifyDelivery() == emailverify.DeliveryCode {
						delivery = "code"
					} else {
						delivery = "link"
					}
				}
			}
			returnTo = resolveAppReturnTo(body.ReturnTo, s.cfg.AppURL, client)
		}
	} else if body.TenantID != "" {
		if parsed, err := uuid.Parse(body.TenantID); err == nil {
			tenantID = &parsed
			appScoped = true
			if tenant, err := s.tenants.FindByID(r.Context(), parsed); err == nil && tenant != nil {
				brandName = tenant.Name
				if delivery == "" {
					if tenant.VerifyDelivery() == emailverify.DeliveryCode {
						delivery = "code"
					} else {
						delivery = "link"
					}
				}
			}
		}
	}

	if delivery == "" {
		delivery = "link"
	}

	err := s.auth.ForgotPasswordWithOpts(r.Context(), auth.ForgotPasswordOpts{
		Email:     body.Email,
		IP:        clientIP(r),
		TenantID:  tenantID,
		ClientID:  clientID,
		ReturnTo:  returnTo,
		Delivery:  delivery,
		BrandName: brandName,
		AppScoped: appScoped,
	})
	if err == auth.ErrRateLimited {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request_failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":  "If an account exists for that email, we sent reset instructions.",
		"delivery": delivery,
	})
}

func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	code := strings.TrimSpace(body.Code)
	tokenStr := strings.TrimSpace(body.Token)
	email := strings.TrimSpace(body.Email)

	var err error
	if code != "" && email != "" {
		err = s.auth.ResetPasswordWithCode(r.Context(), email, code, body.Password, clientIP(r))
	} else if tokenStr != "" {
		err = s.auth.ResetPassword(r.Context(), tokenStr, body.Password, clientIP(r))
	} else {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token_or_code_required"})
		return
	}

	if err == auth.ErrRateLimited {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if errors.Is(err, passwordpolicy.ErrWeak) || (err != nil && strings.Contains(err.Error(), "at least")) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "weak_password", "message": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Password updated. Sign in with your new password."})
}

func (s *Server) handleSignOut(w http.ResponseWriter, r *http.Request) {
	if !csrfValid(r) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "csrf_invalid"})
		return
	}

	sessionToken := cookieValue(r, sessionCookie)
	refreshToken := cookieValue(r, refreshCookie)
	_ = s.auth.SignOut(r.Context(), sessionToken, refreshToken, clientIP(r))
	clearAuthCookies(w, s.cfg.CookieSecure, s.cfg.CookieDomain)
	writeJSON(w, http.StatusOK, map[string]string{"message": "signed_out"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := s.currentUser(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":             user.ID,
		"email":          user.Email,
		"email_verified": user.EmailVerifiedAt != nil,
	})
}

func (s *Server) currentUser(r *http.Request) (*store.User, error) {
	if bearer := bearerToken(r); bearer != "" {
		return s.auth.UserFromAccess(r.Context(), bearer)
	}
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil, nil
	}
	return s.auth.SessionUser(r.Context(), cookie.Value)
}

func setAuthCookies(w http.ResponseWriter, secure bool, domain string, bundle *auth.TokenBundle) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    bundle.SessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		Domain:   domain,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookie,
		Value:    bundle.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		Domain:   domain,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookie,
		Value:    bundle.CSRFToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   secure,
		Domain:   domain,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
}

func setRefreshCookies(w http.ResponseWriter, secure bool, domain string, bundle *auth.TokenBundle) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookie,
		Value:    bundle.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		Domain:   domain,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
	if bundle.CSRFToken != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     csrfCookie,
			Value:    bundle.CSRFToken,
			Path:     "/",
			HttpOnly: false,
			Secure:   secure,
			Domain:   domain,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		})
	}
}

func clearAuthCookies(w http.ResponseWriter, secure bool, domain string) {
	for _, name := range []string{sessionCookie, refreshCookie, csrfCookie} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: name != csrfCookie,
			Secure:   secure,
			Domain:   domain,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
	}
}

func csrfValid(r *http.Request) bool {
	cookie, err := r.Cookie(csrfCookie)
	if err != nil || cookie.Value == "" {
		return false
	}
	header := r.Header.Get(csrfHeader)
	return header != "" && header == cookie.Value
}

func bearerToken(r *http.Request) string {
	authz := r.Header.Get("Authorization")
	if authz == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authz, prefix) {
		return ""
	}
	return strings.TrimSpace(authz[len(prefix):])
}

func cookieValue(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	return r.RemoteAddr
}
