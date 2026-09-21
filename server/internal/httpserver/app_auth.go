package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/i18n"
	"github.com/sooapps/sooauth/server/internal/passwordpolicy"
	"github.com/sooapps/sooauth/server/internal/store"
)

type AppSignUpInput struct {
	ClientID string
	Email    string
	Username string
	Phone    string
	Password string
	Metadata map[string]any
	ReturnTo string
	IP       string
	Lang     string
}

func (s *Server) signUpAppUser(ctx context.Context, in AppSignUpInput) error {
	client, err := s.oauthClients.FindByClientID(ctx, in.ClientID)
	if err != nil || client == nil || client.TenantID == nil {
		return store.ErrInvalidClient
	}
	tenantID := *client.TenantID

	tenant, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil || tenant == nil {
		return store.ErrInvalidClient
	}

	// Validate allowed identifiers
	allowed := make(map[string]bool)
	for _, idKey := range tenant.AuthConfig.AllowedIdentifiers {
		allowed[strings.ToLower(strings.TrimSpace(idKey))] = true
	}
	if len(allowed) == 0 {
		allowed["email"] = true
	}

	email := strings.ToLower(strings.TrimSpace(in.Email))
	username := strings.ToLower(strings.TrimSpace(in.Username))
	phone := strings.TrimSpace(in.Phone)

	if email != "" && !allowed["email"] {
		return errors.New("email sign up is disabled for this project")
	}
	if username != "" && !allowed["username"] {
		return errors.New("username sign up is disabled for this project")
	}
	if phone != "" && !allowed["phone"] {
		return errors.New("phone sign up is disabled for this project")
	}

	if email == "" && username == "" && phone == "" {
		return errors.New("an identifier (email, username, or phone) is required")
	}

	// Validate password unless OTP-only mode
	if tenant.AuthConfig.PrimaryAuthMode != "otp" {
		if strings.TrimSpace(in.Password) == "" {
			return errors.New("password required")
		}
		if err := tenant.PasswordPolicy().Validate(in.Password); err != nil {
			return passwordpolicy.ErrWeak
		}
	}

	// Validate custom fields from registration_schema
	if in.Metadata == nil {
		in.Metadata = make(map[string]any)
	}
	for _, field := range tenant.RegistrationSchema {
		if field.Required {
			val, exists := in.Metadata[field.ID]
			if !exists || val == nil || fmt.Sprintf("%v", val) == "" {
				return fmt.Errorf("field_required:%s", field.Label)
			}
		}
	}

	theme, _ := s.theme.GetByTenant(ctx, tenantID)
	brand := appDisplayBrand(tenant, theme, client)
	appReturnTo := resolveAppReturnTo(in.ReturnTo, s.cfg.AppURL, client)

	user, err := s.users.CreateAppUserFlexible(ctx, tenantID, email, username, phone, in.Password, in.Metadata)
	if err != nil {
		return err
	}

	method := "email"
	if email == "" {
		if username != "" {
			method = "username"
		} else if phone != "" {
			method = "phone"
		}
	}
	if in.Password == "" {
		method = "otp"
	}

	if email == "" || !tenant.EmailVerifyRequired {
		if email != "" {
			_ = s.users.MarkEmailVerified(ctx, user.ID)
		}
		s.webhooks.Emit(ctx, &tenantID, "user.created", user, map[string]any{"method": method, "ip": in.IP})
		return nil
	}

	if err := s.auth.SendAppVerificationEmail(ctx, auth.AppVerificationEmailOpts{
		BrandName: brand,
		User:      user,
		Delivery:  tenant.VerifyDelivery(),
		ClientID:  in.ClientID,
		ReturnTo:  appReturnTo,
		Lang:      in.Lang,
	}); err != nil {
		_ = s.users.DeleteUnverified(ctx, user.ID)
		return fmt.Errorf("%w", auth.ErrEmailDelivery)
	}
	s.webhooks.Emit(ctx, &tenantID, "user.created", user, map[string]any{"method": method, "ip": in.IP})
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
		Lang:      i18n.Resolve(r, tenant.DefaultLocale),
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

func (s *Server) signInAppUser(ctx context.Context, clientID, identifier, password, ip, userAgent string, rememberMe bool) (*auth.TokenBundle, error) {
	client, err := s.oauthClients.FindByClientID(ctx, clientID)
	if err != nil || client == nil || client.TenantID == nil {
		return nil, store.ErrInvalidClient
	}
	tenantID := *client.TenantID

	tenant, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil || tenant == nil {
		return nil, store.ErrInvalidClient
	}

	return s.auth.SignInAppByIdentifierWithOptions(ctx, tenantID, identifier, password, ip, userAgent, tenant.EmailVerifyRequired, auth.SignInOptions{RememberMe: rememberMe})
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
	authCfg := store.DefaultAuthConfig()
	var sanitizedSchema []map[string]any
	if tenant != nil {
		policy = tenant.PasswordPolicy()
		emailVerifyRequired = tenant.EmailVerifyRequired
		verifyDelivery = tenant.VerifyDelivery().String()
		authCfg = tenant.AuthConfig

		for _, field := range tenant.RegistrationSchema {
			fMap := map[string]any{
				"id":          field.ID,
				"label":       field.Label,
				"type":        field.Type,
				"required":    field.Required,
				"placeholder": field.Placeholder,
				"description": field.Description,
			}
			if field.OptionsSource != nil {
				if field.OptionsSource.Type == "static" {
					fMap["options"] = field.OptionsSource.Options
				} else if field.OptionsSource.Type == "dynamic_api" {
					fMap["options_url"] = fmt.Sprintf("%s/v1/widget/fields/%s/options?client_id=%s", strings.TrimRight(s.cfg.AppURL, "/"), field.ID, client.ClientID)
				}
			}
			sanitizedSchema = append(sanitizedSchema, fMap)
		}
	}
	if sanitizedSchema == nil {
		sanitizedSchema = []map[string]any{}
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
		"auth_config":         authCfg,
		"registration_schema": sanitizedSchema,
		"issuer":              strings.TrimRight(s.cfg.AppURL, "/"),
	})
}

func (s *Server) handleWidgetFieldOptions(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	fieldID := r.PathValue("field_id")
	if fieldID == "" {
		fieldID = strings.TrimPrefix(r.URL.Path, "/v1/widget/fields/")
		fieldID = strings.TrimSuffix(fieldID, "/options")
	}
	if fieldID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "field_id_required"})
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

	tenant, err := s.tenants.FindByID(r.Context(), *client.TenantID)
	if err != nil || tenant == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "tenant_not_found"})
		return
	}

	var targetField *store.RegistrationField
	for _, f := range tenant.RegistrationSchema {
		if f.ID == fieldID {
			targetField = &f
			break
		}
	}
	if targetField == nil || targetField.OptionsSource == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "field_not_found"})
		return
	}

	if s.dynamicOptions == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "dynamic_options_not_configured"})
		return
	}

	options, err := s.dynamicOptions.FetchOptions(r.Context(), *targetField.OptionsSource)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed_to_fetch_options", "message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"options": options,
	})
}

type otpPayload struct {
	Code       string `json:"code"`
	Identifier string `json:"identifier"`
	ClientID   string `json:"client_id"`
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

	client, err := s.oauthClients.FindByClientID(r.Context(), clientID)
	if err != nil || client == nil || client.TenantID == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "client_not_found"})
		return
	}
	tenantID := *client.TenantID

	code := fmt.Sprintf("%06d", (time.Now().UnixNano()%900000)+100000)

	cacheKey := fmt.Sprintf("otp:%s:%s", tenantID.String(), identifier)
	if s.ephemeral == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ephemeral_store_not_configured"})
		return
	}
	if err := s.ephemeral.Set(r.Context(), cacheKey, otpPayload{
		Code:       code,
		Identifier: identifier,
		ClientID:   clientID,
	}, 5*time.Minute); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_store_otp"})
		return
	}

	if s.webhooks != nil {
		s.webhooks.Emit(r.Context(), &tenantID, "otp.sent", nil, map[string]any{
			"identifier": identifier,
			"channel":    body.Channel,
			"ip":         clientIP(r),
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Verification code sent.",
	})
}

func (s *Server) handleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	if corsPreflight(w, r) {
		return
	}

	var body struct {
		ClientID   string `json:"client_id"`
		Identifier string `json:"identifier"`
		Code       string `json:"code"`
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

	client, err := s.oauthClients.FindByClientID(r.Context(), clientID)
	if err != nil || client == nil || client.TenantID == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "client_not_found"})
		return
	}
	tenantID := *client.TenantID

	if s.ephemeral == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ephemeral_store_not_configured"})
		return
	}
	cacheKey := fmt.Sprintf("otp:%s:%s", tenantID.String(), identifier)
	var payload otpPayload
	ok, err := s.ephemeral.Get(r.Context(), cacheKey, &payload)
	if err != nil || !ok || payload.Code != code {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_code"})
		return
	}
	_ = s.ephemeral.Delete(r.Context(), cacheKey)

	user, err := s.users.FindAppUserByIdentifierSimple(r.Context(), tenantID, identifier)
	if err != nil || user == nil {
		var email, username, phone string
		if strings.Contains(identifier, "@") {
			email = identifier
		} else if strings.HasPrefix(identifier, "+") || (len(identifier) >= 7 && !strings.ContainsAny(identifier, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")) {
			phone = identifier
		} else {
			username = identifier
		}
		user, err = s.users.CreateAppUserFlexible(r.Context(), tenantID, email, username, phone, "", nil)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed_to_create_user"})
			return
		}
	}

	if user.Phone != nil && *user.Phone == identifier {
		_ = s.users.MarkPhoneVerified(r.Context(), user.ID)
	} else if user.Email == identifier {
		_ = s.users.MarkEmailVerified(r.Context(), user.ID)
	}

	bundle, err := s.auth.SignInVerifiedUser(r.Context(), user, clientIP(r), r.UserAgent())
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "signin_failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":      "Signed in successfully.",
		"access_token": bundle.AccessToken,
		"email":        bundle.User.Email,
		"expires_in":   bundle.ExpiresIn,
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
