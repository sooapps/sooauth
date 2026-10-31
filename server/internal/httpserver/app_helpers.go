package httpserver

import (
	"net/url"
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

// resolveAppReturnTo picks where an end user lands after an email link. An
// explicit return_to wins when it is allowed; a hosted sign-in in progress
// (/oauth/authorize on this server) is kept so the OIDC flow can continue.
func resolveAppReturnTo(returnTo, appURL string, client *store.OAuthClient) string {
	if client == nil {
		return ""
	}
	returnTo = strings.TrimSpace(returnTo)
	if isAuthorizeReturnTo(returnTo, client) {
		return returnTo
	}
	if dest, ok := social.AppReturnTo(returnTo, appURL, client.RedirectURIs); ok {
		return dest
	}
	return defaultAppReturnTo(client.RedirectURIs)
}

// isAuthorizeReturnTo reports whether returnTo is a local /oauth/authorize URL
// for this client.
func isAuthorizeReturnTo(returnTo string, client *store.OAuthClient) bool {
	if !strings.HasPrefix(returnTo, "/oauth/authorize?") {
		return false
	}
	u, err := url.Parse(returnTo)
	return err == nil && u.Query().Get("client_id") == client.ClientID
}

// defaultAppReturnTo only returns a page that can greet a returning user. OIDC
// callbacks are skipped: without a code they can only show an error.
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
	return ""
}
