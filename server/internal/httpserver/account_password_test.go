package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/config"
	appjwt "github.com/sooapps/sooauth/server/internal/crypto/jwt"
	"github.com/sooapps/sooauth/server/internal/crypto/signing"
	"github.com/sooapps/sooauth/server/internal/mail"
	"github.com/sooapps/sooauth/server/internal/migrate"
	"github.com/sooapps/sooauth/server/internal/ratelimit"
	"github.com/sooapps/sooauth/server/internal/store"
)

func TestUnauthenticatedPasswordEndpoints(t *testing.T) {
	s := &Server{
		cfg:  config.Config{AppURL: "http://localhost:8080"},
		auth: &auth.Service{},
	}
	router := s.Router()

	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{"GET", "/auth/account/password", ""},
		{"PUT", "/auth/account/password", `{"current_password":"old","new_password":"new"}`},
		{"POST", "/auth/account/password/set", `{"new_password":"new"}`},
	}

	for _, ep := range endpoints {
		var req *http.Request
		if ep.body != "" {
			req = httptest.NewRequest(ep.method, ep.path, bytes.NewBufferString(ep.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(ep.method, ep.path, nil)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s expected 401, got %d (body: %s)", ep.method, ep.path, w.Code, w.Body.String())
		}
	}
}

func setupTestServer(t *testing.T) (*Server, *pgxpool.Pool, *appjwt.Issuer) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	if err := migrate.Up(dsn); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	key, err := signing.EnsureActiveKey(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	jwtIssuer, err := appjwt.NewIssuer(key, "http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{AppURL: "http://localhost:8080", Env: "development"}
	users := store.NewUsers(db)
	sessions := store.NewSessions(db)
	refresh := store.NewRefreshTokens(db)
	audit := store.NewAudit(db)
	authSvc := auth.NewService(
		cfg,
		users,
		store.NewVerificationTokens(db),
		sessions,
		refresh,
		audit,
		mail.New("", "test@sooauth.local"),
		ratelimit.New(nil),
		jwtIssuer,
	)

	emailSettings := store.NewTenantEmailSettings(db, nil)
	authSvc.SetEmailSettingsStore(emailSettings)

	s := &Server{
		cfg:           cfg,
		db:            db,
		jwt:           jwtIssuer,
		auth:          authSvc,
		users:         users,
		sessions:      sessions,
		refresh:       refresh,
		audit:         audit,
		tenants:       store.NewTenants(db),
		oauthClients:  store.NewOAuthClients(db),
		theme:         store.NewThemeStore(db),
		accounts:      store.NewAccounts(db),
		emailSettings: emailSettings,
	}

	return s, db, jwtIssuer
}

func TestPasswordEndpointsIntegration(t *testing.T) {
	s, db, jwtIssuer := setupTestServer(t)
	defer db.Close()
	ctx := context.Background()
	router := s.Router()

	email := "social-user-" + uuid.New().String()[:8] + "@example.com"
	userID := uuid.New()
	_, err := db.Exec(ctx, `
		INSERT INTO users (id, email, email_verified_at, created_at, updated_at)
		VALUES ($1, $2, now(), now(), now())
	`, userID, email)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	tokenBundle, _, err := jwtIssuer.AccessToken(userID, email)
	if err != nil {
		t.Fatalf("failed to issue access token: %v", err)
	}
	bearer := "Bearer " + tokenBundle

	// 1. Check password status for user without password
	req := httptest.NewRequest("GET", "/auth/account/password", nil)
	req.Header.Set("Authorization", bearer)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for password status, got %d: %s", w.Code, w.Body.String())
	}
	var statusResp struct {
		HasPassword bool `json:"has_password"`
	}
	_ = json.NewDecoder(w.Body).Decode(&statusResp)
	if statusResp.HasPassword {
		t.Fatalf("expected has_password false for social user, got true")
	}

	// 2. Calling PUT /auth/account/password without password should fail
	req = httptest.NewRequest("PUT", "/auth/account/password", bytes.NewBufferString(`{"current_password":"foo","new_password":"valid-password-123"}`))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 no_password_set, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Calling POST /auth/account/password/set with weak password (< 8 chars) should fail
	req = httptest.NewRequest("POST", "/auth/account/password/set", bytes.NewBufferString(`{"new_password":"short"}`))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for weak password, got %d: %s", w.Code, w.Body.String())
	}

	// 4. Calling POST /auth/account/password/set with valid password should succeed
	pass1 := "first-valid-pass-123"
	setPayload, _ := json.Marshal(map[string]string{"new_password": pass1})
	req = httptest.NewRequest("POST", "/auth/account/password/set", bytes.NewBuffer(setPayload))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for set password, got %d: %s", w.Code, w.Body.String())
	}

	// 5. Check password status again - should be true now
	req = httptest.NewRequest("GET", "/auth/account/password", nil)
	req.Header.Set("Authorization", bearer)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	_ = json.NewDecoder(w.Body).Decode(&statusResp)
	if !statusResp.HasPassword {
		t.Fatalf("expected has_password true after setting password")
	}

	// 6. Calling POST /auth/account/password/set again should fail with password_already_set
	req = httptest.NewRequest("POST", "/auth/account/password/set", bytes.NewBuffer(setPayload))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 password_already_set, got %d: %s", w.Code, w.Body.String())
	}

	// 7. Calling PUT /auth/account/password with wrong current password should fail (401)
	wrongPayload, _ := json.Marshal(map[string]string{
		"current_password": "wrong-password",
		"new_password":     "second-valid-pass-456",
	})
	req = httptest.NewRequest("PUT", "/auth/account/password", bytes.NewBuffer(wrongPayload))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid current password, got %d: %s", w.Code, w.Body.String())
	}

	// 8. Calling PUT /auth/account/password with correct current password should succeed
	pass2 := "second-valid-pass-456"
	correctPayload, _ := json.Marshal(map[string]string{
		"current_password": pass1,
		"new_password":     pass2,
	})
	req = httptest.NewRequest("PUT", "/auth/account/password", bytes.NewBuffer(correctPayload))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for password change, got %d: %s", w.Code, w.Body.String())
	}

	// 9. Verify audit log has recorded both actions
	var setCount int
	_ = db.QueryRow(ctx, "SELECT count(*) FROM audit_log WHERE user_id = $1 AND action = 'password_set'", userID).Scan(&setCount)
	if setCount < 1 {
		t.Fatalf("expected at least 1 password_set audit log, got %d", setCount)
	}
	var changeCount int
	_ = db.QueryRow(ctx, "SELECT count(*) FROM audit_log WHERE user_id = $1 AND action = 'password_changed'", userID).Scan(&changeCount)
	if changeCount < 1 {
		t.Fatalf("expected at least 1 password_changed audit log, got %d", changeCount)
	}
}
