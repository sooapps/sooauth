package httpserver

import (
	"strings"

	"github.com/sooapps/sooauth/server/internal/social"
	"github.com/sooapps/sooauth/server/internal/store"
)

func appDisplayBrand(tenant *store.Tenant, theme store.Theme, client *store.OAuthClient) string {
	if tenant != nil && strings.TrimSpace(tenant.Name) != "" {
		return strings.TrimSpace(tenant.Name)
	}
	if name := strings.TrimSpace(theme.BrandName); name != "" && !strings.EqualFold(name, "sooauth") {
		return name
	}
	if client != nil && strings.TrimSpace(client.Name) != "" {
		return strings.TrimSpace(client.Name)
	}
	return "your app"
}

func resolveAppReturnTo(returnTo, appURL string, client *store.OAuthClient) string {
	if client == nil {
		return ""
	}
	if strings.TrimSpace(returnTo) != "" {
		if dest, ok := social.ResolveReturnTo(returnTo, appURL, client.RedirectURIs, client.TenantID, client.ClientID); ok && dest != "" {
			return dest
		}
	}
	return defaultAppReturnTo(client.RedirectURIs)
}

func defaultAppReturnTo(redirectURIs []string) string {
	lower := func(u string) string { return strings.ToLower(u) }
	for _, uri := range redirectURIs {
		if strings.Contains(lower(uri), "/login") {
			return uri
		}
	}
	for _, uri := range redirectURIs {
		if strings.Contains(lower(uri), "/register") {
			return uri
		}
	}
	if len(redirectURIs) > 0 {
		return redirectURIs[0]
	}
	return ""
}
