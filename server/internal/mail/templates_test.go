package mail

import (
	"strings"
	"testing"
)

func TestVerificationEmailContent(t *testing.T) {
	tx := Transactional{BrandName: "sooauth", AppURL: "https://auth.sooauth.com"}
	subject, plain, html := tx.VerificationEmail("https://auth.sooauth.com/auth/verify?token=abc")

	if !strings.Contains(subject, "sooauth") {
		t.Fatalf("subject = %q", subject)
	}
	if !strings.Contains(plain, "Thanks for signing up") {
		t.Fatalf("plain missing intro: %q", plain)
	}
	if !strings.Contains(html, "Confirm email") || !strings.Contains(html, "token=abc") {
		t.Fatalf("html missing button or link")
	}
}

func TestVerificationDeliveryEmail(t *testing.T) {
	tx := Transactional{BrandName: "Soobrief", AppURL: "https://auth.sooauth.com"}

	subject, plain, html := tx.VerificationDeliveryEmail(VerificationDelivery{
		LinkURL: "https://auth.sooauth.com/auth/verify?token=abc",
		Code:    "123456",
	}, true)

	if !strings.Contains(subject, "Soobrief") {
		t.Fatalf("subject = %q", subject)
	}
	if !strings.Contains(plain, "123456") || !strings.Contains(plain, "token=abc") {
		t.Fatalf("plain missing code or link: %q", plain)
	}
	if !strings.Contains(html, "123456") || !strings.Contains(html, "Confirm email") {
		t.Fatalf("html missing code or button")
	}
}

func TestBuildMessage_multipart(t *testing.T) {
	raw := buildMessage("sooauth <noreply@sooauth.com>", Outbound{
		To:      "user@example.com",
		Subject: "Test",
		Plain:   "plain body",
		HTML:    "<p>html body</p>",
	})
	if !strings.Contains(raw, "multipart/alternative") {
		t.Fatal("expected multipart message")
	}
	if !strings.Contains(raw, "plain body") || !strings.Contains(raw, "<p>html body</p>") {
		t.Fatal("missing parts")
	}
}

func TestPasswordResetDeliveryEmail(t *testing.T) {
	tx := Transactional{BrandName: "AcmeApp", AppURL: "https://auth.acme.local"}

	// Test code only delivery
	subCode, plainCode, htmlCode := tx.PasswordResetDeliveryEmail(PasswordResetDelivery{
		Code: "654321",
	}, true)
	if !strings.Contains(subCode, "AcmeApp") {
		t.Fatalf("expected brand in subject: %q", subCode)
	}
	if !strings.Contains(plainCode, "654321") || strings.Contains(plainCode, "Open this link") {
		t.Fatalf("expected code only in plain text: %q", plainCode)
	}
	if !strings.Contains(htmlCode, "654321") || strings.Contains(htmlCode, "Confirm email") {
		t.Fatalf("html incorrect: %q", htmlCode)
	}

	// Test link only delivery
	subLink, plainLink, htmlLink := tx.PasswordResetDeliveryEmail(PasswordResetDelivery{
		LinkURL: "https://auth.acme.local/auth/reset-password?token=secret123",
	}, false)
	if !strings.Contains(subLink, "AcmeApp") {
		t.Fatalf("expected brand in subject: %q", subLink)
	}
	if !strings.Contains(plainLink, "secret123") || strings.Contains(plainLink, "Your password reset code") {
		t.Fatalf("expected link only in plain text: %q", plainLink)
	}
	if !strings.Contains(htmlLink, "Reset password") || !strings.Contains(htmlLink, "secret123") {
		t.Fatalf("html missing reset button or link: %q", htmlLink)
	}
}
