package adminui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var static embed.FS

func Handler() http.Handler {
	sub, err := fs.Sub(static, "static")
	if err != nil {
		return http.NotFoundHandler()
	}
	return http.FileServer(http.FS(sub))
}
