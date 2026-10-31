package httpserver

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/config"
)

func TestProjectCandidates_headerBeatsEveryCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/dashboard/api/me", nil)
	req.Header.Set(projectHeader, "from-url")
	req.AddCookie(&http.Cookie{Name: projectCookie, Value: "stale-host-only"})
	req.AddCookie(&http.Cookie{Name: projectCookie, Value: "domain-wide"})

	got := projectCandidates(req)
	want := []string{"from-url", "stale-host-only", "domain-wide"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("projectCandidates() = %v, want %v", got, want)
	}
}

func TestSetActiveProject_clearsHostOnlyCopy(t *testing.T) {
	s := &Server{cfg: config.Config{CookieDomain: ".sooauth.com"}}
	w := httptest.NewRecorder()
	s.setActiveProject(w, uuid.New())

	cookies := w.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("got %d cookies, want the host-only delete and the domain cookie", len(cookies))
	}
	if cookies[0].MaxAge >= 0 || cookies[0].Domain != "" {
		t.Fatalf("first cookie should delete the host-only copy: %+v", cookies[0])
	}
	if !strings.HasSuffix(cookies[1].Domain, "sooauth.com") || cookies[1].Value == "" {
		t.Fatalf("second cookie should set the domain-wide value: %+v", cookies[1])
	}
}

func TestSetActiveProject_noDomainSetsOneCookie(t *testing.T) {
	s := &Server{cfg: config.Config{}}
	w := httptest.NewRecorder()
	s.setActiveProject(w, uuid.New())
	if n := len(w.Result().Cookies()); n != 1 {
		t.Fatalf("got %d cookies, want 1", n)
	}
}
