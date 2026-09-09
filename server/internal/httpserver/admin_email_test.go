package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/store"
)

func setupDashboardTestEnv(t *testing.T) (*Server, *store.Tenant, string) {
	t.Helper()
	s, db, jwtIssuer := setupTestServer(t)
	ctx := context.Background()

	email := "admin-" + uuid.New().String()[:8] + "@example.com"
	user, err := s.users.CreatePlatformUser(ctx, email, "SuperSecret123!")
	if err != nil {
		t.Fatalf("failed to create platform user: %v", err)
	}
	_ = s.users.MarkEmailVerified(ctx, user.ID)

	acct, err := s.accounts.FindOrCreateByOwner(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	tenant, err := s.tenants.Create(ctx, acct.ID, user.ID, "Email Test Tenant")
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	_, err = s.oauthClients.Create(ctx, store.OAuthClientInput{
		TenantID:     tenant.ID,
		ClientID:     "client-" + uuid.New().String()[:8],
		Name:         "Email Test Client",
		RedirectURIs: []string{"http://localhost:3000/callback"},
		Public:       true,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	tokenStr, _, err := jwtIssuer.AccessToken(user.ID, email)
	if err != nil {
		t.Fatalf("failed to issue access token: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return s, tenant, "Bearer " + tokenStr
}

func TestDashboardEmailEndpoints(t *testing.T) {
	s, tenant, bearer := setupDashboardTestEnv(t)
	router := s.Router()

	// 1. GET /dashboard/api/email initially returns defaults
	req := httptest.NewRequest("GET", "/dashboard/api/email", nil)
	req.Header.Set("Authorization", bearer)
	req.AddCookie(&http.Cookie{Name: projectCookie, Value: tenant.ID.String()})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var initial map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &initial); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if initial["enabled"] != false {
		t.Errorf("expected enabled: false, got %v", initial["enabled"])
	}
	if initial["provider"] != "smtp" {
		t.Errorf("expected provider: smtp, got %v", initial["provider"])
	}

	// 2. PUT /dashboard/api/email with SMTP config
	putSMTP := map[string]any{
		"enabled":       true,
		"provider":      "smtp",
		"from_name":     "Brand Support",
		"from_email":    "support@brand.example",
		"reply_to":      "help@brand.example",
		"smtp_host":     "smtp.mailgun.org",
		"smtp_port":     587,
		"smtp_user":     "postmaster@brand.example",
		"smtp_password": "mailgunpassword123",
		"smtp_tls_mode": "starttls",
	}
	putBody, _ := json.Marshal(putSMTP)
	req = httptest.NewRequest("PUT", "/dashboard/api/email", bytes.NewReader(putBody))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: projectCookie, Value: tenant.ID.String()})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT /dashboard/api/email failed with code %d: %s", w.Code, w.Body.String())
	}

	// 3. GET /dashboard/api/email returns updated SMTP config with masked password
	req = httptest.NewRequest("GET", "/dashboard/api/email", nil)
	req.Header.Set("Authorization", bearer)
	req.AddCookie(&http.Cookie{Name: projectCookie, Value: tenant.ID.String()})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /dashboard/api/email failed: %d", w.Code)
	}

	var smtpResp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &smtpResp)

	if smtpResp["enabled"] != true {
		t.Errorf("expected enabled: true, got %v", smtpResp["enabled"])
	}
	if smtpResp["from_name"] != "Brand Support" {
		t.Errorf("expected from_name: Brand Support, got %v", smtpResp["from_name"])
	}
	if smtpResp["from_email"] != "support@brand.example" {
		t.Errorf("expected from_email: support@brand.example, got %v", smtpResp["from_email"])
	}
	if smtpResp["smtp_has_password"] != true {
		t.Errorf("expected smtp_has_password: true, got %v", smtpResp["smtp_has_password"])
	}

	// 4. PUT /dashboard/api/email with Resend API key
	putResend := map[string]any{
		"enabled":    true,
		"provider":   "resend",
		"from_name":  "Brand Resend",
		"from_email": "auth@brand.example",
		"api_key":    "re_1234567890",
	}
	putBody, _ = json.Marshal(putResend)
	req = httptest.NewRequest("PUT", "/dashboard/api/email", bytes.NewReader(putBody))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: projectCookie, Value: tenant.ID.String()})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT resend failed: %d: %s", w.Code, w.Body.String())
	}

	// 5. Verify Resend config returned
	req = httptest.NewRequest("GET", "/dashboard/api/email", nil)
	req.Header.Set("Authorization", bearer)
	req.AddCookie(&http.Cookie{Name: projectCookie, Value: tenant.ID.String()})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resendResp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resendResp)

	if resendResp["provider"] != "resend" {
		t.Errorf("expected provider: resend, got %v", resendResp["provider"])
	}
	if resendResp["has_api_key"] != true {
		t.Errorf("expected has_api_key: true, got %v", resendResp["has_api_key"])
	}

	// 6. Test diagnostic POST /dashboard/api/email/test validation error on invalid provider
	testBad := map[string]any{
		"provider": "unknown-provider",
	}
	testBadBody, _ := json.Marshal(testBad)
	req = httptest.NewRequest("POST", "/dashboard/api/email/test", bytes.NewReader(testBadBody))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: projectCookie, Value: tenant.ID.String()})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown provider, got %d", w.Code)
	}

	// 7. Test diagnostic POST /dashboard/api/email/test with SMTP settings in dev
	testEmail := map[string]any{
		"to":         "recipient@brand.example",
		"provider":   "smtp",
		"from_name":  "Test Brand",
		"from_email": "auth@brand.example",
	}
	testBody, _ := json.Marshal(testEmail)
	req = httptest.NewRequest("POST", "/dashboard/api/email/test", bytes.NewReader(testBody))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: projectCookie, Value: tenant.ID.String()})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for diagnostic test email, got %d: %s", w.Code, w.Body.String())
	}
}
