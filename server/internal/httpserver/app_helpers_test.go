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
