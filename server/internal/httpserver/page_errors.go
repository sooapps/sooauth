package httpserver

import "strings"

func friendlyPageError(code string) string {
	switch code {
	case "invalid_credentials":
		return "Wrong email or password."
	case "email_not_verified":
		return "Check your inbox — verify your email before signing in."
	case "social_callback_failed", "social_failed":
		return "Google sign-in didn't finish. Close this tab and use the Google button again."
	case "identity_already_linked":
		return "That account is already connected to another platform account."
	case "identity_link_failed":
		return "The connected account could not be added. Try again."
	case "app_account_dashboard":
		return "That Google account is for your app users only. Sign in to the dashboard with your sooauth.com platform account (email + password), or use a private/incognito window to test app login."
	case "rate_limited":
		return "Too many attempts. Wait a minute and try again."
	case "invalid_token":
		return "This link expired or was already used."
	default:
		if code != "" {
			return "Something went wrong. Try again."
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
	if msg := friendlyPageError(code); msg != "" {
		return msg
	}
	return "Something went wrong. Try again."
}
