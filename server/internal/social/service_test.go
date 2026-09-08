package social

import (
	"testing"

	"github.com/google/uuid"
)

func TestResolveReturnToAppScoped(t *testing.T) {
	appURL := "https://auth.sooauth.com"
	tenantID := uuid.MustParse("f6986fda-04f8-497e-a8a6-63f7765186d4")
	redirects := []string{"http://localhost:3001/api/auth/callback/social"}

	dest, appReturn := resolveReturnTo(
		"http://localhost:3001/login",
		appURL,
		redirects,
		&tenantID,
		"app_test",
	)
	if !appReturn || dest != "http://localhost:3001/login" {
		t.Fatalf("expected app return to localhost login, got %q appReturn=%v", dest, appReturn)
	}

	dest, appReturn = resolveReturnTo(
		"https://auth.sooauth.com/dashboard/",
		appURL,
		redirects,
		&tenantID,
		"app_test",
	)
	if !appReturn || dest != redirects[0] {
		t.Fatalf("dashboard return_to should fall back to redirect URI, got %q appReturn=%v", dest, appReturn)
	}

	dest, appReturn = resolveReturnTo("", appURL, redirects, &tenantID, "app_test")
	if !appReturn || dest != redirects[0] {
		t.Fatalf("empty return_to should use first redirect URI, got %q", dest)
	}
}

func TestResolveReturnToPlatformDashboard(t *testing.T) {
	appURL := "https://auth.sooauth.com"
	dest, appReturn := resolveReturnTo("https://auth.sooauth.com/dashboard/", appURL, nil, nil, "")
	if appReturn || dest != "https://auth.sooauth.com/dashboard/" {
		t.Fatalf("platform dashboard login should stay on dashboard flow, got %q appReturn=%v", dest, appReturn)
	}
}
