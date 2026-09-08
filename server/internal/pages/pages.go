package pages

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Renderer struct {
	templates *template.Template
}

type ViewData struct {
	Title           string
	BrandName       string
	LogoURL         string
	AccentColor     string
	Body            template.HTML
	Token           string
	ReturnTo        string
	TenantID        string
	ClientID        string
	SocialProviders []string
	Error           string
	Message         string
	AppURL          string
	SignInURL       string
}

func NewRenderer() (*Renderer, error) {
	funcs := template.FuncMap{
		"providerLabel": providerLabel,
	}
	tmpl, err := template.New("").Funcs(funcs).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Renderer{templates: tmpl}, nil
}

func providerLabel(name string) string {
	switch name {
	case "google":
		return "Google"
	case "github":
		return "GitHub"
	case "facebook":
		return "Facebook"
	case "x":
		return "X"
	default:
		return name
	}
}

func (r *Renderer) Render(w http.ResponseWriter, contentTemplate string, data ViewData) {
	var content bytes.Buffer
	if err := r.templates.ExecuteTemplate(&content, contentTemplate, data); err != nil {
		slog.Error("template render", "content", contentTemplate, "err", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	data.Body = template.HTML(content.String())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := r.templates.ExecuteTemplate(w, "layout", data); err != nil {
		slog.Error("template render", "layout", contentTemplate, "err", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func StaticHandler() http.Handler {
	sub, _ := fs.Sub(staticFS, "static")
	base := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		base.ServeHTTP(w, r)
	})
}
