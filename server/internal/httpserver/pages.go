package httpserver

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/i18n"
	"github.com/sooapps/sooauth/server/internal/pages"
	"github.com/sooapps/sooauth/server/internal/social"
)

func (s *Server) pageData(titleKey string, r *http.Request) pages.ViewData {
	ctx := r.Context()
	returnTo := r.URL.Query().Get("return_to")
	theme := s.themeView(ctx)

	tenantID := s.tenantFromReturnTo(ctx, returnTo)
	clientID := clientIDFromReturnTo(returnTo)
	if clientID == "" {
		clientID = strings.TrimSpace(r.URL.Query().Get("client_id"))
	}
	if tenantID == nil && clientID != "" {
		tenantID = s.tenantFromClientID(ctx, clientID)
	}
	var socialProviders []string
	var tenantIDStr string
	var tenantDefault string

	if tenantID != nil {
		tenantIDStr = tenantID.String()
		if ttheme, err := s.theme.GetByTenant(ctx, *tenantID); err == nil {
			theme = ttheme
		}
		socialProviders, _ = s.social.EnabledForTenant(ctx, *tenantID)
		if tenant, _ := s.tenants.FindByID(ctx, *tenantID); tenant != nil {
			tenantDefault = tenant.DefaultLocale
		}
	} else if strings.Contains(returnTo, "/dashboard") {
		socialProviders = social.EnabledProviders(s.cfg)
	}

	lang := i18n.Resolve(r, tenantDefault)
	clientBytes, _ := json.Marshal(i18n.ClientDict(lang))

	return pages.ViewData{
		Title:           i18n.T(lang, titleKey),
		BrandName:       theme.BrandName,
		LogoURL:         theme.LogoURL,
		AccentColor:     theme.AccentColor,
		ReturnTo:        returnTo,
		TenantID:        tenantIDStr,
		ClientID:        clientID,
		SocialProviders: socialProviders,
		Error:           friendlyPageError(lang, r.URL.Query().Get("error")),
		Message:         friendlyPageMessage(r.URL.Query().Get("message")),
		AppURL:          s.cfg.AppURL,
		Lang:            lang,
		T:               i18n.Dict(lang),
		ClientJSON:      template.JS(clientBytes),
	}
}

func applyTenantLocale(data *pages.ViewData, r *http.Request, tenantDefault, titleKey string) {
	lang := i18n.Resolve(r, tenantDefault)
	data.Lang = lang
	data.T = i18n.Dict(lang)
	data.Title = i18n.T(lang, titleKey)
	clientBytes, _ := json.Marshal(i18n.ClientDict(lang))
	data.ClientJSON = template.JS(clientBytes)
}

func clientIDFromReturnTo(returnTo string) string {
	if returnTo == "" {
		return ""
	}
	u, err := url.Parse(returnTo)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(u.Query().Get("client_id"))
}

func (s *Server) tenantFromClientID(ctx context.Context, clientID string) *uuid.UUID {
	client, err := s.oauthClients.FindByClientID(ctx, clientID)
	if err != nil || client == nil || client.TenantID == nil {
		return nil
	}
	return client.TenantID
}

func (s *Server) tenantFromReturnTo(ctx context.Context, returnTo string) *uuid.UUID {
	clientID := clientIDFromReturnTo(returnTo)
	if clientID == "" {
		return nil
	}
	return s.tenantFromClientID(ctx, clientID)
}

func (s *Server) handlePageSignIn(w http.ResponseWriter, r *http.Request) {
	if user, _ := s.currentUser(r); user != nil {
		if ok, _ := s.users.IsPlatformOwner(r.Context(), user.ID); ok {
			http.Redirect(w, r, "/dashboard/", http.StatusFound)
			return
		}
	}
	s.pages.Render(w, "sign-in-content", s.pageData("title.sign_in", r))
}

func (s *Server) handlePageSignUp(w http.ResponseWriter, r *http.Request) {
	if user, _ := s.currentUser(r); user != nil {
		if ok, _ := s.users.IsPlatformOwner(r.Context(), user.ID); ok {
			http.Redirect(w, r, "/dashboard/", http.StatusFound)
			return
		}
	}
	s.pages.Render(w, "sign-up-content", s.pageData("title.sign_up", r))
}

func (s *Server) handlePageForgotPassword(w http.ResponseWriter, r *http.Request) {
	data := s.pageData("title.reset_password", r)
	data.ClientID = ""
	data.ReturnTo = ""
	clientID := strings.TrimSpace(r.URL.Query().Get("client_id"))
	returnTo := strings.TrimSpace(r.URL.Query().Get("return_to"))
	if clientID != "" {
		if client, _ := s.oauthClients.FindByClientID(r.Context(), clientID); client != nil && client.TenantID != nil {
			data.ClientID = clientID
			if tenant, _ := s.tenants.FindByID(r.Context(), *client.TenantID); tenant != nil {
				theme, _ := s.theme.GetByTenant(r.Context(), *client.TenantID)
				data.BrandName = appDisplayBrand(tenant, theme, client)
				applyTenantLocale(&data, r, tenant.DefaultLocale, "title.reset_password")
			}
			data.ReturnTo = resolveAppReturnTo(returnTo, s.cfg.AppURL, client)
		}
	}
	s.pages.Render(w, "forgot-password-content", data)
}

func (s *Server) handlePageResetPassword(w http.ResponseWriter, r *http.Request) {
	data := s.pageData("title.new_password", r)
	data.Token = r.URL.Query().Get("token")
	data.ClientID = ""
	data.ReturnTo = ""
	clientID := strings.TrimSpace(r.URL.Query().Get("client_id"))
	returnTo := strings.TrimSpace(r.URL.Query().Get("return_to"))
	if clientID != "" {
		if client, _ := s.oauthClients.FindByClientID(r.Context(), clientID); client != nil && client.TenantID != nil {
			data.ClientID = clientID
			if tenant, _ := s.tenants.FindByID(r.Context(), *client.TenantID); tenant != nil {
				theme, _ := s.theme.GetByTenant(r.Context(), *client.TenantID)
				data.BrandName = appDisplayBrand(tenant, theme, client)
				applyTenantLocale(&data, r, tenant.DefaultLocale, "title.new_password")
			}
			data.ReturnTo = resolveAppReturnTo(returnTo, s.cfg.AppURL, client)
		}
	}
	s.pages.Render(w, "reset-password-content", data)
}

func (s *Server) handlePageVerifyEmail(w http.ResponseWriter, r *http.Request) {
	data := s.pageData("title.verify_email", r)
	token := r.URL.Query().Get("token")
	clientID := strings.TrimSpace(r.URL.Query().Get("client_id"))
	returnTo := strings.TrimSpace(r.URL.Query().Get("return_to"))

	var appReturnTo string
	if clientID != "" {
		data.ClientID = clientID
		if client, _ := s.oauthClients.FindByClientID(r.Context(), clientID); client != nil && client.TenantID != nil {
			appReturnTo = resolveAppReturnTo(returnTo, s.cfg.AppURL, client)
			if tenant, _ := s.tenants.FindByID(r.Context(), *client.TenantID); tenant != nil {
				theme, _ := s.theme.GetByTenant(r.Context(), *client.TenantID)
				data.BrandName = appDisplayBrand(tenant, theme, client)
				applyTenantLocale(&data, r, tenant.DefaultLocale, "title.verify_email")
			}
		}
	}
	if appReturnTo != "" {
		data.SignInURL = appReturnTo
	}

	if token != "" {
		already, err := s.auth.VerifyEmail(r.Context(), token, clientIP(r))
		if err != nil {
			data.Error = friendlyPageError(data.Lang, "invalid_token")
		} else {
			if already {
				data.Message = i18n.T(data.Lang, "verify.already")
			} else {
				data.Message = i18n.T(data.Lang, "verify.success")
			}
			if appReturnTo != "" {
				dest, _ := appendEmailVerifiedQuery(appReturnTo)
				http.Redirect(w, r, dest, http.StatusFound)
				return
			}
		}
	}
	s.pages.Render(w, "verify-email-content", data)
}

func (s *Server) handlePageAccount(w http.ResponseWriter, r *http.Request) {
	user, _ := s.currentUser(r)
	if user == nil {
		http.Redirect(w, r, "/auth/sign-in?return_to=/auth/account", http.StatusFound)
		return
	}
	platform, _ := s.users.IsPlatformOwner(r.Context(), user.ID)
	if !platform {
		http.Error(w, "platform account required", http.StatusForbidden)
		return
	}
	data := s.pageData("title.security_center", r)
	data.HideFooter = true
	s.pages.Render(w, "account-content", data)
}

func appendEmailVerifiedQuery(returnTo string) (string, bool) {
	u, err := url.Parse(returnTo)
	if err != nil {
		return returnTo + "?email_verified=1", true
	}
	q := u.Query()
	q.Set("email_verified", "1")
	u.RawQuery = q.Encode()
	return u.String(), true
}
