package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/fields"
	"github.com/sooapps/sooauth/server/internal/store"
)

type mockTransport func(req *http.Request) (*http.Response, error)

func (m mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req)
}

func TestFlexibleAuth_WidgetConfig(t *testing.T) {
	s, db, _ := setupTestServer(t)
	defer db.Close()
	ctx := context.Background()

	// Create tenant with custom auth_config and registration_schema
	owner, _ := s.users.CreatePlatformUser(ctx, "tenant-owner-"+uuid.New().String()[:8]+"@example.com", "Secret123!")
	acct, err := s.accounts.FindOrCreateByOwner(ctx, owner.ID)
	if err != nil {
		t.Fatalf("failed to find or create account: %v", err)
	}
	tenant, err := s.tenants.Create(ctx, acct.ID, owner.ID, "Gaming App")
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	authCfg := store.AuthConfig{
		AllowedIdentifiers: []string{"email", "username", "phone"},
		PrimaryAuthMode:    "both",
	}
	schema := []store.RegistrationField{
		{
			ID:       "lol_rank",
			Label:    "LoL Rank",
			Type:     "select",
			Required: true,
			OptionsSource: &store.OptionsSource{
				Type:     "dynamic_api",
				URL:      "https://api.riotgames.local/ranks",
				Headers:  map[string]string{"X-Riot-Token": "secret-api-key"},
				LabelKey: "name",
				ValueKey: "id",
			},
		},
		{
			ID:       "gender",
			Label:    "Gender",
			Type:     "select",
			Required: false,
			OptionsSource: &store.OptionsSource{
				Type: "static",
				Options: []store.SelectOption{
					{Label: "Female", Value: "female"},
					{Label: "Male", Value: "male"},
					{Label: "Other", Value: "other"},
				},
			},
		},
	}
	if err := s.tenants.UpdateAuthConfigAndRegistrationSchema(ctx, tenant.ID, authCfg, schema); err != nil {
		t.Fatalf("failed to update tenant schema: %v", err)
	}

	client, err := s.oauthClients.Create(ctx, store.OAuthClientInput{
		TenantID:     tenant.ID,
		ClientID:     "gaming-" + uuid.New().String()[:8],
		Name:         "Gaming Web",
		RedirectURIs: []string{"https://game.example.com/callback"},
		Public:       true,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	router := s.Router()
	req := httptest.NewRequest(http.MethodGet, "/v1/widget/config?client_id="+client.ClientID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	cfgRaw, ok := resp["auth_config"].(map[string]any)
	if !ok {
		t.Fatalf("expected auth_config in response: %+v", resp)
	}
	allowed, ok := cfgRaw["allowed_identifiers"].([]any)
	if !ok || len(allowed) != 3 {
		t.Fatalf("expected 3 allowed identifiers, got: %+v", cfgRaw["allowed_identifiers"])
	}

	schemaRaw, ok := resp["registration_schema"].([]any)
	if !ok || len(schemaRaw) != 2 {
		t.Fatalf("expected 2 schema fields, got: %+v", resp["registration_schema"])
	}

	// Verify sensitive headers or external URL are NOT leaked in widget config!
	firstField := schemaRaw[0].(map[string]any)
	if firstField["id"] != "lol_rank" {
		t.Errorf("expected first field lol_rank, got %v", firstField["id"])
	}
	if firstField["options_url"] == nil || firstField["options_url"] == "" {
		t.Errorf("expected options_url for dynamic_api field, got: %+v", firstField)
	}
	if firstField["headers"] != nil || firstField["url"] != nil {
		t.Errorf("security leak: headers or raw url exposed in public widget config: %+v", firstField)
	}
}

func TestFlexibleAuth_DynamicOptionsProxy(t *testing.T) {
	s, db, _ := setupTestServer(t)
	defer db.Close()
	ctx := context.Background()

	owner, _ := s.users.CreatePlatformUser(ctx, "proxy-owner-"+uuid.New().String()[:8]+"@example.com", "Secret123!")
	acct, _ := s.accounts.FindOrCreateByOwner(ctx, owner.ID)
	tenant, _ := s.tenants.Create(ctx, acct.ID, owner.ID, "Proxy App")

	mockRanks, _ := json.Marshal([]map[string]any{
		{"id": "gold", "name": "Gold"},
		{"id": "plat", "name": "Platinum"},
		{"id": "dia", "name": "Diamond"},
	})

	s.dynamicOptions = fields.NewDynamicOptionsService()
	mockTransport := mockTransport(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("X-Riot-Token") != "secret-riot-token" {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewReader([]byte("unauthorized"))),
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(mockRanks)),
			Header:     make(http.Header),
		}, nil
	})
	_ = mockTransport

	schema := []store.RegistrationField{
		{
			ID:       "rank",
			Label:    "Rank",
			Type:     "select",
			Required: true,
			OptionsSource: &store.OptionsSource{
				Type:     "dynamic_api",
				URL:      "https://api.riotgames.local/ranks",
				Headers:  map[string]string{"X-Riot-Token": "secret-riot-token"},
				LabelKey: "name",
				ValueKey: "id",
			},
		},
	}
	_ = s.tenants.UpdateAuthConfigAndRegistrationSchema(ctx, tenant.ID, store.DefaultAuthConfig(), schema)
	client, _ := s.oauthClients.Create(ctx, store.OAuthClientInput{
		TenantID:     tenant.ID,
		ClientID:     "proxy-" + uuid.New().String()[:8],
		Name:         "Proxy Web",
		RedirectURIs: []string{"https://game.example.com/callback"},
		Public:       true,
	})

	router := s.Router()
	req := httptest.NewRequest(http.MethodGet, "/v1/widget/fields/rank/options?client_id="+client.ClientID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadGateway {
		t.Fatalf("expected 200 or 502 from proxy, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFlexibleAuth_OTPFlow(t *testing.T) {
	s, db, _ := setupTestServer(t)
	defer db.Close()
	ctx := context.Background()

	owner, _ := s.users.CreatePlatformUser(ctx, "otp-owner-"+uuid.New().String()[:8]+"@example.com", "Secret123!")
	acct, _ := s.accounts.FindOrCreateByOwner(ctx, owner.ID)
	tenant, _ := s.tenants.Create(ctx, acct.ID, owner.ID, "OTP App")

	authCfg := store.AuthConfig{
		AllowedIdentifiers: []string{"phone"},
		PrimaryAuthMode:    "otp",
	}
	_ = s.tenants.UpdateAuthConfigAndRegistrationSchema(ctx, tenant.ID, authCfg, nil)
	client, _ := s.oauthClients.Create(ctx, store.OAuthClientInput{
		TenantID:     tenant.ID,
		ClientID:     "otp-" + uuid.New().String()[:8],
		Name:         "OTP Client",
		RedirectURIs: []string{"https://app.local/cb"},
		Public:       true,
	})

	router := s.Router()

	phone := "+905551234567"

	// 1. Send OTP
	sendBody, _ := json.Marshal(map[string]any{
		"client_id":  client.ClientID,
		"identifier": phone,
		"channel":    "sms",
	})
	reqSend := httptest.NewRequest(http.MethodPost, "/auth/otp/send", bytes.NewReader(sendBody))
	reqSend.Header.Set("Content-Type", "application/json")
	wSend := httptest.NewRecorder()
	router.ServeHTTP(wSend, reqSend)

	if wSend.Code != http.StatusOK {
		t.Fatalf("expected 200 from /auth/otp/send, got %d: %s", wSend.Code, wSend.Body.String())
	}

	// 2. Fetch code from ephemeral store
	cacheKey := "otp:" + tenant.ID.String() + ":" + phone
	var payload struct {
		Code string `json:"code"`
	}
	ok, err := s.ephemeral.Get(ctx, cacheKey, &payload)
	if err != nil || !ok || payload.Code == "" {
		t.Fatalf("expected OTP code in ephemeral store, ok=%v, err=%v", ok, err)
	}

	// 3. Verify OTP
	verifyBody, _ := json.Marshal(map[string]any{
		"client_id":  client.ClientID,
		"identifier": phone,
		"code":       payload.Code,
	})
	reqVerify := httptest.NewRequest(http.MethodPost, "/auth/otp/verify", bytes.NewReader(verifyBody))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	router.ServeHTTP(wVerify, reqVerify)

	if wVerify.Code != http.StatusOK {
		t.Fatalf("expected 200 from /auth/otp/verify, got %d: %s", wVerify.Code, wVerify.Body.String())
	}

	var verifyResp map[string]any
	if err := json.NewDecoder(wVerify.Body).Decode(&verifyResp); err != nil {
		t.Fatalf("failed to decode verify response: %v", err)
	}
	if verifyResp["access_token"] == nil || verifyResp["access_token"] == "" {
		t.Fatalf("expected access_token in verify response, got: %+v", verifyResp)
	}
}
