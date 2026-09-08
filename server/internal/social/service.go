package social

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	googleoauth "golang.org/x/oauth2/google"

	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/billing"
	"github.com/sooapps/sooauth/server/internal/config"
	"github.com/sooapps/sooauth/server/internal/ephemeral"
	"github.com/sooapps/sooauth/server/internal/providers"
	"github.com/sooapps/sooauth/server/internal/store"
)

var ErrProviderDisabled = errors.New("provider disabled")

type oauthState struct {
	Provider     string     `json:"provider"`
	ReturnTo     string     `json:"return_to"`
	CodeVerifier string     `json:"code_verifier"`
	TenantID     *uuid.UUID `json:"tenant_id,omitempty"`
	ClientID     string     `json:"client_id,omitempty"`
	AppReturn    bool       `json:"app_return,omitempty"`
	RedirectURL  string     `json:"redirect_url,omitempty"`
	LinkUserID   *uuid.UUID `json:"link_user_id,omitempty"`
}

type Service struct {
	cfg        config.Config
	users      *store.Users
	audit      *store.Audit
	auth       *auth.Service
	prov       *store.OAuthProviders
	tenants    *store.Tenants
	accounts   *store.Accounts
	ephem      *ephemeral.Store
	identities *store.SocialIdentities
}

func NewService(cfg config.Config, users *store.Users, audit *store.Audit, authSvc *auth.Service, prov *store.OAuthProviders, tenants *store.Tenants, accounts *store.Accounts, ephem *ephemeral.Store, identities *store.SocialIdentities) *Service {
	return &Service{cfg: cfg, users: users, audit: audit, auth: authSvc, prov: prov, tenants: tenants, accounts: accounts, ephem: ephem, identities: identities}
}

func (s *Service) Begin(ctx context.Context, provider, returnTo string, tenantID *uuid.UUID, clientID string, redirectURIs []string) (string, error) {
	return s.begin(ctx, provider, returnTo, tenantID, clientID, redirectURIs, nil)
}

func (s *Service) BeginLink(ctx context.Context, provider, returnTo string, userID uuid.UUID) (string, error) {
	return s.begin(ctx, provider, returnTo, nil, "", nil, &userID)
}

func (s *Service) begin(ctx context.Context, provider, returnTo string, tenantID *uuid.UUID, clientID string, redirectURIs []string, linkUserID *uuid.UUID) (string, error) {
	oauthCfg, err := s.providerConfig(ctx, provider, tenantID)
	if err != nil {
		return "", err
	}
	if oauthCfg == nil {
		return "", ErrProviderDisabled
	}

	state, err := randomToken(32)
	if err != nil {
		return "", err
	}
	verifier, challenge, err := pkcePair()
	if err != nil {
		return "", err
	}

	dest, appReturn := resolveReturnTo(returnTo, s.cfg.AppURL, redirectURIs, tenantID, clientID)

	if err := s.ephem.Set(ctx, "social:"+state, oauthState{
		Provider:     provider,
		ReturnTo:     dest,
		CodeVerifier: verifier,
		TenantID:     tenantID,
		ClientID:     clientID,
		AppReturn:    appReturn,
		RedirectURL:  oauthCfg.RedirectURL,
		LinkUserID:   linkUserID,
	}, 10*time.Minute); err != nil {
		return "", err
	}

	authURL := oauthCfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	return authURL, nil
}

func (s *Service) Complete(ctx context.Context, provider, state, code, ip, userAgent string, currentUserID *uuid.UUID) (*auth.TokenBundle, oauthState, error) {
	var saved oauthState
	ok, err := s.ephem.Get(ctx, "social:"+state, &saved)
	if err != nil || !ok || saved.Provider != provider {
		return nil, oauthState{}, auth.ErrInvalidToken
	}
	if saved.LinkUserID != nil && (currentUserID == nil || *currentUserID != *saved.LinkUserID) {
		return nil, saved, auth.ErrInvalidToken
	}
	_ = s.ephem.Delete(ctx, "social:"+state)

	oauthCfg, err := s.providerConfig(ctx, provider, saved.TenantID)
	if err != nil || oauthCfg == nil {
		return nil, oauthState{}, ErrProviderDisabled
	}
	if saved.RedirectURL != "" {
		oauthCfg.RedirectURL = saved.RedirectURL
	}

	token, err := oauthCfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", saved.CodeVerifier))
	if err != nil {
		return nil, oauthState{}, fmt.Errorf("token exchange: %w", err)
	}

	email, sub, err := s.fetchProfile(ctx, provider, token)
	if err != nil {
		return nil, oauthState{}, err
	}

	if saved.LinkUserID != nil {
		if err := s.identities.Link(ctx, nil, *saved.LinkUserID, provider, sub, email); err != nil {
			return nil, saved, err
		}
		user, err := s.users.FindByID(ctx, *saved.LinkUserID)
		if err != nil || user == nil {
			return nil, saved, err
		}
		return nil, saved, nil
	}

	user, created, err := s.users.FindOrCreateSocialUser(ctx, saved.TenantID, email, provider, sub)
	if err != nil || user == nil {
		return nil, oauthState{}, err
	}

	bundle, err := s.auth.SignInSocial(ctx, user, ip, userAgent)
	if err != nil {
		return nil, oauthState{}, err
	}

	uid := user.ID
	action := "social_signin"
	if created {
		action = "social_signup"
	}
	s.audit.Log(ctx, &uid, action, map[string]any{"provider": provider, "email": user.Email}, ip)
	return bundle, saved, nil
}

func (s *Service) EnabledForTenant(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	rows, err := s.prov.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, row := range rows {
		if !row.Enabled {
			continue
		}
		if !providers.LoginSupported(row.Provider) {
			continue
		}
		if _, _, ok := providers.Resolve(s.cfg, &row); ok {
			out = append(out, row.Provider)
		}
	}
	return out, nil
}

func (s *Service) providerConfig(ctx context.Context, provider string, tenantID *uuid.UUID) (*oauth2.Config, error) {
	var row *store.OAuthProviderConfig
	var err error

	if tenantID != nil {
		row, err = s.prov.FindByTenantProvider(ctx, *tenantID, provider)
		if err != nil {
			return nil, err
		}
		if row == nil || !row.Enabled {
			return nil, nil
		}
	} else {
		row = &store.OAuthProviderConfig{Provider: provider, Enabled: true, UsePlatform: true}
	}

	clientID, clientSecret, ok := providers.Resolve(s.cfg, row)
	if !ok {
		if tenantID == nil {
			clientID, clientSecret, ok = providers.PlatformCredentials(s.cfg, provider)
		}
		if !ok {
			return nil, nil
		}
	}

	if !providers.LoginSupported(provider) {
		return nil, nil
	}

	usingCustom := row != nil && row.ClientID != "" && !row.UsePlatform
	redirect := s.socialRedirectURL(ctx, provider, tenantID, usingCustom)
	switch provider {
	case "google":
		return &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirect,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     googleoauth.Endpoint,
		}, nil
	case "github":
		return &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirect,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     githuboauth.Endpoint,
		}, nil
	default:
		return nil, errors.New("unknown provider")
	}
}

func (s *Service) socialRedirectURL(ctx context.Context, provider string, tenantID *uuid.UUID, usingCustom bool) string {
	if tenantID == nil {
		return CallbackURL(s.cfg.AppURL, provider, nil, billing.PlanFree, usingCustom)
	}
	tenant, err := s.tenants.FindByID(ctx, *tenantID)
	if err != nil || tenant == nil {
		return CallbackURL(s.cfg.AppURL, provider, nil, billing.PlanFree, usingCustom)
	}
	plan := billing.PlanFree
	if account, err := s.accounts.FindByTenantID(ctx, *tenantID); err == nil && account != nil {
		plan = account.Plan
	}
	return CallbackURL(s.cfg.AppURL, provider, tenant, plan, usingCustom)
}

func (s *Service) fetchProfile(ctx context.Context, provider string, token *oauth2.Token) (email, sub string, err error) {
	switch provider {
	case "google":
		return googleProfile(ctx, token)
	case "github":
		return githubProfile(ctx, token)
	default:
		return "", "", errors.New("unknown provider")
	}
}

func googleProfile(ctx context.Context, token *oauth2.Token) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("google profile: %s", string(body))
	}
	var parsed struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", err
	}
	return parsed.Email, parsed.Sub, nil
}

func githubProfile(ctx context.Context, token *oauth2.Token) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("github profile: %s", string(body))
	}
	var user struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
		Login string `json:"login"`
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return "", "", err
	}

	email := user.Email
	if email == "" {
		email, err = githubPrimaryEmail(ctx, token)
		if err != nil {
			return "", "", err
		}
	}
	return email, fmt.Sprintf("%d", user.ID), nil
}

func githubPrimaryEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", errors.New("github email not available")
}

func pkcePair() (verifier, challenge string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

func randomToken(n int) (string, error) {
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func safeReturnTo(returnTo, appURL string) string {
	dest, _ := resolveReturnTo(returnTo, appURL, nil, nil, "")
	return dest
}

func ResolveReturnTo(returnTo, appURL string, redirectURIs []string, tenantID *uuid.UUID, clientID string) (string, bool) {
	return resolveReturnTo(returnTo, appURL, redirectURIs, tenantID, clientID)
}

func resolveReturnTo(returnTo, appURL string, redirectURIs []string, tenantID *uuid.UUID, clientID string) (string, bool) {
	defaultDest := strings.TrimRight(appURL, "/") + "/dashboard/"
	appScoped := tenantID != nil || strings.TrimSpace(clientID) != ""

	if appScoped {
		if dest, ok := appScopedReturnTo(returnTo, appURL, redirectURIs); ok {
			return dest, true
		}
		if len(redirectURIs) > 0 {
			return redirectURIs[0], true
		}
		if returnTo != "" {
			return returnTo, true
		}
		return strings.TrimRight(appURL, "/") + "/auth/sign-in?message=App+sign-in+complete", true
	}

	if returnTo == "" {
		return defaultDest, false
	}
	u, err := url.Parse(returnTo)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return defaultDest, false
	}
	app, _ := url.Parse(appURL)
	if app != nil && u.Host == app.Host {
		return returnTo, false
	}
	for _, raw := range redirectURIs {
		allowed, err := url.Parse(raw)
		if err != nil {
			continue
		}
		if u.Scheme == allowed.Scheme && u.Host == allowed.Host {
			return returnTo, true
		}
	}
	return defaultDest, false
}

func appScopedReturnTo(returnTo, appURL string, redirectURIs []string) (string, bool) {
	if returnTo == "" {
		return "", false
	}
	u, err := url.Parse(returnTo)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", false
	}
	if strings.Contains(u.Path, "/dashboard") {
		return "", false
	}
	app, _ := url.Parse(appURL)
	if app != nil && strings.EqualFold(u.Host, app.Host) {
		return returnTo, true
	}
	for _, raw := range redirectURIs {
		allowed, err := url.Parse(raw)
		if err != nil {
			continue
		}
		if strings.EqualFold(u.Scheme, allowed.Scheme) && strings.EqualFold(u.Host, allowed.Host) {
			return returnTo, true
		}
	}
	return "", false
}

func EnabledProviders(cfg config.Config) []string {
	var out []string
	for _, p := range store.AllSocialProviders {
		if providers.PlatformReady(cfg, p) && providers.LoginSupported(p) {
			out = append(out, p)
		}
	}
	return out
}
