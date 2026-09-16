package httpserver

import (
	"strings"

	"github.com/sooapps/sooauth/server/internal/i18n"
)

func friendlyPageError(lang, code string) string {
	switch code {
	case "invalid_credentials":
		return i18n.T(lang, "error.invalid_credentials")
	case "email_not_verified":
		return i18n.T(lang, "error.email_not_verified")
	case "social_callback_failed", "social_failed":
		return i18n.T(lang, "error.social_failed")
	case "identity_already_linked":
		return i18n.T(lang, "error.identity_already_linked")
	case "identity_link_failed":
		return i18n.T(lang, "error.identity_link_failed")
	case "app_account_dashboard":
		return i18n.T(lang, "error.app_account_dashboard")
	case "rate_limited":
		return i18n.T(lang, "error.rate_limited")
	case "invalid_token":
		return i18n.T(lang, "error.invalid_token")
	default:
		if code != "" {
			return i18n.T(lang, "error.generic")
		}
		return ""
	}
}

func friendlyPageMessage(raw string) string {
	if raw == "" {
		return ""
	}
	return strings.ReplaceAll(raw, "+", " ")
}

func friendlyAPIError(code string) string {
	if msg := friendlyPageError(i18n.LangEN, code); msg != "" {
		return msg
	}
	return i18n.T(i18n.LangEN, "error.generic")
}
