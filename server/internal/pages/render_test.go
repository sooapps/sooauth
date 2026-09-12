package pages

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRenderAllPages(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	data := ViewData{
		Title:       "Test",
		BrandName:   "sooauth",
		AccentColor: "#FF3B3B",
		AppURL:      "https://auth.sooauth.com",
	}
	pages := []string{
		"sign-in-content",
		"sign-up-content",
		"forgot-password-content",
		"reset-password-content",
		"verify-email-content",
		"account-content",
	}
	for _, name := range pages {
		w := httptest.NewRecorder()
		r.Render(w, name, data)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%q", name, w.Code, w.Body.String())
		}
	}
}
