package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/passwordpolicy"
	"github.com/sooapps/sooauth/server/internal/store"
)

func (s *Server) signUpAppUser(ctx context.Context, clientID, email, password, returnTo, ip string) error {
	client, err := s.oauthClients.FindByClientID(ctx, clientID)
	if err != nil || client == nil || client.TenantID == nil {
		return store.ErrInvalidClient
	}
	tenantID := *client.TenantID

	tenant, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil || tenant == nil {
		return store.ErrInvalidClient
	}
	if err := tenant.PasswordPolicy().Validate(password); err != nil {
		return passwordpolicy.ErrWeak
	}

	theme, _ := s.theme.GetByTenant(ctx, tenantID)
	brand := appDisplayBrand(tenant, theme, client)
	appReturnTo := resolveAppReturnTo(returnTo, s.cfg.AppURL, client)

	user, err := s.users.CreateAppUser(ctx, tenantID, email, password)
	if errors.Is(err, store.ErrEmailTaken) {
		return store.ErrEmailTaken
	}
	if err != nil {
		return err
	}

	if err := s.tenantUsers.LinkWithMethod(ctx, tenantID, user.ID, "email"); err != nil {
		_ = s.users.DeleteUnverified(ctx, user.ID)
		return err
	}

	if !tenant.EmailVerifyRequired {
		_ = s.users.MarkEmailVerified(ctx, user.ID)
		s.webhooks.Emit(ctx, &tenantID, "user.created", user, map[string]any{"method": "email", "ip": ip})
		return nil
	}

	if err := s.auth.SendAppVerificationEmail(ctx, auth.AppVerificationEmailOpts{
		BrandName: brand,
		User:      user,
		Delivery:  tenant.VerifyDelivery(),
		ClientID:  clientID,
		ReturnTo:  appReturnTo,
	}); err != nil {
		_ = s.users.DeleteUnverified(ctx, user.ID)
		return fmt.Errorf("%w", auth.ErrEmailDelivery)
	}
	s.webhooks.Emit(ctx, &tenantID, "user.created", user, map[string]any{"method": "email", "ip": ip})
	return nil
}

func (s *Server) resendAppVerification(r *http.Request, clientID, email, returnTo string) error {
	client, err := s.oauthClients.FindByClientID(r.Context(), clientID)
	if err != nil || client == nil || client.TenantID == nil {
		return store.ErrInvalidClient
	}
	tenant, err := s.tenants.FindByID(r.Context(), *client.TenantID)
	if err != nil || tenant == nil {
		return store.ErrInvalidClient
	}
	user, err := s.users.FindByEmailSimpleInTenant(r.Context(), tenant.ID, email)
	if err != nil || user == nil {
		return nil
	}
	ok, err := s.tenantUsers.HasAccess(r.Context(), tenant.ID, user.ID)
	if err != nil || !ok {
		return nil
	}
	theme, _ := s.theme.GetByTenant(r.Context(), tenant.ID)
	return s.auth.ResendAppVerification(r.Context(), auth.AppVerificationEmailOpts{
		BrandName: appDisplayBrand(tenant, theme, client),
		Delivery:  tenant.VerifyDelivery(),
		ClientID:  clientID,
		ReturnTo:  resolveAppReturnTo(returnTo, s.cfg.AppURL, client),
	}, tenant.ID, email, clientIP(r))
}

func (s *Server) handleVerifyCode(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	var body struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		ClientID string `json:"client_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	clientID := strings.TrimSpace(body.ClientID)
	var tenantID *uuid.UUID
	if clientID != "" {
		client, err := s.oauthClients.FindByClientID(r.Context(), clientID)
		if err != nil || client == nil || client.TenantID == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_client"})
			return
		}
		tenantID = client.TenantID
	}

	user, already, err := s.auth.VerifyEmailByCode(r.Context(), tenantID, body.Email, body.Code, clientIP(r))
	if err == auth.ErrRateLimited {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_code"})
		return
	}
	if user == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_code"})
		return
	}

	if tenantID != nil {
		ok, err := s.tenantUsers.HasAccess(r.Context(), *tenantID, user.ID)
		if err != nil || !ok {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "invalid_code"})
			return
		}
		bundle, err := s.auth.SignInVerifiedUser(r.Context(), user, clientIP(r), r.UserAgent())
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "signin_failed"})
			return
		}
		msg := "Email verified. You are signed in."
		if already {
			msg = "Email already verified. You are signed in."
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"message":      msg,
			"access_token": bundle.AccessToken,
			"email":        bundle.User.Email,
			"expires_in":   bundle.ExpiresIn,
		})
		return
	}

	msg := "Email verified. You can sign in."
	if already {
		msg = "Email already verified. You can sign in."
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": msg})
}

func (s *Server) signInAppUser(ctx context.Context, clientID, email, password, ip, userAgent string, rememberMe bool) (*auth.TokenBundle, error) {
	client, err := s.oauthClients.FindByClientID(ctx, clientID)
	if err != nil || client == nil || client.TenantID == nil {
		return nil, store.ErrInvalidClient
	}
	tenantID := *client.TenantID

	tenant, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil || tenant == nil {
		return nil, store.ErrInvalidClient
	}

	return s.auth.SignInAppWithOptions(ctx, tenantID, email, password, ip, userAgent, tenant.EmailVerifyRequired, auth.SignInOptions{RememberMe: rememberMe})
}

func (s *Server) handleWidgetConfig(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	clientID := strings.TrimSpace(r.URL.Query().Get("client_id"))
	if clientID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "client_id_required"})
		return
	}

	client, err := s.oauthClients.FindByClientID(r.Context(), clientID)
	if err != nil || client == nil || client.TenantID == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "client_not_found"})
		return
	}

	tenantID := *client.TenantID
	tenant, _ := s.tenants.FindByID(r.Context(), tenantID)
	policy := passwordpolicy.Default()
	emailVerifyRequired := true
	verifyDelivery := "link"
	if tenant != nil {
		policy = tenant.PasswordPolicy()
		emailVerifyRequired = tenant.EmailVerifyRequired
		verifyDelivery = tenant.VerifyDelivery().String()
	}

	var theme store.Theme
	if s.theme != nil {
		theme, _ = s.theme.GetByTenant(r.Context(), tenantID)
	}
	var providers []string
	if s.social != nil {
		providers, _ = s.social.EnabledForTenant(r.Context(), tenantID)
	}

	plan := "free"
	if s.accounts != nil {
		account, _ := s.accounts.FindByTenantID(r.Context(), tenantID)
		if account != nil {
			plan = account.Plan
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"client_id":              client.ClientID,
		"tenant_id":              tenantID.String(),
		"brand_name":             appDisplayBrand(tenant, theme, client),
		"logo_url":               theme.LogoURL,
		"accent_color":           theme.AccentColor,
		"show_powered_by":        plan == "free" || !theme.HideBranding,
		"providers":              providers,
		"email_password_enabled": true,
		"remember_me_enabled":    true,
		"remember_me_default":    true,
		"email_verify_required":  emailVerifyRequired,
		"email_verify_delivery":  verifyDelivery,
		"password_policy": map[string]any{
			"min_length":        policy.MinLength,
			"require_uppercase": policy.RequireUppercase,
			"require_number":    policy.RequireNumber,
			"require_special":   policy.RequireSpecial,
			"hint":              policy.Describe(),
		},
		"issuer": strings.TrimRight(s.cfg.AppURL, "/"),
	})
}

func (s *Server) handleWidgetExchange(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	var body struct {
		Code     string `json:"code"`
		ClientID string `json:"client_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	var payload widgetExchangePayload
	ok, err := s.ephemeral.Get(r.Context(), "widget:"+body.Code, &payload)
	if err != nil || !ok || payload.ClientID != body.ClientID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_code"})
		return
	}
	_ = s.ephemeral.Delete(r.Context(), "widget:"+body.Code)

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": payload.AccessToken,
		"email":        payload.Email,
		"expires_in":   payload.ExpiresIn,
	})
}

type widgetExchangePayload struct {
	ClientID    string `json:"client_id"`
	AccessToken string `json:"access_token"`
	Email       string `json:"email"`
	ExpiresIn   int    `json:"expires_in"`
}
