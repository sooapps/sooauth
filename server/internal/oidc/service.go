package oidc

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/auth"
	appjwt "github.com/sooapps/sooauth/server/internal/crypto/jwt"
	"github.com/sooapps/sooauth/server/internal/crypto/password"
	"github.com/sooapps/sooauth/server/internal/crypto/token"
	"github.com/sooapps/sooauth/server/internal/ratelimit"
	"github.com/sooapps/sooauth/server/internal/store"
)

var (
	ErrInvalidRequest      = errors.New("invalid request")
	ErrUnauthorizedClient  = errors.New("unauthorized client")
	ErrLoginRequired       = errors.New("login required")
	ErrInvalidGrant        = errors.New("invalid grant")
	ErrUnsupportedGrant    = errors.New("unsupported grant type")
	ErrInvalidScope        = errors.New("invalid scope")
)

type AuthorizeInput struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	Nonce               string
}

type AuthorizeResult struct {
	RedirectURL string
}

type TokenInput struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	CodeVerifier string
	RefreshToken string
}

type TokenResult struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int
	Scope        string
	TokenType    string
}

type Discovery struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	JWKSURI                           string   `json:"jwks_uri"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
}

type UserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified"`
}

type Service struct {
	issuerURL string
	clients   *store.OAuthClients
	codes     *store.AuthorizationCodes
	users     *store.Users
	auth      *auth.Service
	jwt       *appjwt.Issuer
	limit     *ratelimit.Limiter
}

func NewService(
	issuerURL string,
	clients *store.OAuthClients,
	codes *store.AuthorizationCodes,
	users *store.Users,
	authSvc *auth.Service,
	jwtIssuer *appjwt.Issuer,
	limiter *ratelimit.Limiter,
) *Service {
	return &Service{
		issuerURL: strings.TrimRight(issuerURL, "/"),
		clients:   clients,
		codes:     codes,
		users:     users,
		auth:      authSvc,
		jwt:       jwtIssuer,
		limit:     limiter,
	}
}

func (s *Service) Discovery() Discovery {
	base := s.issuerURL
	return Discovery{
		Issuer:                            base,
		AuthorizationEndpoint:             base + "/oauth/authorize",
		TokenEndpoint:                     base + "/oauth/token",
		UserinfoEndpoint:                  base + "/oauth/userinfo",
		JWKSURI:                           base + "/.well-known/jwks.json",
		ResponseTypesSupported:            []string{"code"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		ScopesSupported:                   []string{"openid", "email", "profile"},
		TokenEndpointAuthMethodsSupported: []string{"none", "client_secret_post"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token"},
	}
}

func (s *Service) Authorize(ctx context.Context, in AuthorizeInput, user *store.User) (*AuthorizeResult, error) {
	if in.ResponseType != "code" {
		return nil, ErrInvalidRequest
	}
	if user == nil {
		return nil, ErrLoginRequired
	}
	if in.ClientID == "" || in.RedirectURI == "" || in.Scope == "" {
		return nil, ErrInvalidRequest
	}
	if in.CodeChallenge == "" || in.CodeChallengeMethod != "S256" {
		return nil, ErrInvalidRequest
	}

	client, err := s.clients.FindByClientID(ctx, in.ClientID)
	if err != nil || client == nil {
		return nil, ErrUnauthorizedClient
	}
	if !client.AllowsRedirectURI(in.RedirectURI) {
		return nil, ErrInvalidRequest
	}
	if !client.AllowsScope(in.Scope) {
		return nil, ErrInvalidScope
	}
	if !strings.Contains(in.Scope, "openid") {
		return nil, ErrInvalidScope
	}

	if client.TenantID != nil {
		_ = s.users.EnsureTenant(ctx, user.ID, *client.TenantID)
	}

	plain, hash, err := token.Generate()
	if err != nil {
		return nil, err
	}
	if err := s.codes.Create(
		ctx,
		client.ClientID,
		user.ID,
		in.RedirectURI,
		in.Scope,
		in.Nonce,
		in.CodeChallenge,
		in.CodeChallengeMethod,
		hash,
	); err != nil {
		return nil, err
	}

	redirectURL, err := buildRedirect(in.RedirectURI, map[string]string{
		"code":  plain,
		"state": in.State,
	})
	if err != nil {
		return nil, err
	}
	return &AuthorizeResult{RedirectURL: redirectURL}, nil
}

func (s *Service) ExchangeCode(ctx context.Context, in TokenInput, ip string) (*TokenResult, error) {
	if ok, _ := s.limit.Allow(ctx, "token_ip", ip, 60, time.Minute); !ok {
		return nil, auth.ErrRateLimited
	}

	client, err := s.validateClient(ctx, in.ClientID, in.ClientSecret)
	if err != nil {
		return nil, err
	}
	if in.Code == "" || in.RedirectURI == "" || in.CodeVerifier == "" {
		return nil, ErrInvalidGrant
	}
	if !client.AllowsRedirectURI(in.RedirectURI) {
		return nil, ErrInvalidGrant
	}

	ac, err := s.codes.Consume(ctx, token.Hash(in.Code))
	if err != nil || ac == nil {
		return nil, ErrInvalidGrant
	}
	if ac.ClientID != client.ClientID || ac.RedirectURI != in.RedirectURI {
		return nil, ErrInvalidGrant
	}
	if !verifyPKCE(ac.CodeChallenge, ac.CodeChallengeMethod, in.CodeVerifier) {
		return nil, ErrInvalidGrant
	}

	user, err := s.users.FindByID(ctx, ac.UserID)
	if err != nil || user == nil {
		return nil, ErrInvalidGrant
	}

	bundle, err := s.auth.IssueOIDCTokens(ctx, user, ip)
	if err != nil {
		return nil, err
	}

	result := &TokenResult{
		AccessToken:  bundle.AccessToken,
		RefreshToken: bundle.RefreshToken,
		ExpiresIn:    bundle.ExpiresIn,
		Scope:        ac.Scope,
		TokenType:    "Bearer",
	}

	if strings.Contains(ac.Scope, "openid") {
		nonce := ""
		if ac.Nonce != nil {
			nonce = *ac.Nonce
		}
		idToken, _, err := s.jwt.IDToken(
			user.ID,
			user.Email,
			client.ClientID,
			nonce,
			user.EmailVerifiedAt != nil,
		)
		if err != nil {
			return nil, err
		}
		result.IDToken = idToken
	}

	return result, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken, ip string) (*TokenResult, error) {
	bundle, err := s.auth.Refresh(ctx, refreshToken, ip)
	if err != nil {
		return nil, err
	}
	return &TokenResult{
		AccessToken:  bundle.AccessToken,
		RefreshToken: bundle.RefreshToken,
		ExpiresIn:    bundle.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) UserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	user, err := s.auth.UserFromAccess(ctx, accessToken)
	if err != nil || user == nil {
		return nil, auth.ErrInvalidToken
	}
	return &UserInfo{
		Sub:           user.ID.String(),
		Email:         user.Email,
		EmailVerified: user.EmailVerifiedAt != nil,
	}, nil
}

func (s *Service) validateClient(ctx context.Context, clientID, clientSecret string) (*store.OAuthClient, error) {
	if clientID == "" {
		return nil, ErrUnauthorizedClient
	}
	client, err := s.clients.FindByClientID(ctx, clientID)
	if err != nil || client == nil {
		return nil, ErrUnauthorizedClient
	}
	if client.Public {
		return client, nil
	}
	if client.ClientSecretHash == nil || clientSecret == "" {
		return nil, ErrUnauthorizedClient
	}
	match, err := password.Verify(clientSecret, *client.ClientSecretHash)
	if err != nil || !match {
		return nil, ErrUnauthorizedClient
	}
	return client, nil
}

func verifyPKCE(challenge, method, verifier string) bool {
	if method != "S256" || verifier == "" {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return computed == challenge
}

func buildRedirect(base string, params map[string]string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func AuthorizeErrorRedirect(redirectURI, errCode, state, description string) (string, error) {
	params := map[string]string{
		"error":             errCode,
		"error_description": description,
		"state":             state,
	}
	return buildRedirect(redirectURI, params)
}

func ParseUUIDSubject(sub string) (uuid.UUID, error) {
	return uuid.Parse(sub)
}

func ErrorDescription(err error) string {
	switch {
	case errors.Is(err, ErrLoginRequired):
		return "login_required"
	case errors.Is(err, ErrInvalidScope):
		return "invalid_scope"
	case errors.Is(err, ErrUnauthorizedClient):
		return "unauthorized_client"
	case errors.Is(err, ErrInvalidGrant):
		return "invalid_grant"
	default:
		return "invalid_request"
	}
}

func OAuthErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrLoginRequired):
		return "login_required"
	case errors.Is(err, ErrInvalidScope):
		return "invalid_scope"
	case errors.Is(err, ErrUnauthorizedClient):
		return "unauthorized_client"
	case errors.Is(err, ErrInvalidGrant):
		return "invalid_grant"
	default:
		return "invalid_request"
	}
}

func (s *Service) IssuerURL() string {
	return s.issuerURL
}
