package httpserver

import (
	"net/http"
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
