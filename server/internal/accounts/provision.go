package accounts

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/billing"
	"github.com/sooapps/sooauth/server/internal/store"
)

// ProvisionOwner ensures the account has a tenant + default OAuth client within plan limits.
func ProvisionOwner(
	ctx context.Context,
	accounts *store.Accounts,
	tenants *store.Tenants,
	clients *store.OAuthClients,
	themes *store.ThemeStore,
	ownerID uuid.UUID,
	email string,
) (*store.Account, *store.Tenant, *store.OAuthClient, error) {
	account, err := accounts.FindOrCreateByOwner(ctx, ownerID)
	if err != nil || account == nil {
		return nil, nil, nil, err
	}

	existing, err := tenants.ListByAccount(ctx, account.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(existing) > 0 {
		tenant := existing[0]
		client, err := clients.FindPrimaryByTenant(ctx, tenant.ID)
		return account, &tenant, client, err
	}

	count, err := tenants.CountByAccount(ctx, account.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	if count >= billing.MaxProjects(account.Plan) {
		return account, nil, nil, fmt.Errorf("project limit reached for %s plan", account.Plan)
	}

	name := projectNameFromEmail(email)
	tenant, err := tenants.Create(ctx, account.ID, ownerID, name)
	if err != nil {
		return nil, nil, nil, err
	}

	clientID, err := randomClientID()
	if err != nil {
		return nil, nil, nil, err
	}

	client, err := clients.Create(ctx, store.OAuthClientInput{
		TenantID:     tenant.ID,
		ClientID:     clientID,
		Name:         name + " app",
		RedirectURIs: []string{"http://localhost:3000/api/auth/callback/oidc"},
		Public:       true,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	_ = themes.EnsureDefault(ctx, tenant.ID)

	return account, tenant, client, nil
}

// CreateProject adds another project when the account plan allows it.
func CreateProject(
	ctx context.Context,
	accounts *store.Accounts,
	tenants *store.Tenants,
	clients *store.OAuthClients,
	themes *store.ThemeStore,
	ownerID uuid.UUID,
	name string,
) (*store.Tenant, *store.OAuthClient, error) {
	account, err := accounts.FindOrCreateByOwner(ctx, ownerID)
	if err != nil || account == nil {
		return nil, nil, err
	}

	count, err := tenants.CountByAccount(ctx, account.ID)
	if err != nil {
		return nil, nil, err
	}
	if count >= billing.MaxProjects(account.Plan) {
		return nil, nil, fmt.Errorf("project limit reached")
	}

	if name == "" {
		name = fmt.Sprintf("Project %d", count+1)
	}

	tenant, err := tenants.Create(ctx, account.ID, ownerID, name)
	if err != nil {
		return nil, nil, err
	}

	clientID, err := randomClientID()
	if err != nil {
		return nil, nil, err
	}

	client, err := clients.Create(ctx, store.OAuthClientInput{
		TenantID:     tenant.ID,
		ClientID:     clientID,
		Name:         name + " app",
		RedirectURIs: []string{"http://localhost:3000/api/auth/callback/oidc"},
		Public:       true,
	})
	if err != nil {
		return nil, nil, err
	}

	_ = themes.EnsureDefault(ctx, tenant.ID)
	return tenant, client, nil
}

func projectNameFromEmail(email string) string {
	local := strings.Split(email, "@")[0]
	local = strings.NewReplacer(".", " ", "_", " ", "+", " ").Replace(local)
	local = strings.TrimSpace(local)
	if local == "" {
		return "My project"
	}
	r := []rune(local)
	r[0] = []rune(strings.ToUpper(string(r[0])))[0]
	return string(r) + " project"
}

func randomClientID() (string, error) {
	raw := make([]byte, 6)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return fmt.Sprintf("app_%s", hex.EncodeToString(raw)), nil
}
