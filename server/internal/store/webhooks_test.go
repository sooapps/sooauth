package store_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sooapps/sooauth/server/internal/migrate"
	"github.com/sooapps/sooauth/server/internal/store"
)

func TestWebhooksAreTenantScoped(t *testing.T) {
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

	owner := uuid.New()
	tenantA, tenantB := uuid.New(), uuid.New()
	_, err = db.Exec(ctx, `INSERT INTO users (id, email, platform_owner) VALUES ($1, $2, TRUE)`, owner, "webhook-test-"+owner.String()+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `INSERT INTO tenants (id, name, owner_user_id) VALUES ($1, 'Webhook A', $2), ($3, 'Webhook B', $2)`, tenantA, owner, tenantB)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = db.Exec(ctx, `DELETE FROM users WHERE id = $1`, owner) }()

	webhooks := store.NewWebhooks(db)
	endpoint, err := webhooks.Create(ctx, tenantA, "https://example.test/hook", "secret", []string{"user.login"})
	if err != nil {
		t.Fatal(err)
	}
	if endpoint.Secret != "secret" {
		t.Fatal("expected secret on create")
	}
	other, err := webhooks.FindByIDForTenant(ctx, tenantB, endpoint.ID)
	if err != nil {
		t.Fatal(err)
	}
	if other != nil {
		t.Fatal("endpoint crossed tenant boundary")
	}
	deleted, err := webhooks.Delete(ctx, tenantB, endpoint.ID)
	if err != nil {
		t.Fatal(err)
	}
	if deleted {
		t.Fatal("cross-tenant delete succeeded")
	}
}
