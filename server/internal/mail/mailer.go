package mail

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"net/url"
	"strings"
)

type Settings struct {
	URL       string
	Host      string
	Port      string
	User      string
	Password  string
	From      string
	BrandName string
	TLSMode   string
}

type Outbound struct {
	To      string
	Subject string
	Plain   string
	HTML    string
}

type Sender interface {
	SendOutbound(msg Outbound) error
}

type Mailer struct {
	settings Settings
	from     string
}

func New(smtpURL, from string) *Mailer {
	return NewWithSettings(Settings{URL: smtpURL, From: from})
}

func NewWithSettings(settings Settings) *Mailer {
	from := settings.From
	if from == "" {
		from = "sooauth@localhost"
	}
	return &Mailer{settings: settings, from: from}
}

func (m *Mailer) Send(to, subject, body string) error {
	return m.SendOutbound(Outbound{To: to, Subject: subject, Plain: body})
}

func (m *Mailer) SendOutbound(msg Outbound) error {
	if !m.configured() {
		slog.Info("email (dev)", "to", msg.To, "subject", msg.Subject, "body", msg.Plain)
		return nil
	}

	host, auth, implicitTLS, err := resolveSMTP(m.settings)
	if err != nil {
		return err
	}

	raw := buildMessage(m.formattedFrom(), msg)
	if strings.EqualFold(m.settings.TLSMode, "implicit") || (implicitTLS && m.settings.TLSMode == "") {
		return sendImplicitTLS(host, auth, m.fromAddress(), msg.To, []byte(raw))
	}
	if strings.EqualFold(m.settings.TLSMode, "none") {
		return sendPlainSMTP(host, auth, m.fromAddress(), msg.To, []byte(raw))
	}
	return smtp.SendMail(host, auth, m.fromAddress(), []string{msg.To}, []byte(raw))
}

func (m *Mailer) formattedFrom() string {
	from := m.from
	if strings.Contains(from, "<") {
		return from
	}
	brand := strings.TrimSpace(m.settings.BrandName)
	if brand == "" {
		brand = "sooauth"
	}
	return fmt.Sprintf("%s <%s>", brand, from)
}

func (m *Mailer) fromAddress() string {
	from := m.from
	if i := strings.LastIndex(from, "<"); i >= 0 {
		from = strings.TrimSuffix(strings.TrimSpace(from[i+1:]), ">")
	}
	return from
}

func buildMessage(from string, msg Outbound) string {
	if msg.HTML == "" {
		return strings.Join([]string{
			fmt.Sprintf("From: %s", from),
			fmt.Sprintf("To: %s", msg.To),
			fmt.Sprintf("Subject: %s", msg.Subject),
			"MIME-Version: 1.0",
			"Content-Type: text/plain; charset=UTF-8",
			"Content-Transfer-Encoding: 8bit",
			"",
			msg.Plain,
		}, "\r\n")
	}

	boundary := "sooauth-" + fmt.Sprintf("%x", len(msg.Subject)+len(msg.Plain))
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", msg.To))
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q\r\n", boundary))
	b.WriteString("\r\n")
	b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	b.WriteString(msg.Plain)
	b.WriteString("\r\n")
	b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	b.WriteString(msg.HTML)
	b.WriteString("\r\n")
	b.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	return b.String()
}

func (m *Mailer) configured() bool {
	if m.settings.Host != "" && m.settings.User != "" {
		return true
	}
	return m.settings.URL != ""
}

func resolveSMTP(settings Settings) (addr string, auth smtp.Auth, implicitTLS bool, err error) {
	if settings.Host != "" {
		return parseSMTPParts(settings.Host, settings.Port, settings.User, settings.Password)
	}
	return parseSMTPURL(settings.URL)
}

func parseSMTPParts(host, port, user, password string) (addr string, auth smtp.Auth, implicitTLS bool, err error) {
	if host == "" {
		return "", nil, false, fmt.Errorf("invalid SMTP config")
	}
	if port == "" {
		port = "465"
	}
	addr = net.JoinHostPort(host, port)
	implicitTLS = port == "465"
	if user != "" {
		auth = smtp.PlainAuth(user, user, password, host)
	}
	return addr, auth, implicitTLS, nil
}

func parseSMTPURL(raw string) (addr string, auth smtp.Auth, implicitTLS bool, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", nil, false, fmt.Errorf("invalid SMTP_URL: encode special characters in password (@ → %%40, ? → %%3F, ! → %%21)")
	}
	if u.Scheme != "smtp" || u.Host == "" {
		return "", nil, false, fmt.Errorf("invalid SMTP_URL")
	}

	hostname := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "587"
	}
	addr = net.JoinHostPort(hostname, port)
	implicitTLS = port == "465"

	if u.User != nil {
		pass, _ := u.User.Password()
		username := u.User.Username()
		auth = smtp.PlainAuth(username, username, pass, hostname)
	}
	return addr, auth, implicitTLS, nil
}

func sendImplicitTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	return w.Close()
}

func sendPlainSMTP(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	return w.Close()
}

