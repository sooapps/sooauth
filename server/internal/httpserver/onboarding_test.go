package httpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestOnboardingChecklist(t *testing.T) {
	f := newIsolationFixture(t)

	get := func() map[string]any {
		w := f.dashboardRequest(http.MethodGet, "/dashboard/api/onboarding", false)
		if w.Code != http.StatusOK {
			t.Fatalf("onboarding: %d %s", w.Code, w.Body.String())
		}
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return out
	}
	stepDone := func(state map[string]any, id string) bool {
		for _, raw := range state["steps"].([]any) {
			step := raw.(map[string]any)
			if step["id"] == id {
				return step["done"].(bool)
			}
		}
		t.Fatalf("step %s missing", id)
		return false
	}

	state := get()
	if state["dismissed"] != false {
		t.Fatal("new project should show the wizard")
	}
	if !stepDone(state, "test") {
		t.Fatal("fixture user already signed in, test step should be done")
	}

	req := httptest.NewRequest(http.MethodPut, "/dashboard/api/onboarding", bytes.NewBufferString(`{"app_type":"nextjs","complete":"integrate","dismissed":true}`))
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: f.adminTok.SessionToken})
	req.AddCookie(&http.Cookie{Name: csrfCookie, Value: f.adminTok.CSRFToken})
	req.Header.Set(csrfHeader, f.adminTok.CSRFToken)
	w := httptest.NewRecorder()
	f.s.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("put: %d %s", w.Code, w.Body.String())
	}
	state = get()
	if state["app_type"] != "nextjs" || state["dismissed"] != true || !stepDone(state, "project") || !stepDone(state, "integrate") {
		t.Fatalf("unexpected state %v", state)
	}
}

func TestAuthorizeAcceptsDashboardTestCallback(t *testing.T) {
	f := newIsolationFixture(t)
	q := url.Values{
		"client_id":             {f.client.ClientID},
		"redirect_uri":          {f.s.testRedirectURIs()[0]},
		"response_type":         {"code"},
		"scope":                 {"openid email"},
		"code_challenge":        {"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"},
		"code_challenge_method": {"S256"},
	}
	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+q.Encode(), nil)
	req.AddCookie(&http.Cookie{Name: appSessionCookie, Value: f.appTok.SessionToken})
	w := httptest.NewRecorder()
	f.s.Router().ServeHTTP(w, req)
	if loc := w.Header().Get("Location"); !strings.Contains(loc, "/dashboard/test-callback?code=") {
		t.Fatalf("test callback not accepted: %d %q", w.Code, loc)
	}
}

func TestAuthorizePromptLoginShowsSignIn(t *testing.T) {
	f := newIsolationFixture(t)
	q := url.Values{
		"client_id":             {f.client.ClientID},
		"redirect_uri":          {f.client.RedirectURIs[0]},
		"response_type":         {"code"},
		"scope":                 {"openid"},
		"code_challenge":        {"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"},
		"code_challenge_method": {"S256"},
		"prompt":                {"login"},
	}
	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+q.Encode(), nil)
	req.AddCookie(&http.Cookie{Name: appSessionCookie, Value: f.appTok.SessionToken})
	w := httptest.NewRecorder()
	f.s.Router().ServeHTTP(w, req)
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, "/auth/sign-in") {
		t.Fatalf("prompt=login must show sign-in, got %q", loc)
	}
	back, _ := url.Parse(loc)
	if strings.Contains(back.Query().Get("return_to"), "prompt=login") {
		t.Fatal("return_to must drop prompt=login to avoid a loop")
	}
}
