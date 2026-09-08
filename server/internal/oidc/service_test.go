package oidc_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sooapps/sooauth/server/internal/auth"
	appjwt "github.com/sooapps/sooauth/server/internal/crypto/jwt"
	"github.com/sooapps/sooauth/server/internal/config"
	"github.com/sooapps/sooauth/server/internal/crypto/signing"
	"github.com/sooapps/sooauth/server/internal/mail"
	"github.com/sooapps/sooauth/server/internal/migrate"
	"github.com/sooapps/sooauth/server/internal/oidc"
	"github.com/sooapps/sooauth/server/internal/ratelimit"
	"github.com/sooapps/sooauth/server/internal/store"
)

func testOIDC(t *testing.T) (*oidc.Service, *auth.Service, *pgxpool.Pool) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	if err := migrate.Up(dsn); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	key, err := signing.EnsureActiveKey(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	jwtIssuer, err := appjwt.NewIssuer(key, "http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{AppURL: "http://localhost:8080", Env: "development"}
	users := store.NewUsers(db)
	authSvc := auth.NewService(
		cfg,
		users,
		store.NewVerificationTokens(db),
		store.NewSessions(db),
		store.NewRefreshTokens(db),
		store.NewAudit(db),
		mail.New("", "test@sooauth.local"),
		ratelimit.New(nil),
		jwtIssuer,
	)
	oidcSvc := oidc.NewService(
		cfg.AppURL,
		store.NewOAuthClients(db),
		store.NewAuthorizationCodes(db),
		users,
		authSvc,
		jwtIssuer,
		limiterSafe(),
	)
	return oidcSvc, authSvc, db
}

func limiterSafe() *ratelimit.Limiter {
	return ratelimit.New(nil)
}

func pkcePair(t *testing.T) (challenge, verifier string) {
	t.Helper()
	verifier = "test-verifier-0123456789012345678901234567890"
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return challenge, verifier
}

func TestOIDCAuthorizationCodeFlow(t *testing.T) {
	oidcSvc, authSvc, db := testOIDC(t)
	defer db.Close()
	ctx := context.Background()

	email := "m5-oidc@sooauth.local"
	_, _ = db.Exec(ctx, `DELETE FROM users WHERE email = $1`, email)

	if err := authSvc.SignUp(ctx, email, "password-one-two", "127.0.0.1"); err != nil {
		t.Fatal(err)
	}
	user, _, err := store.NewUsers(db).FindByEmail(ctx, email)
	if err != nil || user == nil {
		t.Fatal("user not created")
	}
	if err := store.NewUsers(db).MarkEmailVerified(ctx, user.ID); err != nil {
		t.Fatal(err)
	}

	challenge, verifier := pkcePair(t)
	authResult, err := oidcSvc.Authorize(ctx, oidc.AuthorizeInput{
		ClientID:            "dev",
		RedirectURI:         "http://localhost:3000/callback",
		ResponseType:        "code",
		Scope:               "openid email profile",
		State:               "state-123",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		Nonce:               "nonce-abc",
	}, user)
	if err != nil {
		t.Fatal(err)
	}
	if authResult.RedirectURL == "" {
		t.Fatal("expected redirect url")
	}

	code := extractQueryParam(authResult.RedirectURL, "code")
	if code == "" {
		t.Fatalf("missing code in redirect: %s", authResult.RedirectURL)
	}

	tokenResult, err := oidcSvc.ExchangeCode(ctx, oidc.TokenInput{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  "http://localhost:3000/callback",
		ClientID:     "dev",
		CodeVerifier: verifier,
	}, "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if tokenResult.AccessToken == "" || tokenResult.IDToken == "" {
		t.Fatal("expected access and id tokens")
	}

	info, err := oidcSvc.UserInfo(ctx, tokenResult.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if info.Sub != user.ID.String() || info.Email != email {
		t.Fatalf("unexpected userinfo: %+v", info)
	}

	discovery := oidcSvc.Discovery()
	if discovery.AuthorizationEndpoint == "" || discovery.TokenEndpoint == "" {
		t.Fatal("discovery missing endpoints")
	}
}

func TestOIDCAuthorizeRequiresLogin(t *testing.T) {
	oidcSvc, _, db := testOIDC(t)
	defer db.Close()
	ctx := context.Background()

	challenge, _ := pkcePair(t)
	_, err := oidcSvc.Authorize(ctx, oidc.AuthorizeInput{
		ClientID:            "dev",
		RedirectURI:         "http://localhost:3000/callback",
		ResponseType:        "code",
		Scope:               "openid email",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	}, nil)
	if err != oidc.ErrLoginRequired {
		t.Fatalf("expected login required, got %v", err)
	}
}

func extractQueryParam(rawURL, key string) string {
	idx := len(rawURL)
	for i := 0; i < len(rawURL); i++ {
		if rawURL[i] == '?' {
			idx = i + 1
			break
		}
	}
	query := rawURL[idx:]
	for _, part := range splitAmp(query) {
		if len(part) > len(key)+1 && part[:len(key)+1] == key+"=" {
			return part[len(key)+1:]
		}
	}
	return ""
}

func splitAmp(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '&' {
			if i > start {
				out = append(out, s[start:i])
			}
			start = i + 1
		}
	}
	return out
}
