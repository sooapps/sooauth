// Package adminui serves the React dashboard (admin/) from the Go binary for
// self-hosted installs. The build is copied into dist/ by
// `pnpm --filter @sooauth/admin build:embed` (the Dockerfile does this).
package adminui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var dist embed.FS

const notBuilt = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>sooauth dashboard</title>
<style>body{font-family:system-ui,sans-serif;max-width:560px;margin:15vh auto;padding:0 16px;line-height:1.6;color:#051b23}code{background:#f3f2f0;padding:2px 6px;border-radius:4px}</style></head>
<body><h1>Dashboard not built</h1><p>This binary was built without the dashboard. Build it with:</p>
<p><code>pnpm --filter @sooauth/admin build:embed</code></p><p>then rebuild the server, or use the Docker image which includes it.</p></body></html>`

// Handler serves files from the build and falls back to index.html for
// client-side routes such as /dashboard/p/<id>/users. Mount it with
// http.StripPrefix("/dashboard/", ...).
func Handler() http.Handler {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		return http.NotFoundHandler()
	}
	index, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(notBuilt))
		})
	}
	files := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name != "" && name != "index.html" {
			if info, err := fs.Stat(sub, name); err == nil && !info.IsDir() {
				if strings.HasPrefix(name, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				files.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(index)
	})
}
