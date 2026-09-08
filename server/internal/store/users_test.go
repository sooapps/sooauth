package store_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sooapps/sooauth/server/internal/crypto/password"
	"github.com/sooapps/sooauth/server/internal/migrate"
	"github.com/sooapps/sooauth/server/internal/store"
)

func TestCreatePlatformUserIntegration(t *testing.T) {
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
	defer db.Close()

	users := store.NewUsers(db)
	email := "m2-platform@sooauth.local"
	_, _ = db.Exec(ctx, `DELETE FROM users WHERE email = $1 AND platform_owner = TRUE`, email)

	user, err := users.CreatePlatformUser(ctx, email, "integration-pass")
	if err != nil {
		t.Fatal(err)
	}
	if user.Email != email {
		t.Fatalf("email mismatch")
	}

	found, hash, err := users.FindPlatformByEmail(ctx, email)
	if err != nil || found == nil {
		t.Fatalf("find failed: %v", err)
	}
	ok, err := password.Verify("integration-pass", hash)
	if err != nil || !ok {
		t.Fatalf("verify failed")
	}
}

func TestTenantScopedAppUsersIntegration(t *testing.T) {
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
	defer db.Close()

	users := store.NewUsers(db)
	email := "shared-app-user@sooauth.local"
	tenantA := uuid.New()
	tenantB := uuid.New()

	owner, err := users.CreatePlatformUser(ctx, "owner-"+tenantA.String()+"@sooauth.local", "owner-pass")
	if err != nil {
		t.Fatal(err)
	}

	accountID := uuid.New()
	_, err = db.Exec(ctx, `
		INSERT INTO accounts (id, owner_user_id, plan, created_at)
		VALUES ($1, $2, 'free', now())
		ON CONFLICT DO NOTHING
	`, accountID, owner.ID)
	if err != nil {
		t.Fatal(err)
	}

	for _, tenantID := range []uuid.UUID{tenantA, tenantB} {
		_, err = db.Exec(ctx, `
			INSERT INTO tenants (id, account_id, name, owner_user_id, created_at)
			VALUES ($1, $2, $3, $4, now())
		`, tenantID, accountID, tenantID.String(), owner.ID)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, _ = db.Exec(ctx, `DELETE FROM users WHERE email = $1 AND NOT platform_owner`, email)

	userA, err := users.CreateAppUser(ctx, tenantA, email, "password-a")
	if err != nil {
		t.Fatal(err)
	}
	userB, err := users.CreateAppUser(ctx, tenantB, email, "password-b")
	if err != nil {
		t.Fatal(err)
	}
	if userA.ID == userB.ID {
		t.Fatal("expected different user ids for same email in different tenants")
	}

	foundA, hashA, err := users.FindByEmailInTenant(ctx, tenantA, email)
	if err != nil || foundA == nil {
		t.Fatalf("tenant A lookup failed: %v", err)
	}
	foundB, hashB, err := users.FindByEmailInTenant(ctx, tenantB, email)
	if err != nil || foundB == nil {
		t.Fatalf("tenant B lookup failed: %v", err)
	}
	if ok, _ := password.Verify("password-a", hashA); !ok {
		t.Fatal("tenant A password mismatch")
	}
	if ok, _ := password.Verify("password-b", hashB); !ok {
		t.Fatal("tenant B password mismatch")
	}

	changed, err := users.SetDisabledForTenant(ctx, tenantA, userA.ID, true)
	if err != nil || !changed {
		t.Fatalf("disable tenant A user failed: %v", err)
	}
	foundA, _, err = users.FindByEmailInTenant(ctx, tenantA, email)
	if err != nil || foundA == nil || foundA.DisabledAt == nil {
		t.Fatalf("tenant A user was not disabled: %v", err)
	}
	foundB, _, err = users.FindByEmailInTenant(ctx, tenantB, email)
	if err != nil || foundB == nil || foundB.DisabledAt != nil {
		t.Fatalf("tenant B user was unexpectedly disabled: %v", err)
	}

	changed, err = users.SetDisabledForTenant(ctx, tenantB, userA.ID, true)
	if err != nil || changed {
		t.Fatal("cross-tenant disable should not affect a user")
	}
	changed, err = users.SetDisabledForTenant(ctx, tenantA, userA.ID, false)
	if err != nil || !changed {
		t.Fatalf("enable tenant A user failed: %v", err)
	}
}
