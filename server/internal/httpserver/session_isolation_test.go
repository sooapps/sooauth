package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/accounts"
	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/crypto/token"
	"github.com/sooapps/sooauth/server/internal/store"
)

type isolationFixture struct {
	s        *Server
	tenant   *store.Tenant
	client   *store.OAuthClient
	admin    *store.User
	appUser  *store.User
	adminTok *auth.TokenBundle
	appTok   *auth.TokenBundle
}

// newIsolationFixture builds the reported scenario: a dashboard owner and an end
// user of the owner's own project that share the same email address.
func newIsolationFixture(t *testing.T) *isolationFixture {
	t.Helper()
	s, db, _ := setupTestServer(t)
	t.Cleanup(db.Close)
	ctx := context.Background()

	email := "same-" + uuid.New().String()[:8] + "@example.com"
	admin, err := s.users.CreatePlatformUser(ctx, email, "Secret123!x")
	if err != nil {
		t.Fatal(err)
	}
	_ = s.users.MarkEmailVerified(ctx, admin.ID)
	_, tenant, client, err := accounts.ProvisionOwner(ctx, s.accounts, s.tenants, s.oauthClients, s.theme, admin.ID, email)
	if err != nil || tenant == nil || client == nil {
		t.Fatalf("provision: %v", err)
	}
	appUser, err := s.users.CreateAppUserFlexible(ctx, tenant.ID, email, "", "", "Secret123!x", nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = s.users.MarkEmailVerified(ctx, appUser.ID)
	if err := s.tenantUsers.LinkWithMethod(ctx, tenant.ID, appUser.ID, "email"); err != nil {
		t.Fatal(err)
	}

	adminTok, err := s.auth.SignIn(ctx, email, "Secret123!x", "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("admin sign-in: %v", err)
	}
	appTok, err := s.auth.SignInAppByIdentifierWithOptions(ctx, tenant.ID, email, "Secret123!x", "127.0.0.1", "test", true, auth.SignInOptions{RememberMe: true})
	if err != nil {
		t.Fatalf("app sign-in: %v", err)
	}
	if adminTok.User.ID == appTok.User.ID {
		t.Fatal("admin and app user must be different rows")
	}
	return &isolationFixture{s: s, tenant: tenant, client: client, admin: admin, appUser: appUser, adminTok: adminTok, appTok: appTok}
}

func (f *isolationFixture) dashboardRequest(method, path string, withCSRF bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: f.adminTok.SessionToken})
	req.AddCookie(&http.Cookie{Name: csrfCookie, Value: f.adminTok.CSRFToken})
	req.AddCookie(&http.Cookie{Name: projectCookie, Value: f.tenant.ID.String()})
	if withCSRF {
		req.Header.Set(csrfHeader, f.adminTok.CSRFToken)
	}
	w := httptest.NewRecorder()
	f.s.Router().ServeHTTP(w, req)
	return w
}

func TestDisableAppUserKeepsDashboardOwnerSignedIn(t *testing.T) {
	f := newIsolationFixture(t)
	ctx := context.Background()

	w := f.dashboardRequest(http.MethodPost, "/dashboard/api/users/"+f.appUser.ID.String()+"/disable", true)
	if w.Code != http.StatusOK {
		t.Fatalf("disable: %d %s", w.Code, w.Body.String())
	}

	if w := f.dashboardRequest(http.MethodGet, "/dashboard/api/me", false); w.Code != http.StatusOK {
		t.Fatalf("dashboard owner was signed out: %d %s", w.Code, w.Body.String())
	}
	if u, err := f.s.auth.SessionUser(ctx, f.appTok.SessionToken); err == nil && u != nil {
		t.Fatal("disabled user's session still works")
	}
	if _, err := f.s.auth.Refresh(ctx, f.appTok.RefreshToken, "127.0.0.2"); err == nil {
		t.Fatal("disabled user can still refresh tokens")
	}
	if _, err := f.s.auth.UserFromAccess(ctx, f.appTok.AccessToken); err == nil {
		t.Fatal("disabled user's access token still resolves")
	}
}

func TestDeleteAppUserRequiresCSRFAndKeepsOwner(t *testing.T) {
	f := newIsolationFixture(t)
	ctx := context.Background()
	path := "/dashboard/api/users/" + f.appUser.ID.String()

	if w := f.dashboardRequest(http.MethodDelete, path, false); w.Code != http.StatusForbidden {
		t.Fatalf("delete without csrf: want 403, got %d", w.Code)
	}
	if w := f.dashboardRequest(http.MethodDelete, path, true); w.Code != http.StatusOK {
		t.Fatalf("delete: %d %s", w.Code, w.Body.String())
	}
	if u, _ := f.s.users.FindByID(ctx, f.appUser.ID); u != nil {
		t.Fatal("app user row still exists")
	}
	if u, _ := f.s.users.FindByID(ctx, f.admin.ID); u == nil {
		t.Fatal("dashboard owner was deleted")
	}
	if w := f.dashboardRequest(http.MethodGet, "/dashboard/api/me", false); w.Code != http.StatusOK {
		t.Fatalf("dashboard owner was signed out: %d", w.Code)
	}
}

func TestAuthorizeIgnoresDashboardSessionForProjectClient(t *testing.T) {
	f := newIsolationFixture(t)
	ctx := context.Background()

	q := url.Values{
		"client_id":             {f.client.ClientID},
		"redirect_uri":          {f.client.RedirectURIs[0]},
		"response_type":         {"code"},
		"scope":                 {"openid email"},
		"state":                 {"s"},
		"code_challenge":        {"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"},
		"code_challenge_method": {"S256"},
	}
	authorize := func(cookie *http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+q.Encode(), nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		f.s.Router().ServeHTTP(w, req)
		return w
	}

	w := authorize(&http.Cookie{Name: sessionCookie, Value: f.adminTok.SessionToken})
	if loc := w.Header().Get("Location"); !strings.Contains(loc, "/auth/sign-in") {
		t.Fatalf("dashboard session must not sign in to a project app, got %d %q", w.Code, loc)
	}

	w = authorize(&http.Cookie{Name: appSessionCookie, Value: f.appTok.SessionToken})
	if loc := w.Header().Get("Location"); !strings.HasPrefix(loc, f.client.RedirectURIs[0]) || !strings.Contains(loc, "code=") {
		t.Fatalf("app session should get a code, got %d %q", w.Code, loc)
	}

	admin, _ := f.s.users.FindByID(ctx, f.admin.ID)
	if admin.TenantID != nil {
		t.Fatal("dashboard owner row was scoped to a project")
	}
}

func TestProjectSessionsExcludeOwner(t *testing.T) {
	f := newIsolationFixture(t)
	rows, err := f.s.sessions.ListActiveByTenant(context.Background(), f.tenant.ID, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.UserID == f.admin.ID {
			t.Fatal("dashboard owner session listed under project sessions")
		}
	}
	sess, err := f.s.sessions.FindByTokenHash(context.Background(), token.Hash(f.adminTok.SessionToken))
	if err != nil || sess == nil {
		t.Fatal("admin session not found")
	}
	ok, _ := f.s.sessions.RevokeByIDForTenant(context.Background(), sess.ID, f.tenant.ID)
	if ok {
		t.Fatal("project could revoke the owner's dashboard session")
	}
}
