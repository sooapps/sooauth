package mail

import (
	"fmt"
	"html"
	"strings"
)

type VerificationDelivery struct {
	LinkURL string
	Code    string
}

func (t Transactional) VerificationDeliveryEmail(d VerificationDelivery, appScoped bool) (subject, plain, htmlBody string) {
	brand := t.brand()
	hasLink := strings.TrimSpace(d.LinkURL) != ""
	hasCode := strings.TrimSpace(d.Code) != ""

	subject = fmt.Sprintf("Confirm your %s account", brand)

	var lead, footer string
	if appScoped {
		footer = fmt.Sprintf("Authentication for %s is provided by sooauth. If you didn't sign up, you can safely ignore this email.", brand)
		switch {
		case hasLink && hasCode:
			lead = fmt.Sprintf("You signed up for %s. Confirm your email using the button below or enter the 6-digit code in the app.", brand)
		case hasCode:
			lead = fmt.Sprintf("You signed up for %s. Enter this 6-digit code in the app to verify your email.", brand)
		default:
			lead = fmt.Sprintf("You signed up for %s. Tap the button below to verify your email and finish creating your account.", brand)
		}
	} else {
		footer = fmt.Sprintf("If you didn't create a %s account, you can safely ignore this email.", brand)
		switch {
		case hasLink && hasCode:
			lead = fmt.Sprintf("Thanks for signing up for %s. Confirm your email with the link below or the 6-digit code.", brand)
		case hasCode:
			lead = fmt.Sprintf("Thanks for signing up for %s. Enter this 6-digit code to verify your email.", brand)
		default:
			lead = fmt.Sprintf("Thanks for signing up for %s. Tap the button below to verify your address and finish setting up your account.", brand)
		}
	}

	var plainParts []string
	plainParts = append(plainParts, fmt.Sprintf("Hi,\n\n%s", lead))
	if hasCode {
		plainParts = append(plainParts, fmt.Sprintf("\nYour verification code: %s", d.Code))
	}
	if hasLink {
		plainParts = append(plainParts, fmt.Sprintf("\nOr confirm with this link:\n%s", d.LinkURL))
	}
	plainParts = append(plainParts, "\nThis expires in 24 hours.")
	plain = strings.Join(plainParts, "") + fmt.Sprintf("\n\n— %s\n%s", brand, strings.TrimSuffix(t.AppURL, "/"))

	htmlBody = layoutVerificationEmail(verificationLayout{
		Brand:   brand,
		Title:   "Confirm your email",
		Lead:    lead,
		LinkURL: d.LinkURL,
		Code:    d.Code,
		Note:    "This expires in 24 hours.",
		Footer:  footer,
		AppURL:  strings.TrimSuffix(t.AppURL, "/"),
	})

	return subject, plain, htmlBody
}

type verificationLayout struct {
	Brand      string
	Title      string
	Lead       string
	ButtonText string
	LinkURL    string
	Code       string
	Note       string
	Footer     string
	AppURL     string
}

func layoutVerificationEmail(l verificationLayout) string {
	accent := "#E8FF3F"
	hasLink := strings.TrimSpace(l.LinkURL) != ""
	hasCode := strings.TrimSpace(l.Code) != ""
	btnText := strings.TrimSpace(l.ButtonText)
	if btnText == "" {
		btnText = "Confirm email"
	}

	var body strings.Builder
	body.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
</head>
<body style="margin:0;padding:0;background:#E3EFD7;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#051B23;">
  <div style="display:none;max-height:0;overflow:hidden;opacity:0;">%s</div>
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#E3EFD7;padding:32px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:520px;background:#FFFFFF;border-radius:16px;border:1px solid #051B23;overflow:hidden;">
          <tr>
            <td style="padding:28px 32px 8px;font-family:Georgia,'Times New Roman',serif;font-size:22px;font-weight:700;letter-spacing:-0.02em;color:#051B23;">%s</td>
          </tr>
          <tr>
            <td style="padding:8px 32px 0;font-size:20px;font-weight:600;line-height:1.35;color:#051B23;">%s</td>
          </tr>
          <tr>
            <td style="padding:16px 32px 0;font-size:15px;line-height:1.6;color:#5D6B70;">%s</td>
          </tr>`,
		html.EscapeString(l.Title),
		html.EscapeString(l.Lead),
		html.EscapeString(l.Brand),
		html.EscapeString(l.Title),
		html.EscapeString(l.Lead),
	))

	if hasCode {
		body.WriteString(fmt.Sprintf(`
          <tr>
            <td style="padding:24px 32px 8px;" align="center">
              <div style="display:inline-block;padding:16px 24px;border-radius:12px;background:#F5F5F5;font-size:32px;font-weight:700;letter-spacing:0.35em;color:#051B23;">%s</div>
            </td>
          </tr>`, html.EscapeString(l.Code)))
	}

	if hasLink {
		body.WriteString(fmt.Sprintf(`
          <tr>
            <td style="padding:28px 32px 8px;" align="center">
              <a href="%s" style="display:inline-block;background:%s;color:#051B23;text-decoration:none;font-size:15px;font-weight:600;padding:14px 28px;border-radius:10px;">%s</a>
            </td>
          </tr>
          <tr>
            <td style="padding:24px 32px 0;font-size:12px;line-height:1.6;color:#5D6B70;word-break:break-all;">Or copy this link:<br><a href="%s" style="color:#051B23;">%s</a></td>
          </tr>`,
			html.EscapeString(l.LinkURL),
			accent,
			html.EscapeString(btnText),
			html.EscapeString(l.LinkURL),
			html.EscapeString(l.LinkURL),
		))
	}

	body.WriteString(fmt.Sprintf(`
          <tr>
            <td style="padding:8px 32px 0;font-size:13px;line-height:1.5;color:#5D6B70;text-align:center;">%s</td>
          </tr>
          <tr>
            <td style="padding:24px 32px 28px;font-size:12px;line-height:1.5;color:#5D6B70;border-top:1px solid #C4C4C4;">%s</td>
          </tr>
        </table>
        <p style="margin:16px 0 0;font-size:12px;color:#5D6B70;"><a href="%s" style="color:#051B23;text-decoration:none;">%s</a></p>
      </td>
    </tr>
  </table>
</body>
</html>`,
		html.EscapeString(l.Note),
		html.EscapeString(l.Footer),
		html.EscapeString(l.AppURL),
		html.EscapeString(l.AppURL),
	))

	return body.String()
}

type PasswordResetDelivery struct {
	LinkURL string
	Code    string
}

func (t Transactional) PasswordResetDeliveryEmail(d PasswordResetDelivery, appScoped bool) (subject, plain, htmlBody string) {
	brand := t.brand()
	hasCode := strings.TrimSpace(d.Code) != ""

	subject = fmt.Sprintf("Reset your %s password", brand)

	var lead, footer string
	if appScoped {
		footer = fmt.Sprintf("Authentication for %s is provided by sooauth. If you didn't request a password reset, you can safely ignore this email.", brand)
	} else {
		footer = "If you didn't request a password reset, you can safely ignore this email."
	}

	if hasCode {
		lead = fmt.Sprintf("We received a request to reset your %s password. Enter this 6-digit code in the app to set a new password.", brand)
		plain = fmt.Sprintf(`Hi,

We received a request to reset your %s password. Enter this 6-digit code in the app:

Your password reset code: %s

This code expires in 15 minutes. If you didn't request this, ignore this email.

— %s
%s`, brand, d.Code, brand, strings.TrimSuffix(t.AppURL, "/"))

		htmlBody = layoutVerificationEmail(verificationLayout{
			Brand:   brand,
			Title:   "Reset your password",
			Lead:    lead,
			Code:    d.Code,
			Note:    "This code expires in 15 minutes.",
			Footer:  footer,
			AppURL:  strings.TrimSuffix(t.AppURL, "/"),
		})
	} else {
		lead = fmt.Sprintf("We received a request to reset your %s password. Tap the button below to choose a new password.", brand)
		plain = fmt.Sprintf(`Hi,

We received a request to reset your %s password. Open this link to choose a new password:

%s

This link expires in 1 hour. If you didn't request a reset, ignore this email.

— %s
%s`, brand, d.LinkURL, brand, strings.TrimSuffix(t.AppURL, "/"))

		htmlBody = layoutVerificationEmail(verificationLayout{
			Brand:      brand,
			Title:      "Reset your password",
			Lead:       lead,
			ButtonText: "Reset password",
			LinkURL:    d.LinkURL,
			Note:       "This link expires in 1 hour.",
			Footer:     footer,
			AppURL:     strings.TrimSuffix(t.AppURL, "/"),
		})
	}

	return subject, plain, htmlBody
}
