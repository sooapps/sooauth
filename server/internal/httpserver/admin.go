package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/accounts"
	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/billing"
	"github.com/sooapps/sooauth/server/internal/emailverify"
	"github.com/sooapps/sooauth/server/internal/passwordpolicy"
	"github.com/sooapps/sooauth/server/internal/providers"
	"github.com/sooapps/sooauth/server/internal/social"
	"github.com/sooapps/sooauth/server/internal/store"
	"github.com/sooapps/sooauth/server/internal/webhooks"
)

const projectCookie = "sooauth_project"

type dashboardCtx struct {
	user    *store.User
	account *store.Account
	tenant  *store.Tenant
	client  *store.OAuthClient
}

func (s *Server) requireDashboard(w http.ResponseWriter, r *http.Request) (*dashboardCtx, bool) {
	user, err := s.currentUser(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return nil, false
	}

	platform, err := s.users.IsPlatformOwner(r.Context(), user.ID)
	if err != nil || !platform {
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error":   "platform_access_denied",
			"message": "This account is for app sign-in only. Create a separate account at sooauth.com to manage projects.",
		})
		return nil, false
	}

	account, err := s.accounts.FindByOwner(r.Context(), user.ID)
	if err != nil || account == nil {
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error":   "platform_access_denied",
			"message": "No platform account found. Sign up at sooauth.com to use the dashboard.",
		})
		return nil, false
	}

	tenant, client, err := s.loadDashboardProject(r, account)
	if err != nil || tenant == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "no_project"})
		return nil, false
	}

	return &dashboardCtx{user: user, account: account, tenant: tenant, client: client}, true
}

func (s *Server) loadDashboardProject(r *http.Request, account *store.Account) (*store.Tenant, *store.OAuthClient, error) {
	if active := s.activeTenant(r, account.ID); active != nil {
		if t, err := s.tenants.FindByIDForAccount(r.Context(), active.ID, account.ID); err == nil && t != nil {
			client, _ := s.oauthClients.FindPrimaryByTenant(r.Context(), t.ID)
			return t, client, nil
		}
	}
	list, err := s.tenants.ListByAccount(r.Context(), account.ID)
	if err != nil || len(list) == 0 {
		return nil, nil, err
	}
	tenant := list[0]
	client, err := s.oauthClients.FindPrimaryByTenant(r.Context(), tenant.ID)
	return &tenant, client, err
}

func (s *Server) activeTenant(r *http.Request, accountID uuid.UUID) *store.Tenant {
	if c, err := r.Cookie(projectCookie); err == nil && c.Value != "" {
		if id, err := uuid.Parse(c.Value); err == nil {
			if t, err := s.tenants.FindByIDForAccount(r.Context(), id, accountID); err == nil && t != nil {
				return t
			}
		}
	}
	list, err := s.tenants.ListByAccount(r.Context(), accountID)
	if err != nil || len(list) == 0 {
		return nil
	}
	return &list[0]
}

func (s *Server) setActiveProject(w http.ResponseWriter, tenantID uuid.UUID) {
	http.SetCookie(w, &http.Cookie{
		Name:     projectCookie,
		Value:    tenantID.String(),
		Path:     "/",
		HttpOnly: false,
		Secure:   s.cfg.CookieSecure,
		Domain:   s.cfg.CookieDomain,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((365 * 24 * time.Hour).Seconds()),
	})
}

func (s *Server) handleDashboardMe(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}

	projects, _ := s.tenants.ListByAccount(r.Context(), dash.account.ID)
	projectList := make([]map[string]any, 0, len(projects))
	for _, p := range projects {
		projectList = append(projectList, map[string]any{
			"id":   p.ID,
			"name": p.Name,
		})
	}

	integration := map[string]any{
		"issuer":    strings.TrimRight(s.cfg.AppURL, "/"),
		"client_id": "",
	}
	if dash.client != nil {
		integration["client_id"] = dash.client.ClientID
		integration["redirect_uris"] = dash.client.RedirectURIs
		integration["public"] = dash.client.Public
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"email": dash.user.Email,
		"id":    dash.user.ID,
		"account": map[string]any{
			"plan":          dash.account.Plan,
			"plan_label":    billing.PlanLabel(dash.account.Plan),
			"max_projects":  billing.MaxProjects(dash.account.Plan),
			"project_count": len(projects),
		},
		"projects": projectList,
		"project": map[string]any{
			"id":                    dash.tenant.ID,
			"name":                  dash.tenant.Name,
			"email_verify_required": dash.tenant.EmailVerifyRequired,
		},
		"integration": integration,
	})
}

func (s *Server) handleDashboardProjects(w http.ResponseWriter, r *http.Request) {
	user, err := s.currentUser(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	account, err := s.accounts.FindOrCreateByOwner(r.Context(), user.ID)
	if err != nil || account == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "account_failed"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		projects, err := s.tenants.ListByAccount(r.Context(), account.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
			return
		}
		out := make([]map[string]any, 0, len(projects))
		for _, p := range projects {
			out = append(out, map[string]any{"id": p.ID, "name": p.Name})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"projects":     out,
			"plan":         account.Plan,
			"max_projects": billing.MaxProjects(account.Plan),
		})
	case http.MethodPost:
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		tenant, client, err := accounts.CreateProject(r.Context(), s.accounts, s.tenants, s.oauthClients, s.theme, user.ID, strings.TrimSpace(body.Name))
		if err != nil {
			if strings.Contains(err.Error(), "limit") {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error":   "project_limit",
					"message": "Upgrade your plan to add more projects.",
				})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create_failed"})
			return
		}
		s.setActiveProject(w, tenant.ID)
		writeJSON(w, http.StatusCreated, map[string]any{
			"id":        tenant.ID,
			"name":      tenant.Name,
			"client_id": client.ClientID,
		})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
	}
}

func (s *Server) handleDashboardSelectProject(w http.ResponseWriter, r *http.Request) {
	user, err := s.currentUser(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	account, err := s.accounts.FindOrCreateByOwner(r.Context(), user.ID)
	if err != nil || account == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "account_failed"})
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/dashboard/api/projects/")
	idStr = strings.TrimSuffix(idStr, "/select")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_id"})
		return
	}
	t, err := s.tenants.FindByIDForAccount(r.Context(), id, account.ID)
	if err != nil || t == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	s.setActiveProject(w, id)
	writeJSON(w, http.StatusOK, map[string]string{"message": "selected"})
}

func (s *Server) handleDashboardProject(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}

	if r.Method == http.MethodGet {
		policy := dash.tenant.PasswordPolicy()
		writeJSON(w, http.StatusOK, map[string]any{
			"name":                     dash.tenant.Name,
			"email_verify_required":    dash.tenant.EmailVerifyRequired,
			"email_verify_delivery":    dash.tenant.VerifyDelivery().String(),
			"password_min_length":      policy.MinLength,
			"password_require_upper":   policy.RequireUppercase,
			"password_require_number":  policy.RequireNumber,
			"password_require_special": policy.RequireSpecial,
			"password_hint":            policy.Describe(),
		})
		return
	}

	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	var body struct {
		Name                   string `json:"name"`
		EmailVerifyRequired    bool   `json:"email_verify_required"`
		EmailVerifyDelivery    string `json:"email_verify_delivery"`
		PasswordMinLength      int    `json:"password_min_length"`
		PasswordRequireUpper   bool   `json:"password_require_upper"`
		PasswordRequireNumber  bool   `json:"password_require_number"`
		PasswordRequireSpecial bool   `json:"password_require_special"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = dash.tenant.Name
	}
	policy := passwordpolicy.Policy{
		MinLength:        body.PasswordMinLength,
		RequireUppercase: body.PasswordRequireUpper,
		RequireNumber:    body.PasswordRequireNumber,
		RequireSpecial:   body.PasswordRequireSpecial,
	}.Normalize()
	if err := s.tenants.UpdateSettings(r.Context(), dash.tenant.ID, name, body.EmailVerifyRequired, emailverify.ParseDelivery(body.EmailVerifyDelivery), policy); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "saved"})
}

func (s *Server) handleDashboardIntegration(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	if dash.client == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "client_missing"})
		return
	}

	base := strings.TrimRight(s.cfg.AppURL, "/")
	if r.Method == http.MethodGet {
		platformGoogleCallback := base + "/auth/social/google/callback"
		appGoogleCallback := social.CallbackURL(s.cfg.AppURL, "google", dash.tenant, dash.account.Plan, true)
		writeJSON(w, http.StatusOK, map[string]any{
			"issuer":                       base,
			"client_id":                    dash.client.ClientID,
			"redirect_uris":                dash.client.RedirectURIs,
			"public":                       dash.client.Public,
			"scopes":                       []string{"openid", "email", "profile"},
			"authorize_url":                base + "/oauth/authorize",
			"token_url":                    base + "/oauth/token",
			"userinfo_url":                 base + "/oauth/userinfo",
			"discovery_url":                base + "/.well-known/openid-configuration",
			"plan":                         dash.account.Plan,
			"can_custom_social_callback":   billing.CanCustomSocialCallback(dash.account.Plan),
			"social_callback_origin":       strings.TrimSpace(dash.tenant.SocialCallbackOrigin),
			"platform_google_callback_url": platformGoogleCallback,
			"app_google_callback_url":      appGoogleCallback,
			"google_proxy_route":           "/api/auth/google/callback",
		})
		return
	}

	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	var body struct {
		RedirectURIs         []string `json:"redirect_uris"`
		SocialCallbackOrigin *string  `json:"social_callback_origin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	uris := make([]string, 0, len(body.RedirectURIs))
	for _, u := range body.RedirectURIs {
		u = strings.TrimSpace(u)
		if u != "" {
			uris = append(uris, u)
		}
	}
	if len(uris) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "redirect_uris_required"})
		return
	}
	if body.SocialCallbackOrigin != nil {
		origin := strings.TrimSpace(*body.SocialCallbackOrigin)
		if origin != "" {
			if !billing.CanCustomSocialCallback(dash.account.Plan) {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error":   "plan_required",
					"message": "Custom Google callback origin requires Pro or Business.",
				})
				return
			}
			if err := social.ValidateCallbackOrigin(origin, uris, s.cfg.Env == "production"); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error":   "invalid_social_callback_origin",
					"message": err.Error(),
				})
				return
			}
		}
		if err := s.tenants.UpdateSocialCallbackOrigin(r.Context(), dash.tenant.ID, origin); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save_failed"})
			return
		}
	}
	if err := s.oauthClients.Update(r.Context(), dash.tenant.ID, dash.client.ClientID, dash.client.Name, uris); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "saved"})
}

func (s *Server) handleDashboardUsers(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	users, err := s.users.ListAppEndUsersByTenant(r.Context(), dash.tenant.ID, dash.tenant.OwnerUserID, 100)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{
			"id":             u.ID,
			"email":          u.Email,
			"email_verified": u.EmailVerifiedAt != nil,
			"disabled":       u.DisabledAt != nil,
			"disabled_at":    u.DisabledAt,
			"has_password":   u.HasPassword,
			"providers":      u.Providers,
			"signup_method":  signupMethodLabel(u.SignupMethod, u.HasPassword, u.Providers),
			"created_at":     u.CreatedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": out})
}

func (s *Server) handleDashboardDeleteUser(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	idStr := r.PathValue("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_id"})
		return
	}

	linked, err := s.tenantUsers.IsLinked(r.Context(), dash.tenant.ID, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lookup_failed"})
		return
	}
	if !linked {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}

	platformOwner, err := s.users.IsPlatformOwner(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lookup_failed"})
		return
	}
	if platformOwner {
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error":   "protected_platform_owner",
			"message": "This account owns the sooauth dashboard and cannot be removed from app users.",
		})
		return
	}
	if dash.tenant.OwnerUserID == userID {
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error":   "protected_project_owner",
			"message": "This account owns the project and cannot be removed from app users.",
		})
		return
	}

	_ = s.sessions.RevokeAllForUser(r.Context(), userID)
	if err := s.tenantUsers.Unlink(r.Context(), dash.tenant.ID, userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "remove_failed"})
		return
	}
	_ = s.users.DeleteOrphanAppUser(r.Context(), userID)

	uid := dash.user.ID
	s.audit.Log(r.Context(), &uid, "app_user_removed", map[string]any{
		"removed_user_id": userID.String(),
		"tenant_id":       dash.tenant.ID.String(),
	}, clientIP(r))

	writeJSON(w, http.StatusOK, map[string]string{"message": "removed"})
}

func (s *Server) handleDashboardDisableUser(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok || !accountCSRFValid(w, r) {
		return
	}
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_id"})
		return
	}
	linked, err := s.tenantUsers.IsLinked(r.Context(), dash.tenant.ID, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lookup_failed"})
		return
	}
	if !linked {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	platformOwner, err := s.users.IsPlatformOwner(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lookup_failed"})
		return
	}
	if platformOwner || dash.tenant.OwnerUserID == userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "protected_user"})
		return
	}
	changed, err := s.users.SetDisabledForTenant(r.Context(), dash.tenant.ID, userID, true)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "disable_failed"})
		return
	}
	if !changed {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	_ = s.sessions.RevokeAllForUser(r.Context(), userID)
	uid := dash.user.ID
	s.audit.Log(r.Context(), &uid, "app_user_disabled", map[string]any{"disabled_user_id": userID.String(), "tenant_id": dash.tenant.ID.String()}, clientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "disabled"})
}

func (s *Server) handleDashboardEnableUser(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok || !accountCSRFValid(w, r) {
		return
	}
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_id"})
		return
	}
	changed, err := s.users.SetDisabledForTenant(r.Context(), dash.tenant.ID, userID, false)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "enable_failed"})
		return
	}
	if !changed {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	uid := dash.user.ID
	s.audit.Log(r.Context(), &uid, "app_user_enabled", map[string]any{"enabled_user_id": userID.String(), "tenant_id": dash.tenant.ID.String()}, clientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "enabled"})
}

func (s *Server) handleDashboardResendUserVerification(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok || !accountCSRFValid(w, r) {
		return
	}
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_id"})
		return
	}
	linked, err := s.tenantUsers.IsLinked(r.Context(), dash.tenant.ID, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lookup_failed"})
		return
	}
	if !linked {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	user, err := s.users.FindByID(r.Context(), userID)
	if err != nil || user == nil || user.TenantID == nil || *user.TenantID != dash.tenant.ID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	platformOwner, err := s.users.IsPlatformOwner(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lookup_failed"})
		return
	}
	if platformOwner || userID == dash.tenant.OwnerUserID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "protected_user"})
		return
	}
	if user.EmailVerifiedAt != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "already_verified"})
		return
	}
	client := dash.client
	if client == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "client_unavailable"})
		return
	}
	theme, _ := s.theme.GetByTenant(r.Context(), dash.tenant.ID)
	err = s.auth.ResendAppVerification(r.Context(), auth.AppVerificationEmailOpts{
		BrandName: appDisplayBrand(dash.tenant, theme, client),
		Delivery:  dash.tenant.VerifyDelivery(),
		ClientID:  client.ClientID,
		ReturnTo:  resolveAppReturnTo("", s.cfg.AppURL, client),
	}, dash.tenant.ID, user.Email, clientIP(r))
	if err == auth.ErrRateLimited {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if errors.Is(err, auth.ErrEmailDelivery) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "email_delivery_failed"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resend_failed"})
		return
	}
	uid := dash.user.ID
	s.audit.Log(r.Context(), &uid, "app_user_verification_resent", map[string]any{"app_user_id": userID.String(), "tenant_id": dash.tenant.ID.String()}, clientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "verification_sent"})
}

func signupMethodLabel(method *string, hasPassword bool, providers []string) string {
	if method != nil && *method != "" {
		return providerDisplayName(*method)
	}
	if hasPassword {
		return "Email"
	}
	if len(providers) > 0 {
		return providerDisplayName(providers[0])
	}
	return "—"
}

func providerDisplayName(name string) string {
	switch name {
	case "google":
		return "Google"
	case "github":
		return "GitHub"
	default:
		return name
	}
}

func authMethods(hasPassword bool, providers []string) string {
	var parts []string
	for _, p := range providers {
		switch p {
		case "google":
			parts = append(parts, "Google")
		case "github":
			parts = append(parts, "GitHub")
		default:
			if p != "" {
				parts = append(parts, p)
			}
		}
	}
	if hasPassword {
		parts = append(parts, "Email")
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, " · ")
}

func (s *Server) handleDashboardSessions(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	rows, err := s.sessions.ListActiveByTenant(r.Context(), dash.tenant.ID, dash.user.ID, 100)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{
			"id":         row.ID,
			"email":      row.Email,
			"ip":         deref(row.IP),
			"expires_at": row.ExpiresAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": out})
}

func (s *Server) handleDashboardRevokeSession(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/dashboard/api/sessions/")
	idStr = strings.TrimSuffix(idStr, "/revoke")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_id"})
		return
	}
	revoked, err := s.sessions.RevokeByIDForTenant(r.Context(), id, dash.tenant.ID, dash.user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "revoke_failed"})
		return
	}
	if !revoked {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	s.webhooks.Emit(r.Context(), &dash.tenant.ID, "session.revoked", nil, map[string]any{"session_id": id.String(), "ip": clientIP(r)})
	writeJSON(w, http.StatusOK, map[string]string{"message": "revoked"})
}

func (s *Server) handleDashboardAudit(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	entries, err := s.audit.ListByTenant(r.Context(), dash.tenant.ID, dash.user.ID, 100)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}
	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]any{
			"action":     e.Action,
			"ip":         deref(e.IP),
			"created_at": e.CreatedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": out})
}

func (s *Server) handleDashboardWebhooks(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	webhookStore := s.webhooksStore()
	if r.Method == http.MethodGet {
		endpoints, err := webhookStore.ListByTenant(r.Context(), dash.tenant.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
			return
		}
		out := make([]map[string]any, 0, len(endpoints))
		for _, e := range endpoints {
			out = append(out, map[string]any{"id": e.ID, "url": e.URL, "events": e.Events, "active": e.Active, "created_at": e.CreatedAt.Format(time.RFC3339)})
		}
		writeJSON(w, http.StatusOK, map[string]any{"webhooks": out})
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}
	var body struct {
		URL    string   `json:"url"`
		Events []string `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	body.URL = strings.TrimSpace(body.URL)
	if err := webhooks.ValidateURL(body.URL); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_url"})
		return
	}
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "secret_failed"})
		return
	}
	secret := base64.RawURLEncoding.EncodeToString(secretBytes)
	endpoint, err := webhookStore.Create(r.Context(), dash.tenant.ID, body.URL, secret, cleanWebhookEvents(body.Events))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create_failed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": endpoint.ID, "url": endpoint.URL, "events": endpoint.Events, "active": endpoint.Active, "secret": secret})
}

func (s *Server) handleDashboardWebhookDelete(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_id"})
		return
	}
	deleted, err := s.webhooksStore().Delete(r.Context(), dash.tenant.ID, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete_failed"})
		return
	}
	if !deleted {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (s *Server) handleDashboardWebhookTest(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_id"})
		return
	}
	endpoint, err := s.webhooksStore().FindByIDForTenant(r.Context(), dash.tenant.ID, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lookup_failed"})
		return
	}
	if endpoint == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	if err := s.webhooks.DeliverTest(r.Context(), endpoint, dash.tenant.ID); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "delivery_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "delivered"})
}

func (s *Server) handleDashboardWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_id"})
		return
	}
	attempts, err := s.webhooksStore().ListAttemptsByEndpoint(r.Context(), id, dash.tenant.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "deliveries_unavailable"})
		return
	}
	out := make([]map[string]any, 0, len(attempts))
	for _, attempt := range attempts {
		out = append(out, map[string]any{
			"id": attempt.ID, "delivery_id": attempt.DeliveryID, "event": attempt.EventType,
			"payload": attempt.Payload, "attempt": attempt.Attempt, "status_code": attempt.StatusCode,
			"response": attempt.Response, "error": attempt.Error, "delivered_at": attempt.DeliveredAt,
			"created_at": attempt.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"deliveries": out})
}

func (s *Server) webhooksStore() *store.Webhooks { return s.webhookStore }

func cleanWebhookEvents(events []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(events))
	for _, event := range events {
		event = strings.TrimSpace(event)
		if event != "" && !seen[event] {
			seen[event] = true
			out = append(out, event)
		}
	}
	return out
}

func (s *Server) handleDashboardProviders(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	rows, err := s.oauthProviders.ListByTenant(r.Context(), dash.tenant.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}

	out := make([]map[string]any, 0, len(rows))
	for _, p := range rows {
		platformReady := providers.PlatformReady(s.cfg, p.Provider)
		usingCustom := p.ClientID != "" && !p.UsePlatform
		loginReady := providers.LoginSupported(p.Provider) && (platformReady || usingCustom)
		hasSecret := usingCustom && strings.TrimSpace(p.ClientSecretEncrypted) != ""
		callbackURL := social.CallbackURL(s.cfg.AppURL, p.Provider, dash.tenant, dash.account.Plan, usingCustom)
		entry := map[string]any{
			"provider":              p.Provider,
			"enabled":               p.Enabled,
			"use_platform":          !usingCustom,
			"credential_mode":       credentialMode(usingCustom),
			"platform_ready":        platformReady,
			"login_ready":           loginReady,
			"has_custom":            usingCustom,
			"has_client_secret":     hasSecret,
			"can_use_custom":        providers.LoginSupported(p.Provider),
			"callback_url":          callbackURL,
			"platform_callback_url": strings.TrimRight(s.cfg.AppURL, "/") + "/auth/social/" + p.Provider + "/callback",
		}
		if usingCustom {
			entry["client_id"] = p.ClientID
			entry["client_id_hint"] = maskClientID(p.ClientID)
		}
		out = append(out, entry)
	}
	writeJSON(w, http.StatusOK, map[string]any{"providers": out})
}

func credentialMode(custom bool) string {
	if custom {
		return "custom"
	}
	return "platform"
}

func maskClientID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	if len(id) <= 10 {
		return id[:2] + "…"
	}
	return id[:6] + "…" + id[len(id)-4:]
}

func (s *Server) handleDashboardProviderUpdate(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	provider := strings.TrimPrefix(r.URL.Path, "/dashboard/api/providers/")
	if provider != "google" && provider != "github" && provider != "facebook" && provider != "x" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown_provider"})
		return
	}

	var body struct {
		Enabled      bool   `json:"enabled"`
		Custom       bool   `json:"custom"`
		UsePlatform  bool   `json:"use_platform"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	existing, _ := s.oauthProviders.FindByTenantProvider(r.Context(), dash.tenant.ID, provider)
	enabled := body.Enabled
	if !body.Enabled && existing == nil {
		enabled = false
	} else if existing != nil && !body.Custom && !body.UsePlatform && body.ClientID == "" && body.ClientSecret == "" {
		enabled = body.Enabled
	}

	if body.UsePlatform {
		if err := s.oauthProviders.RevertToPlatform(r.Context(), dash.tenant.ID, provider, enabled); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "saved"})
		return
	}

	if body.Custom {
		if !providers.LoginSupported(provider) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "custom_not_supported"})
			return
		}
		if strings.TrimSpace(body.ClientID) == "" {
			if existing != nil && existing.ClientID != "" && !existing.UsePlatform {
				body.ClientID = existing.ClientID
			} else {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "client_id_required"})
				return
			}
		}
		secret := body.ClientSecret
		if secret == "" && existing != nil && !existing.UsePlatform {
			secret = existing.ClientSecretEncrypted
		}
		if secret == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "client_secret_required"})
			return
		}
		if err := s.oauthProviders.UpsertCustom(r.Context(), dash.tenant.ID, provider, body.ClientID, secret, enabled); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "saved"})
		return
	}

	if existing != nil {
		if err := s.oauthProviders.SetEnabled(r.Context(), dash.tenant.ID, provider, enabled); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "saved"})
		return
	}

	if enabled && providers.LoginSupported(provider) && !providers.PlatformReady(s.cfg, provider) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "platform_unavailable",
			"message": "Platform OAuth is not configured for " + provider + ". Add custom credentials instead.",
		})
		return
	}

	if err := s.oauthProviders.TogglePlatform(r.Context(), dash.tenant.ID, provider, enabled); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "saved"})
}

func (s *Server) handleDashboardThemeGet(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	theme, _ := s.theme.GetByTenant(r.Context(), dash.tenant.ID)
	canCustomize := billing.CanCustomizeTheme(dash.account.Plan)
	canHide := billing.CanHideBranding(dash.account.Plan)

	brandName := theme.BrandName
	logoURL := theme.LogoURL
	accent := theme.AccentColor
	hideBranding := theme.HideBranding
	if !canCustomize {
		brandName = "sooauth"
		logoURL = ""
		hideBranding = false
	} else if !canHide {
		hideBranding = false
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"brand_name":           brandName,
		"logo_url":             logoURL,
		"accent_color":         accent,
		"hide_branding":        hideBranding,
		"can_customize_brand":  canCustomize,
		"can_hide_branding":    canHide,
		"free_branding_locked": !canCustomize,
	})
}

func (s *Server) handleDashboardThemePut(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	var body struct {
		BrandName    string `json:"brand_name"`
		LogoURL      string `json:"logo_url"`
		AccentColor  string `json:"accent_color"`
		HideBranding bool   `json:"hide_branding"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	canCustomize := billing.CanCustomizeTheme(dash.account.Plan)
	canHide := billing.CanHideBranding(dash.account.Plan)

	brandName := "sooauth"
	logoURL := ""
	accent := body.AccentColor
	hideBranding := false

	if canCustomize {
		brandName = strings.TrimSpace(body.BrandName)
		if brandName == "" {
			brandName = "sooauth"
		}
		logoURL = strings.TrimSpace(body.LogoURL)
	}
	if accent == "" {
		accent = "#E8FF3F"
	}
	if canHide {
		hideBranding = body.HideBranding
	}

	if err := s.theme.UpdateByTenant(r.Context(), dash.tenant.ID, brandName, logoURL, accent, hideBranding); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "saved"})
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *Server) themeView(ctx context.Context) store.Theme {
	theme, err := s.theme.Get(ctx)
	if err != nil {
		return store.Theme{
			BrandName:   s.cfg.BrandName,
			LogoURL:     s.cfg.BrandLogoURL,
			AccentColor: s.cfg.BrandAccent,
		}
	}
	return theme
}
