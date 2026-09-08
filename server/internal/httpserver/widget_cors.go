package httpserver

import "net/http"

func setPublicCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Sooauth-Embed")
}

func (s *Server) handlePublicCORS(w http.ResponseWriter, r *http.Request) {
	setPublicCORS(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func corsPreflight(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodOptions {
		return false
	}
	setPublicCORS(w, r)
	w.WriteHeader(http.StatusNoContent)
	return true
}
