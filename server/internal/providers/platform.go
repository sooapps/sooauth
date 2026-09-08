package providers

import (
	"github.com/sooapps/sooauth/server/internal/config"
	"github.com/sooapps/sooauth/server/internal/store"
)

func PlatformCredentials(cfg config.Config, name string) (clientID, clientSecret string, ok bool) {
	switch name {
	case "google":
		return cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleClientID != ""
	case "github":
		return cfg.GithubClientID, cfg.GithubClientSecret, cfg.GithubClientID != ""
	case "facebook":
		return cfg.FacebookClientID, cfg.FacebookClientSecret, cfg.FacebookClientID != ""
	case "x":
		return cfg.XClientID, cfg.XClientSecret, cfg.XClientID != ""
	default:
		return "", "", false
	}
}

func PlatformReady(cfg config.Config, name string) bool {
	_, _, ok := PlatformCredentials(cfg, name)
	return ok
}

// Resolve picks tenant custom credentials or falls back to platform OAuth.
func Resolve(cfg config.Config, row *store.OAuthProviderConfig) (clientID, clientSecret string, ok bool) {
	if row != nil && row.Enabled {
		if !row.UsePlatform && row.ClientID != "" {
			return row.ClientID, row.ClientSecretEncrypted, true
		}
		if row.UsePlatform || row.ClientID == "" {
			return PlatformCredentials(cfg, row.Provider)
		}
	}
	return "", "", false
}

func LoginSupported(name string) bool {
	switch name {
	case "google", "github":
		return true
	default:
		return false
	}
}
