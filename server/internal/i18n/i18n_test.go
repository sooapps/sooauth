package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"tr":      LangTR,
		"TR":      LangTR,
		"tr-TR":   LangTR,
		"en":      LangEN,
		"en-US":   LangEN,
		"de":      LangEN,
		"":        LangEN,
		"  tr  ":  LangTR,
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Fatalf("Normalize(%q)=%q want %q", in, got, want)
		}
	}
}

func TestResolvePriority(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/auth/sign-in?lang=tr", nil)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if got := Resolve(req, "en"); got != LangTR {
		t.Fatalf("query should win: got %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/auth/sign-in", nil)
	req.Header.Set("Accept-Language", "tr-TR,tr;q=0.9")
	if got := Resolve(req, "en"); got != LangTR {
		t.Fatalf("Accept-Language should win: got %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/auth/sign-in", nil)
	if got := Resolve(req, "tr"); got != LangTR {
		t.Fatalf("tenant default should win: got %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/auth/sign-in", nil)
	if got := Resolve(req, ""); got != LangEN {
		t.Fatalf("fallback en: got %q", got)
	}
}

func TestTFallback(t *testing.T) {
	if got := T(LangTR, "title.sign_in"); got == "title.sign_in" || got == "" {
		t.Fatalf("expected Turkish title, got %q", got)
	}
	if got := T(LangEN, "title.sign_in"); got != "Sign in" {
		t.Fatalf("got %q", got)
	}
	if got := T(LangTR, "missing.key.xyz"); got != "missing.key.xyz" {
		t.Fatalf("missing key should return key, got %q", got)
	}
}
