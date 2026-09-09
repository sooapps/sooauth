package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sooapps/sooauth/server/internal/config"
)

func TestRouterPatternsDoNotPanic(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {})
	mux.Handle("GET /auth/static/", http.NotFoundHandler())
	mux.Handle("GET /dashboard/", http.NotFoundHandler())
	for _, method := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"} {
		mux.HandleFunc(method+" /admin/", redirectLegacyAdmin)
	}
}

func TestServerRouterBuilds(t *testing.T) {
	s := &Server{cfg: config.Config{AppURL: "http://localhost:8080"}}
	_ = s.Router()
}

func TestForgotPasswordAndResetPasswordCORS(t *testing.T) {
	s := &Server{cfg: config.Config{AppURL: "http://localhost:8080"}}
	router := s.Router()

	req := httptest.NewRequest("OPTIONS", "/auth/forgot-password", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected CORS header on OPTIONS /auth/forgot-password, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}

	req = httptest.NewRequest("OPTIONS", "/auth/reset-password", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected CORS header on OPTIONS /auth/reset-password, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}
