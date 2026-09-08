package social

import (
	"testing"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/billing"
	"github.com/sooapps/sooauth/server/internal/store"
)

func TestCallbackURL(t *testing.T) {
	tenant := &store.Tenant{SocialCallbackOrigin: "https://soobrief.com"}
	got := CallbackURL("https://auth.sooauth.com", "google", tenant, billing.PlanPro, true)
	want := "https://soobrief.com/api/auth/google/callback"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestCallbackURLFreePlanUsesPlatform(t *testing.T) {
	tenant := &store.Tenant{SocialCallbackOrigin: "https://soobrief.com"}
	got := CallbackURL("https://auth.sooauth.com", "google", tenant, billing.PlanFree, true)
	want := "https://auth.sooauth.com/auth/social/google/callback"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestCallbackURLPlatformOAuth(t *testing.T) {
	tenant := &store.Tenant{SocialCallbackOrigin: "https://soobrief.com", ID: uuid.New()}
	got := CallbackURL("https://auth.sooauth.com", "google", tenant, billing.PlanPro, false)
	want := "https://auth.sooauth.com/auth/social/google/callback"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestValidateCallbackOrigin(t *testing.T) {
	uris := []string{"https://soobrief.com/login", "https://api.soobrief.com/callback"}
	if err := ValidateCallbackOrigin("https://api.soobrief.com", uris, true); err != nil {
		t.Fatalf("expected valid origin: %v", err)
	}
	if err := ValidateCallbackOrigin("https://evil.com", uris, true); err == nil {
		t.Fatal("expected mismatch error")
	}
}
