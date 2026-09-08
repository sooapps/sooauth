package httpserver

import (
	"net/http"

	"github.com/sooapps/sooauth/server/internal/widget"
)

func (s *Server) handleWidgetEmbedJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = w.Write(widget.EmbedJS)
}
