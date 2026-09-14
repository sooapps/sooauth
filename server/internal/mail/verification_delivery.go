package mail

import (
	"fmt"
	"html"
	"strings"

	"github.com/sooapps/sooauth/server/internal/i18n"
)

type VerificationDelivery struct {
	LinkURL string
	Code    string
}

func (t Transactional) lang() string {
	return i18n.Normalize(t.Lang)
}

func (t Transactional) VerificationDeliveryEmail(d VerificationDelivery, appScoped bool) (subject, plain, htmlBody string) {
	brand := t.brand()
	lang := t.lang()
	hasLink := strings.TrimSpace(d.LinkURL) != ""
	hasCode := strings.TrimSpace(d.Code) != ""

	subject = fmt.Sprintf(i18n.T(lang, "mail.verify.subject"), brand)

	var lead, footer string
	if appScoped {
		footer = fmt.Sprintf(i18n.T(lang, "mail.verify.app.footer"), brand)
		switch {
		case hasLink && hasCode:
			lead = fmt.Sprintf(i18n.T(lang, "mail.verify.app.lead_both"), brand)
		case hasCode:
			lead = fmt.Sprintf(i18n.T(lang, "mail.verify.app.lead_code"), brand)
		default:
			lead = fmt.Sprintf(i18n.T(lang, "mail.verify.app.lead_link"), brand)
		}
	} else {
		footer = fmt.Sprintf(i18n.T(lang, "mail.verify.platform.footer"), brand)
		switch {
		case hasLink && hasCode:
			lead = fmt.Sprintf(i18n.T(lang, "mail.verify.platform.lead_both"), brand)
		case hasCode:
			lead = fmt.Sprintf(i18n.T(lang, "mail.verify.platform.lead_code"), brand)
		default:
			lead = fmt.Sprintf(i18n.T(lang, "mail.verify.platform.lead_link"), brand)
		}
	}

	var plainParts []string
	plainParts = append(plainParts, fmt.Sprintf("%s\n\n%s", i18n.T(lang, "mail.verify.hi"), lead))
	if hasCode {
		plainParts = append(plainParts, fmt.Sprintf("\n"+i18n.T(lang, "mail.verify.code_label"), d.Code))
	}
	if hasLink {
		plainParts = append(plainParts, fmt.Sprintf("\n"+i18n.T(lang, "mail.verify.or_link"), d.LinkURL))
	}
	plainParts = append(plainParts, "\n"+i18n.T(lang, "mail.verify.expires"))
	plain = strings.Join(plainParts, "") + fmt.Sprintf("\n\n— %s\n%s", brand, strings.TrimSuffix(t.AppURL, "/"))

	htmlBody = layoutVerificationEmail(verificationLayout{
		Brand:   brand,
		Title:   i18n.T(lang, "mail.verify.title"),
		Lead:    lead,
		LinkURL: d.LinkURL,
		Code:    d.Code,
		Note:    i18n.T(lang, "mail.verify.expires"),
		Footer:  footer,
		AppURL:  strings.TrimSuffix(t.AppURL, "/"),
		Lang:    lang,
		ButtonText: i18n.T(lang, "mail.verify.button"),
		CopyLink:   i18n.T(lang, "mail.copy_link"),
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
	Lang       string
	CopyLink   string
}

func layoutVerificationEmail(l verificationLayout) string {
	accent := "#FF3B3B"
	hasLink := strings.TrimSpace(l.LinkURL) != ""
	hasCode := strings.TrimSpace(l.Code) != ""
	btnText := strings.TrimSpace(l.ButtonText)
	if btnText == "" {
		btnText = "Confirm email"
	}
	lang := i18n.Normalize(l.Lang)
	copyLink := strings.TrimSpace(l.CopyLink)
	if copyLink == "" {
		copyLink = i18n.T(lang, "mail.copy_link")
	}

	var body strings.Builder
	body.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
</head>
<body style="margin:0;padding:0;background:#FFF4E6;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#1C1412;">
  <div style="display:none;max-height:0;overflow:hidden;opacity:0;">%s</div>
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#FFF4E6;padding:32px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:520px;background:#FFFFFF;border-radius:16px;border:1px solid #E4DFD9;overflow:hidden;">
          <tr>
            <td style="padding:28px 32px 8px;font-family:Georgia,'Times New Roman',serif;font-size:22px;font-weight:700;letter-spacing:-0.02em;color:#1C1412;">%s</td>
          </tr>
          <tr>
            <td style="padding:8px 32px 0;font-size:20px;font-weight:600;line-height:1.35;color:#1C1412;">%s</td>
          </tr>
          <tr>
            <td style="padding:16px 32px 0;font-size:15px;line-height:1.6;color:#6B6360;">%s</td>
          </tr>`,
		html.EscapeString(lang),
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
              <div style="display:inline-block;padding:16px 24px;border-radius:12px;background:#FFF4E6;font-size:32px;font-weight:700;letter-spacing:0.35em;color:#1C1412;border:1px solid #F7E4CD;">%s</div>
            </td>
          </tr>`, html.EscapeString(l.Code)))
	}

	if hasLink {
		body.WriteString(fmt.Sprintf(`
          <tr>
            <td style="padding:28px 32px 8px;" align="center">
              <a href="%s" style="display:inline-block;background:%s;color:#FFFFFF;text-decoration:none;font-size:15px;font-weight:600;padding:14px 28px;border-radius:10px;">%s</a>
            </td>
          </tr>
          <tr>
            <td style="padding:24px 32px 0;font-size:12px;line-height:1.6;color:#5D6B70;word-break:break-all;">%s<br><a href="%s" style="color:#051B23;">%s</a></td>
          </tr>`,
			html.EscapeString(l.LinkURL),
			accent,
			html.EscapeString(btnText),
			html.EscapeString(copyLink),
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
	lang := t.lang()
	hasCode := strings.TrimSpace(d.Code) != ""

	subject = fmt.Sprintf(i18n.T(lang, "mail.reset.subject"), brand)

	var lead, footer string
	if appScoped {
		footer = fmt.Sprintf(i18n.T(lang, "mail.reset.app.footer"), brand)
	} else {
		footer = i18n.T(lang, "mail.reset.footer")
	}

	if hasCode {
		lead = fmt.Sprintf(i18n.T(lang, "mail.reset.lead_code_app"), brand)
		plain = fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n— %s\n%s",
			i18n.T(lang, "mail.reset.hi"),
			lead,
			fmt.Sprintf(i18n.T(lang, "mail.reset.code_label"), d.Code),
			i18n.T(lang, "mail.reset.code_expires"),
			brand,
			strings.TrimSuffix(t.AppURL, "/"),
		)

		htmlBody = layoutVerificationEmail(verificationLayout{
			Brand:   brand,
			Title:   i18n.T(lang, "mail.reset.title"),
			Lead:    lead,
			Code:    d.Code,
			Note:    i18n.T(lang, "mail.reset.code_expires"),
			Footer:  footer,
			AppURL:  strings.TrimSuffix(t.AppURL, "/"),
			Lang:    lang,
			CopyLink: i18n.T(lang, "mail.copy_link"),
		})
	} else {
		lead = fmt.Sprintf(i18n.T(lang, "mail.reset.lead_link"), brand)
		plain = fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n— %s\n%s",
			i18n.T(lang, "mail.reset.hi"),
			lead,
			d.LinkURL,
			i18n.T(lang, "mail.reset.expires"),
			brand,
			strings.TrimSuffix(t.AppURL, "/"),
		)

		htmlBody = layoutVerificationEmail(verificationLayout{
			Brand:      brand,
			Title:      i18n.T(lang, "mail.reset.title"),
			Lead:       lead,
			ButtonText: i18n.T(lang, "mail.reset.button"),
			LinkURL:    d.LinkURL,
			Note:       i18n.T(lang, "mail.reset.expires"),
			Footer:     footer,
			AppURL:     strings.TrimSuffix(t.AppURL, "/"),
			Lang:       lang,
			CopyLink:   i18n.T(lang, "mail.copy_link"),
		})
	}

	return subject, plain, htmlBody
}
