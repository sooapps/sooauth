package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/store"
)

func TestSignIn_RememberMe(t *testing.T) {
	s, db, _ := setupTestServer(t)
	defer db.Close()
	ctx := context.Background()
	router := s.Router()

	testCases := []struct {
		name                string
		payloadTemplate     string
		expectSessionMaxAge int
		expectRefreshMaxAge int
		expectTransient     bool
	}{
		{
			name:                "default (unspecified) defaults to persistent",
			payloadTemplate:     `{"email":"%s","password":"%s"}`,
			expectSessionMaxAge: 30 * 86400,
			expectRefreshMaxAge: 30 * 86400,
			expectTransient:     false,
		},
		{
			name:                "explicit true is persistent (30 days)",
			payloadTemplate:     `{"email":"%s","password":"%s","remember_me":true}`,
			expectSessionMaxAge: 30 * 86400,
			expectRefreshMaxAge: 30 * 86400,
			expectTransient:     false,
		},
		{
			name:                "explicit false is transient session cookie and 24h refresh",
			payloadTemplate:     `{"email":"%s","password":"%s","remember_me":false}`,
			expectSessionMaxAge: 0,
			expectRefreshMaxAge: 86400,
			expectTransient:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rawPass := "SecurePass123!"
			email := "remember-" + uuid.New().String()[:8] + "@example.com"
			user, err := s.users.CreatePlatformUser(ctx, email, rawPass)
			if err != nil {
				t.Fatalf("failed to insert user: %v", err)
			}
			if err := s.users.MarkEmailVerified(ctx, user.ID); err != nil {
				t.Fatalf("failed to mark email verified: %v", err)
			}

			body := fmt.Sprintf(tc.payloadTemplate, email, rawPass)
			req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("sign-in failed: code=%d, body=%s", w.Code, w.Body.String())
			}

			cookies := w.Result().Cookies()
			var sessionCookie, refreshCookie, csrfCookie *http.Cookie
			for _, c := range cookies {
				switch c.Name {
				case "sooauth_session":
					sessionCookie = c
				case "sooauth_refresh":
					refreshCookie = c
				case "sooauth_csrf":
					csrfCookie = c
				}
			}

			if sessionCookie == nil {
				t.Fatal("expected sooauth_session cookie")
			}
			if refreshCookie == nil {
				t.Fatal("expected sooauth_refresh cookie")
			}
			if csrfCookie == nil {
				t.Fatal("expected sooauth_csrf cookie")
			}

			if sessionCookie.MaxAge != tc.expectSessionMaxAge {
				t.Errorf("expected session cookie MaxAge=%d, got=%d", tc.expectSessionMaxAge, sessionCookie.MaxAge)
			}
			if refreshCookie.MaxAge != tc.expectRefreshMaxAge {
				t.Errorf("expected refresh cookie MaxAge=%d, got=%d", tc.expectRefreshMaxAge, refreshCookie.MaxAge)
			}
			if csrfCookie.MaxAge != tc.expectSessionMaxAge {
				t.Errorf("expected csrf cookie MaxAge=%d, got=%d", tc.expectSessionMaxAge, csrfCookie.MaxAge)
			}

			// Verify TTL stored in database
			var sessionExpiresAt time.Time
			err = db.QueryRow(ctx, `SELECT expires_at FROM sessions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, user.ID).Scan(&sessionExpiresAt)
			if err != nil {
				t.Fatalf("failed to query session expires_at: %v", err)
			}

			var refreshExpiresAt time.Time
			err = db.QueryRow(ctx, `SELECT expires_at FROM refresh_tokens WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, user.ID).Scan(&refreshExpiresAt)
			if err != nil {
				t.Fatalf("failed to query refresh token expires_at: %v", err)
			}

			now := time.Now().UTC()
			sessionRemaining := sessionExpiresAt.Sub(now)
			refreshRemaining := refreshExpiresAt.Sub(now)

			if tc.expectTransient {
				// Should be around 24 hours
				if sessionRemaining > 25*time.Hour || sessionRemaining < 23*time.Hour {
					t.Errorf("expected transient session TTL ~24h, got %v", sessionRemaining)
				}
				if refreshRemaining > 25*time.Hour || refreshRemaining < 23*time.Hour {
					t.Errorf("expected transient refresh TTL ~24h, got %v", refreshRemaining)
				}
			} else {
				// Should be around 30 days
				if sessionRemaining < 29*24*time.Hour {
					t.Errorf("expected persistent session TTL ~30d, got %v", sessionRemaining)
				}
				if refreshRemaining < 29*24*time.Hour {
					t.Errorf("expected persistent refresh TTL ~30d, got %v", refreshRemaining)
				}
			}
		})
	}
}

func TestWidgetConfig_RememberMe(t *testing.T) {
	s, db, _ := setupTestServer(t)
	defer db.Close()
	ctx := context.Background()
	router := s.Router()

	owner, err := s.users.CreatePlatformUser(ctx, "owner-"+uuid.New().String()[:8]+"@example.com", "SecretPass123!")
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	acct, err := s.accounts.FindOrCreateByOwner(ctx, owner.ID)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	tenant, err := s.tenants.Create(ctx, acct.ID, owner.ID, "Widget Tenant")
	if err != nil {
		t.Fatalf("failed to insert tenant: %v", err)
	}

	clientID := "client-" + uuid.New().String()[:8]
	client, err := s.oauthClients.Create(ctx, store.OAuthClientInput{
		TenantID:     tenant.ID,
		ClientID:     clientID,
		Name:         "Test App",
		RedirectURIs: []string{"http://localhost:3000/callback"},
		Public:       true,
	})
	if err != nil {
		t.Fatalf("failed to insert app: %v", err)
	}

	req := httptest.NewRequest("GET", "/v1/widget/config?client_id="+client.ClientID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if enabled, ok := res["remember_me_enabled"].(bool); !ok || !enabled {
		t.Errorf("expected remember_me_enabled: true, got %v", res["remember_me_enabled"])
	}
	if def, ok := res["remember_me_default"].(bool); !ok || !def {
		t.Errorf("expected remember_me_default: true, got %v", res["remember_me_default"])
	}
}
