package httpserver

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/crypto/signing"
	"github.com/sooapps/sooauth/server/internal/oidc"
)

func (s *Server) handleOpenIDConfiguration(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.oidc.Discovery())
}

func (s *Server) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	in := oidc.AuthorizeInput{
		ClientID:            r.URL.Query().Get("client_id"),
		RedirectURI:         r.URL.Query().Get("redirect_uri"),
		ResponseType:        r.URL.Query().Get("response_type"),
		Scope:               r.URL.Query().Get("scope"),
		State:               r.URL.Query().Get("state"),
		CodeChallenge:       r.URL.Query().Get("code_challenge"),
		CodeChallengeMethod: r.URL.Query().Get("code_challenge_method"),
		Nonce:               r.URL.Query().Get("nonce"),
	}

	user, _ := s.currentUser(r)
	result, err := s.oidc.Authorize(r.Context(), in, user)
	if err == oidc.ErrLoginRequired {
		loginURL := strings.TrimRight(s.cfg.AppURL, "/") + "/auth/sign-in?return_to=" + url.QueryEscape(r.URL.String())
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}
	if err != nil {
		if in.RedirectURI != "" {
			if redirect, rerr := oidc.AuthorizeErrorRedirect(
				in.RedirectURI,
				oidc.OAuthErrorCode(err),
				in.State,
				oidc.ErrorDescription(err),
			); rerr == nil {
				http.Redirect(w, r, redirect, http.StatusFound)
				return
			}
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": oidc.OAuthErrorCode(err)})
		return
	}

	http.Redirect(w, r, result.RedirectURL, http.StatusFound)
}

func (s *Server) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	access := bearerToken(r)
	if access == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
		return
	}

	info, err := s.oidc.UserInfo(r.Context(), access)
	if err == auth.ErrInvalidToken {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
		return
	}

	writeJSON(w, http.StatusOK, info)
}

func (s *Server) handleOAuthToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_form"})
		return
	}

	grantType := r.FormValue("grant_type")
	switch grantType {
	case "refresh_token":
		s.handleRefreshTokenGrant(w, r)
	case "authorization_code":
		s.handleAuthorizationCodeGrant(w, r)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
	}
}

func (s *Server) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")
	if refreshToken == "" {
		refreshToken = cookieValue(r, refreshCookie)
	}
	if refreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token_required"})
		return
	}

	result, err := s.oidc.Refresh(r.Context(), refreshToken, clientIP(r))
	if err == auth.ErrRateLimited {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if err == auth.ErrInvalidToken {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_grant"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token_failed"})
		return
	}

	if cookieValue(r, refreshCookie) != "" {
		setRefreshCookies(w, s.cfg.CookieSecure, s.cfg.CookieDomain, &auth.TokenBundle{
			RefreshToken: result.RefreshToken,
		})
	}

	writeTokenResponse(w, result)
}

func (s *Server) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	in := oidc.TokenInput{
		GrantType:    "authorization_code",
		Code:         r.FormValue("code"),
		RedirectURI:  r.FormValue("redirect_uri"),
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
		CodeVerifier: r.FormValue("code_verifier"),
	}

	result, err := s.oidc.ExchangeCode(r.Context(), in, clientIP(r))
	if err == auth.ErrRateLimited {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
		return
	}
	if err == oidc.ErrUnauthorizedClient {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_client"})
		return
	}
	if err == oidc.ErrInvalidGrant || err == auth.ErrInvalidToken {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	writeTokenResponse(w, result)
}

func writeTokenResponse(w http.ResponseWriter, result *oidc.TokenResult) {
	body := map[string]any{
		"access_token": result.AccessToken,
		"expires_in":   result.ExpiresIn,
		"token_type":   result.TokenType,
	}
	if result.RefreshToken != "" {
		body["refresh_token"] = result.RefreshToken
	}
	if result.IDToken != "" {
		body["id_token"] = result.IDToken
	}
	if result.Scope != "" {
		body["scope"] = result.Scope
	}
	writeJSON(w, http.StatusOK, body)
}

func (s *Server) handleJWKS(w http.ResponseWriter, r *http.Request) {
	jwks, err := signing.JWKSFromDB(r.Context(), s.db)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "jwks_unavailable"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_ = json.NewEncoder(w).Encode(jwks)
}
