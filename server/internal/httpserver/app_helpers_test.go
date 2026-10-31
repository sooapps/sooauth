package httpserver

import (
	"testing"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/store"
)

func TestResolveAppReturnTo_prefersLoginOverFirstRedirectURI(t *testing.T) {
	tenantID := uuid.New()
	client := &store.OAuthClient{
		ClientID: "app_test",
		TenantID: &tenantID,
		RedirectURIs: []string{
			"https://soobrief.com/api/auth/oauth2/callback/sooauth",
			"https://soobrief.com/login",
			"https://soobrief.com/register",
		},
	}

	got := resolveAppReturnTo("", "https://auth.sooauth.com", client)
	want := "https://soobrief.com/login"
	if got != want {
		t.Fatalf("resolveAppReturnTo() = %q, want %q", got, want)
	}
}

func TestResolveAppReturnTo_honorsExplicitReturnTo(t *testing.T) {
	tenantID := uuid.New()
	client := &store.OAuthClient{
		ClientID: "app_test",
		TenantID: &tenantID,
		RedirectURIs: []string{
			"https://soobrief.com/api/auth/oauth2/callback/sooauth",
			"https://soobrief.com/login",
		},
	}

	got := resolveAppReturnTo("https://soobrief.com/register", "https://auth.sooauth.com", client)
	want := "https://soobrief.com/register"
	if got != want {
		t.Fatalf("resolveAppReturnTo() = %q, want %q", got, want)
	}
}

func TestResolveAppReturnTo_skipsOIDCCallback(t *testing.T) {
	tenantID := uuid.New()
	client := &store.OAuthClient{
		ClientID:     "app_test",
		TenantID:     &tenantID,
		RedirectURIs: []string{"https://myapp.com/api/auth/callback/oidc"},
	}

	if got := resolveAppReturnTo("", "https://auth.sooauth.com", client); got != "" {
		t.Fatalf("resolveAppReturnTo() = %q, want no destination", got)
	}
	if got := resolveAppReturnTo("https://evil.example/login", "https://auth.sooauth.com", client); got != "" {
		t.Fatalf("resolveAppReturnTo() = %q, want rejected host to be dropped", got)
	}
}

func TestResolveAppReturnTo_keepsHostedAuthorize(t *testing.T) {
	tenantID := uuid.New()
	client := &store.OAuthClient{
		ClientID:     "app_test",
		TenantID:     &tenantID,
		RedirectURIs: []string{"https://myapp.com/callback"},
	}

	own := "/oauth/authorize?client_id=app_test&state=abc"
	if got := resolveAppReturnTo(own, "https://auth.sooauth.com", client); got != own {
		t.Fatalf("resolveAppReturnTo() = %q, want %q", got, own)
	}
	if got := resolveAppReturnTo("/oauth/authorize?client_id=other", "https://auth.sooauth.com", client); got != "" {
		t.Fatalf("resolveAppReturnTo() = %q, want another client's flow to be dropped", got)
	}
	if got := resolveAppReturnTo("//evil.example/oauth/authorize?client_id=app_test", "https://auth.sooauth.com", client); got != "" {
		t.Fatalf("resolveAppReturnTo() = %q, want protocol-relative URL to be dropped", got)
	}
}
