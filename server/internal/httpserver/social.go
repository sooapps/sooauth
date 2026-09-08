package httpserver

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/accounts"
	"github.com/sooapps/sooauth/server/internal/social"
	"github.com/sooapps/sooauth/server/internal/store"
)

func (s *Server) handleSocialStart(w http.ResponseWriter, r *http.Request) {
	provider := strings.TrimPrefix(r.URL.Path, "/auth/social/")
	provider = strings.TrimSuffix(provider, "/start")
	returnTo := r.URL.Query().Get("return_to")
	clientID := strings.TrimSpace(r.URL.Query().Get("client_id"))
	tenantID := parseTenantID(r.URL.Query().Get("tenant_id"))

	var redirectURIs []string
	if clientID != "" {
		client, err := s.oauthClients.FindByClientID(r.Context(), clientID)
		if err != nil || client == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_client"})
			return
		}
		redirectURIs = client.RedirectURIs
		// Provider credentials are tenant-scoped. Prefer the OAuth app's tenant over
		// the query param so a stale/wrong tenant_id cannot silently fall back to platform OAuth.
		if client.TenantID != nil {
			tenantID = client.TenantID
		}
	}

	url, err := s.social.Begin(r.Context(), provider, returnTo, tenantID, clientID, redirectURIs)
	if err == social.ErrProviderDisabled {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provider_disabled"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "social_start_failed"})
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (s *Server) handleAccountLinkIdentity(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requirePlatformUser(w, r)
	if !ok {
		return
	}
	provider := r.PathValue("provider")
	if provider != "google" && provider != "github" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provider_not_supported"})
		return
	}
	url, err := s.social.BeginLink(r.Context(), provider, "/auth/account?message=identity_linked", user.ID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "social_link_start_failed"})
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (s *Server) handleSocialCallback(w http.ResponseWriter, r *http.Request) {
	provider := strings.TrimPrefix(r.URL.Path, "/auth/social/")
	provider = strings.TrimSuffix(provider, "/callback")

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if errCode := r.URL.Query().Get("error"); errCode != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errCode})
		return
	}

	var currentUserID *uuid.UUID
	if user, _ := s.currentUser(r); user != nil {
		currentUserID = &user.ID
	}
	bundle, saved, err := s.social.Complete(r.Context(), provider, state, code, clientIP(r), r.UserAgent(), currentUserID)
	if err != nil {
		slog.Error("social callback", "provider", provider, "err", err)
		if saved.LinkUserID != nil {
			if errors.Is(err, store.ErrIdentityTaken) {
				http.Redirect(w, r, "/auth/account?error=identity_already_linked", http.StatusFound)
				return
			}
			http.Redirect(w, r, "/auth/account?error=identity_link_failed", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/auth/sign-in?error=social_failed", http.StatusFound)
		return
	}
	if saved.LinkUserID != nil {
		http.Redirect(w, r, saved.ReturnTo, http.StatusFound)
		return
	}

	if saved.TenantID != nil && bundle.User != nil {
		_ = s.tenantUsers.LinkWithMethod(r.Context(), *saved.TenantID, bundle.User.ID, provider)
	}

	// App OAuth (tenant widget / embedded login) must use widget exchange — never
	// overwrite the platform dashboard session cookie.
	if saved.TenantID != nil || strings.TrimSpace(saved.ClientID) != "" {
		saved.AppReturn = true
	}

	if saved.AppReturn {
		widgetCode, err := randomWidgetCode()
		if err != nil {
			http.Redirect(w, r, appendQuery(saved.ReturnTo, "error", "social_failed"), http.StatusFound)
			return
		}
		payload := widgetExchangePayload{
			ClientID:    saved.ClientID,
			AccessToken: bundle.AccessToken,
			Email:       bundle.User.Email,
			ExpiresIn:   bundle.ExpiresIn,
		}
		_ = s.ephemeral.Set(r.Context(), "widget:"+widgetCode, payload, 2*time.Minute)
		http.Redirect(w, r, appendQuery(saved.ReturnTo, "widget_code", widgetCode), http.StatusFound)
		return
	}

	setAuthCookies(w, s.cfg.CookieSecure, s.cfg.CookieDomain, bundle)
	target := saved.ReturnTo
	if target == "" || strings.Contains(target, "/auth/sign-in") {
		target = "/dashboard/"
	}
	if isDashboardDestination(target) && bundle.User != nil {
		platform, _ := s.users.IsPlatformOwner(r.Context(), bundle.User.ID)
		if !platform {
			http.Redirect(w, r, "/auth/sign-in?error=app_account_dashboard", http.StatusFound)
			return
		}
		_, _, _, err = accounts.ProvisionOwner(
			r.Context(), s.accounts, s.tenants, s.oauthClients, s.theme, bundle.User.ID, bundle.User.Email,
		)
		if err != nil {
			slog.Error("provision owner", "err", err)
		}
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func isDashboardDestination(target string) bool {
	return strings.Contains(target, "/dashboard")
}

func parseTenantID(raw string) *uuid.UUID {
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

func appendQuery(raw, key, val string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	q.Set(key, val)
	u.RawQuery = q.Encode()
	return u.String()
}

func randomWidgetCode() (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
