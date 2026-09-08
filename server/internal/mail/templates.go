package mail

import (
	"fmt"
	"html"
	"strings"
)

type Transactional struct {
	BrandName string
	AppURL    string
}

func (t Transactional) brand() string {
	if strings.TrimSpace(t.BrandName) != "" {
		return t.BrandName
	}
	return "sooauth"
}

func (t Transactional) VerificationEmail(verifyURL string) (subject, plain, htmlBody string) {
	return t.verificationEmail(verifyURL, false)
}

func (t Transactional) AppVerificationEmail(verifyURL string) (subject, plain, htmlBody string) {
	return t.verificationEmail(verifyURL, true)
}

func (t Transactional) verificationEmail(verifyURL string, appScoped bool) (subject, plain, htmlBody string) {
	brand := t.brand()
	subject = fmt.Sprintf("Confirm your %s account", brand)

	var lead, footer, plainIntro string
	if appScoped {
		lead = fmt.Sprintf("You signed up for %s. Tap the button below to verify your email and finish creating your account.", brand)
		footer = fmt.Sprintf("Authentication for %s is provided by sooauth. If you didn't sign up, you can safely ignore this email.", brand)
		plainIntro = fmt.Sprintf("You signed up for %s. Confirm your email to finish creating your account:", brand)
	} else {
		lead = fmt.Sprintf("Thanks for signing up for %s. Tap the button below to verify your address and finish setting up your account.", brand)
		footer = fmt.Sprintf("If you didn't create a %s account, you can safely ignore this email.", brand)
		plainIntro = fmt.Sprintf("Thanks for signing up for %s. Confirm your email to finish creating your account:", brand)
	}

	plain = fmt.Sprintf(`Hi,

%s

%s

This link expires in 24 hours. If you didn't create an account, you can ignore this email.

— %s
%s`, plainIntro, verifyURL, brand, strings.TrimSuffix(t.AppURL, "/"))

	htmlBody = layoutEmail(emailLayout{
		Brand:  brand,
		Title:  "Confirm your email",
		Lead:   lead,
		Button: "Confirm email",
		URL:    verifyURL,
		Note:   "This link expires in 24 hours.",
		Footer: footer,
		AppURL: strings.TrimSuffix(t.AppURL, "/"),
	})

	return subject, plain, htmlBody
}

func (t Transactional) PasswordResetEmail(resetURL string) (subject, plain, htmlBody string) {
	brand := t.brand()
	subject = fmt.Sprintf("Reset your %s password", brand)

	plain = fmt.Sprintf(`Hi,

We received a request to reset your %s password. Open this link to choose a new password:

%s

This link expires in 1 hour. If you didn't request a reset, ignore this email.

— %s
%s`, brand, resetURL, brand, strings.TrimSuffix(t.AppURL, "/"))

	htmlBody = layoutEmail(emailLayout{
		Brand:  brand,
		Title:  "Reset your password",
		Lead:   fmt.Sprintf("We received a request to reset your %s password. Use the button below to choose a new one.", brand),
		Button: "Reset password",
		URL:    resetURL,
		Note:   "This link expires in 1 hour.",
		Footer: "If you didn't request this, you can safely ignore this email.",
		AppURL: strings.TrimSuffix(t.AppURL, "/"),
	})

	return subject, plain, htmlBody
}

type emailLayout struct {
	Brand  string
	Title  string
	Lead   string
	Button string
	URL    string
	Note   string
	Footer string
	AppURL string
}

func layoutEmail(l emailLayout) string {
	accent := "#E8FF3F"
	return fmt.Sprintf(`<!DOCTYPE html>
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
          </tr>
          <tr>
            <td style="padding:28px 32px 8px;" align="center">
              <a href="%s" style="display:inline-block;background:%s;color:#051B23;text-decoration:none;font-size:15px;font-weight:600;padding:14px 28px;border-radius:10px;">%s</a>
            </td>
          </tr>
          <tr>
            <td style="padding:8px 32px 0;font-size:13px;line-height:1.5;color:#5D6B70;text-align:center;">%s</td>
          </tr>
          <tr>
            <td style="padding:24px 32px 0;font-size:12px;line-height:1.6;color:#5D6B70;word-break:break-all;">Or copy this link:<br><a href="%s" style="color:#051B23;">%s</a></td>
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
		html.EscapeString(l.Title),
		html.EscapeString(l.Lead),
		html.EscapeString(l.Brand),
		html.EscapeString(l.Title),
		html.EscapeString(l.Lead),
		html.EscapeString(l.URL),
		accent,
		html.EscapeString(l.Button),
		html.EscapeString(l.Note),
		html.EscapeString(l.URL),
		html.EscapeString(l.URL),
		html.EscapeString(l.Footer),
		html.EscapeString(l.AppURL),
		html.EscapeString(l.AppURL),
	)
}
