package httpserver

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/pages"
	"github.com/sooapps/sooauth/server/internal/social"
)

func (s *Server) pageData(title string, r *http.Request) pages.ViewData {
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

	if tenantID != nil {
		tenantIDStr = tenantID.String()
		if ttheme, err := s.theme.GetByTenant(ctx, *tenantID); err == nil {
			theme = ttheme
		}
		socialProviders, _ = s.social.EnabledForTenant(ctx, *tenantID)
	} else if strings.Contains(returnTo, "/dashboard") {
		socialProviders = social.EnabledProviders(s.cfg)
	}

	return pages.ViewData{
		Title:           title,
		BrandName:       theme.BrandName,
		LogoURL:         theme.LogoURL,
		AccentColor:     theme.AccentColor,
		ReturnTo:        returnTo,
		TenantID:        tenantIDStr,
		ClientID:        clientID,
		SocialProviders: socialProviders,
		Error:           friendlyPageError(r.URL.Query().Get("error")),
		Message:         friendlyPageMessage(r.URL.Query().Get("message")),
		AppURL:          s.cfg.AppURL,
	}
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
	s.pages.Render(w, "sign-in-content", s.pageData("Sign in", r))
}

func (s *Server) handlePageSignUp(w http.ResponseWriter, r *http.Request) {
	if user, _ := s.currentUser(r); user != nil {
		if ok, _ := s.users.IsPlatformOwner(r.Context(), user.ID); ok {
			http.Redirect(w, r, "/dashboard/", http.StatusFound)
			return
		}
	}
	s.pages.Render(w, "sign-up-content", s.pageData("Create account", r))
}

func (s *Server) handlePageForgotPassword(w http.ResponseWriter, r *http.Request) {
	data := s.pageData("Reset password", r)
	clientID := strings.TrimSpace(r.URL.Query().Get("client_id"))
	returnTo := strings.TrimSpace(r.URL.Query().Get("return_to"))
	if clientID != "" {
		data.ClientID = clientID
		if client, _ := s.oauthClients.FindByClientID(r.Context(), clientID); client != nil && client.TenantID != nil {
			if tenant, _ := s.tenants.FindByID(r.Context(), *client.TenantID); tenant != nil {
				theme, _ := s.theme.GetByTenant(r.Context(), *client.TenantID)
				data.BrandName = appDisplayBrand(tenant, theme, client)
			}
			data.ReturnTo = resolveAppReturnTo(returnTo, s.cfg.AppURL, client)
		}
	}
	s.pages.Render(w, "forgot-password-content", data)
}

func (s *Server) handlePageResetPassword(w http.ResponseWriter, r *http.Request) {
	data := s.pageData("New password", r)
	data.Token = r.URL.Query().Get("token")
	clientID := strings.TrimSpace(r.URL.Query().Get("client_id"))
	returnTo := strings.TrimSpace(r.URL.Query().Get("return_to"))
	if clientID != "" {
		data.ClientID = clientID
		if client, _ := s.oauthClients.FindByClientID(r.Context(), clientID); client != nil && client.TenantID != nil {
			if tenant, _ := s.tenants.FindByID(r.Context(), *client.TenantID); tenant != nil {
				theme, _ := s.theme.GetByTenant(r.Context(), *client.TenantID)
				data.BrandName = appDisplayBrand(tenant, theme, client)
			}
			data.ReturnTo = resolveAppReturnTo(returnTo, s.cfg.AppURL, client)
		}
	}
	s.pages.Render(w, "reset-password-content", data)
}

func (s *Server) handlePageVerifyEmail(w http.ResponseWriter, r *http.Request) {
	data := s.pageData("Verify email", r)
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
			}
		}
	}
	if appReturnTo != "" {
		data.SignInURL = appReturnTo
	}

	if token != "" {
		already, err := s.auth.VerifyEmail(r.Context(), token, clientIP(r))
		if err != nil {
			data.Error = friendlyPageError("invalid_token")
		} else {
			if already {
				data.Message = "This link was already used. Your email is verified — you can sign in."
			} else {
				data.Message = "Email verified. You can sign in."
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
	s.pages.Render(w, "account-content", s.pageData("Security Center", r))
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
