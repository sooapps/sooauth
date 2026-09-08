package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/skip2/go-qrcode"

	"github.com/sooapps/sooauth/server/internal/crypto/token"
	"github.com/sooapps/sooauth/server/internal/mfa"
	"github.com/sooapps/sooauth/server/internal/passwordpolicy"
	"github.com/sooapps/sooauth/server/internal/store"
)

func (s *Server) accountUser(w http.ResponseWriter, r *http.Request) (*store.User, bool) {
	user, ok := s.requirePlatformUser(w, r)
	if !ok {
		return nil, false
	}
	if !accountCSRFValid(w, r) {
		return nil, false
	}
	return user, true
}

func accountCSRFValid(w http.ResponseWriter, r *http.Request) bool {
	if cookieValue(r, sessionCookie) == "" {
		return true
	}
	if !csrfValid(r) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "csrf_invalid"})
		return false
	}
	return true
}

func (s *Server) handleAccountSessions(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requirePlatformUser(w, r)
	if !ok {
		return
	}
	sessions, err := s.sessions.ListActiveForUser(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "sessions_unavailable"})
		return
	}
	currentID := uuid.Nil
	if raw := cookieValue(r, sessionCookie); raw != "" {
		if current, _ := s.sessions.FindByTokenHash(r.Context(), token.Hash(raw)); current != nil {
			currentID = current.ID
		}
	}
	type sessionResponse struct {
		ID        uuid.UUID `json:"id"`
		IP        string    `json:"ip"`
		UserAgent string    `json:"user_agent"`
		CreatedAt string    `json:"created_at"`
		ExpiresAt string    `json:"expires_at"`
		Current   bool      `json:"current"`
	}
	result := make([]sessionResponse, 0, len(sessions))
	for _, sess := range sessions {
		result = append(result, sessionResponse{ID: sess.ID, IP: sess.IP, UserAgent: sess.UserAgent, CreatedAt: sess.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"), ExpiresAt: sess.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"), Current: sess.ID == currentID})
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": result})
}

func (s *Server) handleAccountRevokeSession(w http.ResponseWriter, r *http.Request) {
	user, ok := s.accountUser(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_session"})
		return
	}
	current := false
	if session := cookieValue(r, sessionCookie); session != "" {
		if active, _ := s.sessions.FindByTokenHash(r.Context(), token.Hash(session)); active != nil {
			current = active.ID == id
		}
	}
	if err := s.sessions.RevokeForUser(r.Context(), user.ID, id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "session_revoke_failed"})
		return
	}
	if current {
		clearAuthCookies(w, s.cfg.CookieSecure, s.cfg.CookieDomain)
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "session_revoked"})
}

func (s *Server) handleAccountRevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	user, ok := s.accountUser(w, r)
	if !ok {
		return
	}
	if err := s.sessions.RevokeAllForUser(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "sessions_revoke_failed"})
		return
	}
	if err := s.refresh.RevokeAllForUser(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "refresh_tokens_revoke_failed"})
		return
	}
	clearAuthCookies(w, s.cfg.CookieSecure, s.cfg.CookieDomain)
	writeJSON(w, http.StatusOK, map[string]string{"message": "sessions_revoked"})
}

func (s *Server) handleAccountPasskeys(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requirePlatformUser(w, r)
	if !ok {
		return
	}
	credentials, err := s.credentials.ListForUser(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "passkeys_unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"passkeys": credentials})
}

func (s *Server) handleAccountDeletePasskey(w http.ResponseWriter, r *http.Request) {
	user, ok := s.accountUser(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_passkey"})
		return
	}
	if err := s.credentials.DeleteForUser(r.Context(), user.ID, id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "passkey_delete_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "passkey_deleted"})
}

func (s *Server) handleAccountIdentities(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requirePlatformUser(w, r)
	if !ok {
		return
	}
	identities, err := s.identities.ListForUser(r.Context(), nil, user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "identities_unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"identities": identities})
}

func (s *Server) handleAccountUnlinkIdentity(w http.ResponseWriter, r *http.Request) {
	user, ok := s.accountUser(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_identity"})
		return
	}
	err = s.identities.UnlinkForUser(r.Context(), nil, user.ID, id)
	if errors.Is(err, store.ErrLastCredential) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "last_credential"})
		return
	}
	if errors.Is(err, store.ErrIdentityScope) || errors.Is(err, store.ErrIdentityTaken) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_identity"})
		return
	}
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "identity_not_found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "identity_unlink_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "identity_unlinked"})
}

func (s *Server) handleAccountMFA(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requirePlatformUser(w, r)
	if !ok {
		return
	}
	enabled, err := s.mfa.Enabled(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "mfa_status_unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": enabled})
}

func (s *Server) handleAccountMFAQRCode(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requirePlatformUser(w, r)
	if !ok {
		return
	}
	setupURL, err := s.mfa.PendingURL(r.Context(), user)
	if errors.Is(err, mfa.ErrNoPendingSetup) {
		http.Error(w, "no pending MFA setup", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "MFA setup unavailable", http.StatusInternalServerError)
		return
	}
	png, err := qrcode.Encode(setupURL, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "MFA QR code unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "image/png")
	_, _ = w.Write(png)
}

func (s *Server) handleAccountChangePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := s.accountUser(w, r)
	if !ok {
		return
	}
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	valid, err := s.users.VerifyPassword(r.Context(), user.ID, body.CurrentPassword)
	if err != nil || !valid {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_current_password"})
		return
	}
	if err := passwordpolicy.Default().Validate(body.NewPassword); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "weak_password"})
		return
	}
	if err := s.users.UpdatePassword(r.Context(), user.ID, body.NewPassword); err != nil {
		if errors.Is(err, passwordpolicy.ErrWeak) || strings.Contains(err.Error(), "at least") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "weak_password"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "password_change_failed"})
		return
	}
	if err := s.sessions.RevokeAllForUser(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "sessions_revoke_failed"})
		return
	}
	if err := s.refresh.RevokeAllForUser(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "refresh_tokens_revoke_failed"})
		return
	}
	clearAuthCookies(w, s.cfg.CookieSecure, s.cfg.CookieDomain)
	writeJSON(w, http.StatusOK, map[string]string{"message": "password_changed"})
}
