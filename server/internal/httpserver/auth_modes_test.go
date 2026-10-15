package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"

	"github.com/sooapps/sooauth/server/internal/store"
)

type modeFixture struct {
	s      *Server
	tenant *store.Tenant
	client *store.OAuthClient
}

func newModeFixture(t *testing.T, cfg store.AuthConfig) *modeFixture {
	t.Helper()
	s, db, _ := setupTestServer(t)
	t.Cleanup(db.Close)
	if s.ephemeral == nil {
		t.Skip("REDIS_URL not set")
	}
	ctx := context.Background()
	owner, _ := s.users.CreatePlatformUser(ctx, "mode-owner-"+uuid.New().String()[:8]+"@example.com", "Secret123!")
	acct, _ := s.accounts.FindOrCreateByOwner(ctx, owner.ID)
	tenant, _ := s.tenants.Create(ctx, acct.ID, owner.ID, "Mode App")
	if err := s.tenants.UpdateAuthConfigAndRegistrationSchema(ctx, tenant.ID, cfg.Normalize(), nil); err != nil {
		t.Fatal(err)
	}
	_ = s.tenants.UpdateSettings(ctx, tenant.ID, tenant.Name, false, tenant.VerifyDelivery(), tenant.PasswordPolicy(), "en")
	client, err := s.oauthClients.Create(ctx, store.OAuthClientInput{
		TenantID: tenant.ID, ClientID: "mode-" + uuid.New().String()[:8], Name: "Mode", RedirectURIs: []string{"https://app.local/cb"}, Public: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	tenant, _ = s.tenants.FindByID(ctx, tenant.ID)
	return &modeFixture{s: s, tenant: tenant, client: client}
}

func (f *modeFixture) post(t *testing.T, path string, body map[string]any) (int, map[string]any) {
	t.Helper()
	if _, ok := body["client_id"]; !ok && path != "/auth/mfa/verify" && path != "/auth/mfa/setup" {
		body["client_id"] = f.client.ClientID
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.s.Router().ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func (f *modeFixture) storedCode(t *testing.T, key string) string {
	t.Helper()
	var payload otpPayload
	ok, err := f.s.ephemeral.Get(context.Background(), key, &payload)
	if err != nil || !ok {
		t.Fatalf("no code stored under %s", key)
	}
	return payload.Code
}

func (f *modeFixture) createUser(t *testing.T, email string) *store.User {
	t.Helper()
	u, err := f.s.users.CreateAppUserFlexible(context.Background(), f.tenant.ID, email, "", "", "Secret123!x", nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.s.users.MarkEmailVerified(context.Background(), u.ID)
	_ = f.s.tenantUsers.LinkWithMethod(context.Background(), f.tenant.ID, u.ID, "email")
	return u
}

func TestOTPRejectedForPasswordOnlyProject(t *testing.T) {
	f := newModeFixture(t, store.AuthConfig{AllowedIdentifiers: []string{"email"}, PrimaryAuthMode: "password"})
	code, body := f.post(t, "/auth/otp/send", map[string]any{"identifier": "a@example.com"})
	if code != http.StatusForbidden || body["error"] != "otp_disabled" {
		t.Fatalf("want 403 otp_disabled, got %d %v", code, body)
	}
}

func TestPasswordRejectedForOTPOnlyProject(t *testing.T) {
	f := newModeFixture(t, store.AuthConfig{AllowedIdentifiers: []string{"email"}, PrimaryAuthMode: "otp"})
	email := "otp-only-" + uuid.New().String()[:8] + "@example.com"
	f.createUser(t, email)
	code, body := f.post(t, "/auth/sign-in", map[string]any{"identifier": email, "password": "Secret123!x"})
	if code != http.StatusForbidden || body["error"] != "password_login_disabled" {
		t.Fatalf("want 403 password_login_disabled, got %d %v", code, body)
	}
}

func TestBothModeAllowsPasswordOrCode(t *testing.T) {
	f := newModeFixture(t, store.AuthConfig{AllowedIdentifiers: []string{"email"}, PrimaryAuthMode: "both"})
	email := "both-" + uuid.New().String()[:8] + "@example.com"
	f.createUser(t, email)

	if code, body := f.post(t, "/auth/sign-in", map[string]any{"identifier": email, "password": "Secret123!x"}); code != http.StatusOK {
		t.Fatalf("password sign-in: %d %v", code, body)
	}
	if code, body := f.post(t, "/auth/otp/send", map[string]any{"identifier": email}); code != http.StatusOK {
		t.Fatalf("otp send: %d %v", code, body)
	}
	otp := f.storedCode(t, otpCacheKey(f.tenant.ID, email))
	if code, body := f.post(t, "/auth/otp/verify", map[string]any{"identifier": email, "code": otp}); code != http.StatusOK || body["access_token"] == nil {
		t.Fatalf("otp verify: %d %v", code, body)
	}
}

func TestOTPLocksAfterTooManyWrongCodes(t *testing.T) {
	f := newModeFixture(t, store.AuthConfig{AllowedIdentifiers: []string{"email"}, PrimaryAuthMode: "otp"})
	email := "lock-" + uuid.New().String()[:8] + "@example.com"
	if code, body := f.post(t, "/auth/otp/send", map[string]any{"identifier": email}); code != http.StatusOK {
		t.Fatalf("otp send: %d %v", code, body)
	}
	real := f.storedCode(t, otpCacheKey(f.tenant.ID, email))
	wrong := "000000"
	if real == wrong {
		wrong = "111111"
	}
	var last int
	for i := 0; i < otpMaxAttempts; i++ {
		last, _ = f.post(t, "/auth/otp/verify", map[string]any{"identifier": email, "code": wrong})
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("want 429 after %d wrong codes, got %d", otpMaxAttempts, last)
	}
	if code, _ := f.post(t, "/auth/otp/verify", map[string]any{"identifier": email, "code": real}); code == http.StatusOK {
		t.Fatal("code must be discarded after lockout")
	}
}

func TestOTPSignUpAsksForRequiredFields(t *testing.T) {
	f := newModeFixture(t, store.AuthConfig{AllowedIdentifiers: []string{"email"}, PrimaryAuthMode: "otp"})
	ctx := context.Background()
	schema := []store.RegistrationField{{ID: "company", Label: "Company", Type: "text", Required: true}}
	_ = f.s.tenants.UpdateAuthConfigAndRegistrationSchema(ctx, f.tenant.ID, f.tenant.AuthConfig, schema)

	email := "fields-" + uuid.New().String()[:8] + "@example.com"
	f.post(t, "/auth/otp/send", map[string]any{"identifier": email})
	otp := f.storedCode(t, otpCacheKey(f.tenant.ID, email))
	code, body := f.post(t, "/auth/otp/verify", map[string]any{"identifier": email, "code": otp})
	if code != http.StatusUnprocessableEntity || body["error"] != "registration_required" {
		t.Fatalf("want 422 registration_required, got %d %v", code, body)
	}
	code, body = f.post(t, "/auth/otp/verify", map[string]any{"identifier": email, "code": otp, "metadata": map[string]any{"company": "Acme"}})
	if code != http.StatusOK {
		t.Fatalf("retry with metadata: %d %v", code, body)
	}
}

func TestRequiredEmailMFA(t *testing.T) {
	f := newModeFixture(t, store.AuthConfig{
		AllowedIdentifiers: []string{"email"}, PrimaryAuthMode: "password",
		MFAPolicy: "required", MFAMethods: []string{"email_otp"},
	})
	email := "mfa-email-" + uuid.New().String()[:8] + "@example.com"
	f.createUser(t, email)

	code, body := f.post(t, "/auth/sign-in", map[string]any{"identifier": email, "password": "Secret123!x"})
	if code != http.StatusUnauthorized || body["error"] != "mfa_required" || body["method"] != "email_otp" {
		t.Fatalf("want mfa_required/email_otp, got %d %v", code, body)
	}
	challenge := body["challenge"].(string)
	if code, _ := f.post(t, "/auth/mfa/verify", map[string]any{"challenge": challenge, "code": "999999x", "method": "email_otp"}); code != http.StatusUnauthorized {
		t.Fatalf("wrong code accepted: %d", code)
	}
	otp := f.storedCode(t, mfaEmailKey(challenge))
	code, body = f.post(t, "/auth/mfa/verify", map[string]any{"challenge": challenge, "code": otp, "method": "email_otp"})
	if code != http.StatusOK || body["access_token"] == nil {
		t.Fatalf("mfa verify: %d %v", code, body)
	}
	if code, _ := f.post(t, "/auth/mfa/verify", map[string]any{"challenge": challenge, "code": otp, "method": "email_otp"}); code == http.StatusOK {
		t.Fatal("challenge reused")
	}
}

func TestRequiredTOTPEnrollsDuringSignIn(t *testing.T) {
	f := newModeFixture(t, store.AuthConfig{
		AllowedIdentifiers: []string{"email"}, PrimaryAuthMode: "password",
		MFAPolicy: "required", MFAMethods: []string{"totp"},
	})
	email := "mfa-totp-" + uuid.New().String()[:8] + "@example.com"
	f.createUser(t, email)

	_, body := f.post(t, "/auth/sign-in", map[string]any{"identifier": email, "password": "Secret123!x"})
	if body["method"] != "totp_setup" {
		t.Fatalf("want totp_setup, got %v", body)
	}
	challenge := body["challenge"].(string)
	code, setup := f.post(t, "/auth/mfa/setup", map[string]any{"challenge": challenge})
	if code != http.StatusOK || setup["secret"] == nil || setup["qr_code"] == nil {
		t.Fatalf("setup: %d %v", code, setup)
	}
	totpCode, _ := totp.GenerateCode(setup["secret"].(string), time.Now())
	code, body = f.post(t, "/auth/mfa/verify", map[string]any{"challenge": challenge, "code": totpCode, "method": "totp_setup"})
	if code != http.StatusOK || body["backup_codes"] == nil {
		t.Fatalf("enroll verify: %d %v", code, body)
	}

	// Next sign-in asks for the authenticator code.
	_, body = f.post(t, "/auth/sign-in", map[string]any{"identifier": email, "password": "Secret123!x"})
	if body["method"] != "totp" {
		t.Fatalf("want totp on second sign-in, got %v", body)
	}
}

func TestOptionalMFASkipsUnenrolledUsers(t *testing.T) {
	f := newModeFixture(t, store.AuthConfig{
		AllowedIdentifiers: []string{"email"}, PrimaryAuthMode: "password",
		MFAPolicy: "optional", MFAMethods: []string{"totp", "email_otp"},
	})
	email := "mfa-opt-" + uuid.New().String()[:8] + "@example.com"
	f.createUser(t, email)
	if code, body := f.post(t, "/auth/sign-in", map[string]any{"identifier": email, "password": "Secret123!x"}); code != http.StatusOK {
		t.Fatalf("optional MFA should not block unenrolled user: %d %v", code, body)
	}
}
