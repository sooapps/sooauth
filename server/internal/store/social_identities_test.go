package store_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sooapps/sooauth/server/internal/migrate"
	"github.com/sooapps/sooauth/server/internal/store"
)

func TestSocialIdentityScopeCollisionAndUnlinkIntegration(t *testing.T) {
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
	identities := store.NewSocialIdentities(db)
	user, err := users.CreatePlatformUser(ctx, "identity-owner@sooauth.local", "identity-pass")
	if err != nil {
		t.Fatal(err)
	}
	other, err := users.CreatePlatformUser(ctx, "identity-other@sooauth.local", "identity-pass")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Exec(ctx, `DELETE FROM users WHERE id IN ($1, $2)`, user.ID, other.ID) })

	if err := identities.Link(ctx, nil, user.ID, "google", "collision-sub", user.Email); err != nil {
		t.Fatal(err)
	}
	if err := identities.Link(ctx, nil, other.ID, "google", "collision-sub", other.Email); !errors.Is(err, store.ErrIdentityTaken) {
		t.Fatalf("expected identity collision, got %v", err)
	}
	linked, err := identities.ListForUser(ctx, nil, user.ID)
	if err != nil || len(linked) != 1 {
		t.Fatalf("list linked identities: %v, %d", err, len(linked))
	}
	if err := identities.UnlinkForUser(ctx, nil, user.ID, linked[0].ID); err != nil {
		t.Fatalf("password is a usable fallback: %v", err)
	}
	if err := identities.Link(ctx, nil, user.ID, "github", "last-sub", user.Email); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `DELETE FROM credentials WHERE user_id = $1`, user.ID); err != nil {
		t.Fatal(err)
	}
	linked, err = identities.ListForUser(ctx, nil, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := identities.UnlinkForUser(ctx, nil, user.ID, linked[0].ID); !errors.Is(err, store.ErrLastCredential) {
		t.Fatalf("expected last credential protection, got %v", err)
	}
}
