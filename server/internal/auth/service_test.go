package auth_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/config"
	appjwt "github.com/sooapps/sooauth/server/internal/crypto/jwt"
	"github.com/sooapps/sooauth/server/internal/crypto/signing"
	"github.com/sooapps/sooauth/server/internal/crypto/token"
	"github.com/sooapps/sooauth/server/internal/mail"
	"github.com/sooapps/sooauth/server/internal/migrate"
	"github.com/sooapps/sooauth/server/internal/ratelimit"
	"github.com/sooapps/sooauth/server/internal/store"
)

func testService(t *testing.T) (*auth.Service, *pgxpool.Pool, *appjwt.Issuer) {
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
	svc := auth.NewService(
		cfg,
		store.NewUsers(db),
		store.NewVerificationTokens(db),
		store.NewSessions(db),
		store.NewRefreshTokens(db),
		store.NewAudit(db),
		mail.New("", "test@sooauth.local"),
		ratelimit.New(nil),
		jwtIssuer,
	)
	return svc, db, jwtIssuer
}

func TestSignUpVerifySignInResetFlow(t *testing.T) {
	svc, db, _ := testService(t)
	defer db.Close()
	ctx := context.Background()

	email := "m3-flow@sooauth.local"
	_, _ = db.Exec(ctx, `DELETE FROM users WHERE email = $1`, email)

	if err := svc.SignUp(ctx, email, "password-one-two", "127.0.0.1"); err != nil {
		t.Fatal(err)
	}

	user, _, err := store.NewUsers(db).FindByEmail(ctx, email)
	if err != nil || user == nil {
		t.Fatal("user not created")
	}
	if err := store.NewUsers(db).MarkEmailVerified(ctx, user.ID); err != nil {
		t.Fatal(err)
	}

	_, err = svc.SignIn(ctx, email, "wrong-pass", "127.0.0.1", "test")
	if err != auth.ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}

	bundle, err := svc.SignIn(ctx, email, "password-one-two", "127.0.0.1", "test")
	if err != nil || bundle.AccessToken == "" || bundle.User == nil {
		t.Fatalf("signin failed: %v", err)
	}

	me, err := svc.SessionUser(ctx, bundle.SessionToken)
	if err != nil || me == nil || me.Email != email {
		t.Fatalf("session user failed")
	}

	bearerUser, err := svc.UserFromAccess(ctx, bundle.AccessToken)
	if err != nil || bearerUser == nil || bearerUser.Email != email {
		t.Fatalf("bearer user failed: %v", err)
	}

	if err := svc.ForgotPassword(ctx, email, "127.0.0.1"); err != nil {
		t.Fatal(err)
	}

	if err := svc.ForgotPassword(ctx, "missing@sooauth.local", "127.0.0.1"); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshRotationAndReuseRevokesFamily(t *testing.T) {
	svc, db, jwtIssuer := testService(t)
	defer db.Close()
	ctx := context.Background()

	email := "m4-refresh@sooauth.local"
	_, _ = db.Exec(ctx, `DELETE FROM users WHERE email = $1`, email)

	if err := svc.SignUp(ctx, email, "password-one-two", "127.0.0.1"); err != nil {
		t.Fatal(err)
	}
	user, _, err := store.NewUsers(db).FindByEmail(ctx, email)
	if err != nil || user == nil {
		t.Fatal("user not created")
	}
	if err := store.NewUsers(db).MarkEmailVerified(ctx, user.ID); err != nil {
		t.Fatal(err)
	}

	initial, err := svc.SignIn(ctx, email, "password-one-two", "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	rt1 := initial.RefreshToken

	rotated, err := svc.Refresh(ctx, rt1, "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	rt2 := rotated.RefreshToken
	if rt2 == rt1 {
		t.Fatal("refresh token should rotate")
	}

	if _, err := jwtIssuer.ParseAccess(rotated.AccessToken); err != nil {
		t.Fatalf("access token invalid: %v", err)
	}

	_, err = svc.Refresh(ctx, rt1, "127.0.0.1")
	if err != auth.ErrInvalidToken {
		t.Fatalf("expected invalid token on reuse, got %v", err)
	}

	_, err = svc.Refresh(ctx, rt2, "127.0.0.1")
	if err != auth.ErrInvalidToken {
		t.Fatalf("expected family revoked, got %v", err)
	}
}

func TestJWKSValidatesAccessToken(t *testing.T) {
	svc, db, jwtIssuer := testService(t)
	defer db.Close()
	ctx := context.Background()

	email := "m4-jwks@sooauth.local"
	_, _ = db.Exec(ctx, `DELETE FROM users WHERE email = $1`, email)

	if err := svc.SignUp(ctx, email, "password-one-two", "127.0.0.1"); err != nil {
		t.Fatal(err)
	}
	user, _, err := store.NewUsers(db).FindByEmail(ctx, email)
	if err != nil || user == nil {
		t.Fatal("user not created")
	}
	if err := store.NewUsers(db).MarkEmailVerified(ctx, user.ID); err != nil {
		t.Fatal(err)
	}

	bundle, err := svc.SignIn(ctx, email, "password-one-two", "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}

	claims, err := jwtIssuer.ParseAccess(bundle.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Email != email {
		t.Fatalf("expected email %s, got %s", email, claims.Email)
	}

	jwks, err := signing.JWKSFromDB(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(jwks.Keys) == 0 {
		t.Fatal("expected at least one jwk")
	}
}

func TestPasswordResetOptsAndCodeFlow(t *testing.T) {
	svc, db, _ := testService(t)
	defer db.Close()
	ctx := context.Background()

	email := "reset-code-test@sooauth.local"
	_, _ = db.Exec(ctx, `DELETE FROM users WHERE email = $1`, email)

	if err := svc.SignUp(ctx, email, "old-password-123", "127.0.0.1"); err != nil {
		t.Fatal(err)
	}
	user, _, err := store.NewUsers(db).FindByEmail(ctx, email)
	if err != nil || user == nil {
		t.Fatal("user not found")
	}
	if err := store.NewUsers(db).MarkEmailVerified(ctx, user.ID); err != nil {
		t.Fatal(err)
	}

	// Request code delivery
	err = svc.ForgotPasswordWithOpts(ctx, auth.ForgotPasswordOpts{
		Email:    email,
		IP:       "127.0.0.1",
		Delivery: "code",
	})
	if err != nil {
		t.Fatalf("ForgotPasswordWithOpts failed: %v", err)
	}

	// Query token from database
	var codeHash string
	err = db.QueryRow(ctx, `
		SELECT token_hash FROM verification_tokens
		WHERE user_id = $1 AND type = 'password_reset_code' AND used_at IS NULL
	`, user.ID).Scan(&codeHash)
	if err != nil {
		t.Fatalf("code token not found in db: %v", err)
	}

	// Try with wrong code
	err = svc.ResetPasswordWithCode(ctx, email, "999999", "new-password-123", "127.0.0.1")
	if err == nil {
		t.Fatal("expected error with wrong code, got nil")
	}

	// Try with invalid password length (< 8)
	err = svc.ResetPasswordWithCode(ctx, email, "000000", "short", "127.0.0.1")
	if err == nil {
		t.Fatal("expected weak password error, got nil")
	}

	// Insert known verification code and test success
	knownCode := "842109"
	err = store.NewVerificationTokens(db).Create(ctx, user.ID, email, "password_reset_code", token.Hash(knownCode), time.Now().UTC().Add(15*time.Minute))
	if err != nil {
		t.Fatalf("failed to insert test verification token: %v", err)
	}
	newPass := "brand-new-password-123"
	if err := svc.ResetPasswordWithCode(ctx, email, knownCode, newPass, "127.0.0.1"); err != nil {
		t.Fatalf("expected ResetPasswordWithCode success, got: %v", err)
	}
	bundle, err := svc.SignIn(ctx, email, newPass, "127.0.0.1", "test")
	if err != nil || bundle.AccessToken == "" {
		t.Fatalf("sign-in with updated password failed: %v", err)
	}
}
