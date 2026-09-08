package httpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sooapps/sooauth/server/internal/auth"
)

func (s *Server) handlePasskeyRegisterBegin(w http.ResponseWriter, r *http.Request) {
	user, err := s.currentUser(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	options, err := s.passkeys.BeginRegistration(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "passkey_begin_failed"})
		return
	}
	writeJSON(w, http.StatusOK, options)
}

func (s *Server) handlePasskeyRegisterFinish(w http.ResponseWriter, r *http.Request) {
	user, err := s.currentUser(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	body, _ := io.ReadAll(r.Body)
	if err := s.passkeys.FinishRegistration(r.Context(), user.ID, bytes.NewReader(body)); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "passkey_finish_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "passkey_registered"})
}

func (s *Server) handlePasskeySignInBegin(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	options, state, err := s.passkeys.BeginLogin(r.Context(), payload.Email)
	if err == auth.ErrInvalidCredentials {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_credentials"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "passkey_begin_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"publicKey": options.Response, "state": state})
}

func (s *Server) handlePasskeySignInFinish(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		State    string          `json:"state"`
		Response json.RawMessage `json:"response"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	bundle, err := s.passkeys.FinishLogin(r.Context(), payload.State, bytes.NewReader(payload.Response), clientIP(r), r.UserAgent())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "passkey_finish_failed"})
		return
	}
	setAuthCookies(w, s.cfg.CookieSecure, s.cfg.CookieDomain, bundle)
	writeJSON(w, http.StatusOK, map[string]string{"message": "signed_in"})
}
