package social

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sooapps/sooauth/server/internal/billing"
	"github.com/sooapps/sooauth/server/internal/store"
)

const appCallbackPath = "/api/auth/%s/callback"

// CallbackURL returns the OAuth redirect URI for a social provider.
func CallbackURL(appURL, provider string, tenant *store.Tenant, plan string, usingCustom bool) string {
	platform := strings.TrimRight(appURL, "/") + "/auth/social/" + provider + "/callback"
	if tenant == nil || !usingCustom || !billing.CanCustomSocialCallback(plan) {
		return platform
	}
	origin := strings.TrimSpace(tenant.SocialCallbackOrigin)
	if origin == "" {
		return platform
	}
	return strings.TrimRight(origin, "/") + fmt.Sprintf(appCallbackPath, provider)
}

// ValidateCallbackOrigin ensures the origin host matches a registered app redirect URI host.
func ValidateCallbackOrigin(origin string, redirectURIs []string, requireHTTPS bool) error {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return nil
	}
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid origin")
	}
	if u.Path != "" && u.Path != "/" {
		return fmt.Errorf("origin must not include a path")
	}
	if requireHTTPS && u.Scheme != "https" {
		return fmt.Errorf("origin must use https")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid origin scheme")
	}
	host := strings.ToLower(u.Hostname())
	for _, raw := range redirectURIs {
		allowed, err := url.Parse(strings.TrimSpace(raw))
		if err != nil || allowed.Host == "" {
			continue
		}
		if strings.EqualFold(allowed.Hostname(), host) {
			return nil
		}
	}
	return fmt.Errorf("origin host must match a redirect URI host")
}
