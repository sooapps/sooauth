package mail

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestResendSender_Success(t *testing.T) {
	var receivedBody map[string]any
	var authHeader string

	mockClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			authHeader = r.Header.Get("Authorization")
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &receivedBody)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"id":"resend-123"}`)),
			}, nil
		}),
	}

	sender, err := NewSender(ProviderConfig{
		Provider:   "resend",
		APIKey:     "re_test_key",
		FromName:   "Auth Team",
		FromEmail:  "auth@example.com",
		ReplyTo:    "support@example.com",
		BaseURL:    "https://api.resend.test",
		HTTPClient: mockClient,
	})
	if err != nil {
		t.Fatalf("unexpected error creating sender: %v", err)
	}

	err = sender.SendOutbound(Outbound{
		To:      "recipient@example.com",
		Subject: "Test Subject",
		Plain:   "Hello text",
		HTML:    "<p>Hello html</p>",
	})
	if err != nil {
		t.Fatalf("SendOutbound failed: %v", err)
	}

	if authHeader != "Bearer re_test_key" {
		t.Errorf("expected Bearer re_test_key, got %s", authHeader)
	}
	if receivedBody["from"] != "Auth Team <auth@example.com>" {
		t.Errorf("expected from 'Auth Team <auth@example.com>', got %v", receivedBody["from"])
	}
	if toList, ok := receivedBody["to"].([]any); !ok || len(toList) != 1 || toList[0] != "recipient@example.com" {
		t.Errorf("unexpected to: %v", receivedBody["to"])
	}
	if receivedBody["reply_to"] != "support@example.com" {
		t.Errorf("expected reply_to support@example.com, got %v", receivedBody["reply_to"])
	}
}

func TestPostmarkSender_Success(t *testing.T) {
	var receivedBody map[string]any
	var tokenHeader string

	mockClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			tokenHeader = r.Header.Get("X-Postmark-Server-Token")
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &receivedBody)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"ErrorCode":0,"Message":"OK"}`)),
			}, nil
		}),
	}

	sender, err := NewSender(ProviderConfig{
		Provider:   "postmark",
		APIKey:     "pm_test_token",
		FromName:   "Security",
		FromEmail:  "security@example.com",
		BaseURL:    "https://api.postmarkapp.test",
		HTTPClient: mockClient,
	})
	if err != nil {
		t.Fatalf("unexpected error creating sender: %v", err)
	}

	err = sender.SendOutbound(Outbound{
		To:      "target@example.com",
		Subject: "Security alert",
		Plain:   "Alert content",
	})
	if err != nil {
		t.Fatalf("SendOutbound failed: %v", err)
	}

	if tokenHeader != "pm_test_token" {
		t.Errorf("expected pm_test_token, got %s", tokenHeader)
	}
	if receivedBody["From"] != "Security <security@example.com>" {
		t.Errorf("unexpected From: %v", receivedBody["From"])
	}
	if receivedBody["To"] != "target@example.com" {
		t.Errorf("unexpected To: %v", receivedBody["To"])
	}
}

func TestSESSender_Success(t *testing.T) {
	var receivedBody map[string]any
	var authHeader string

	mockClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			authHeader = r.Header.Get("Authorization")
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &receivedBody)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"MessageId":"ses-msg-123"}`)),
			}, nil
		}),
	}

	sender, err := NewSender(ProviderConfig{
		Provider:     "ses",
		AWSAccessKey: "AKIAIOSFODNN7EXAMPLE",
		AWSSecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		AWSRegion:    "us-west-2",
		FromName:     "AWS Team",
		FromEmail:    "noreply@example.com",
		BaseURL:      "https://email.us-west-2.amazonaws.test",
		HTTPClient:   mockClient,
	})
	if err != nil {
		t.Fatalf("unexpected error creating sender: %v", err)
	}

	err = sender.SendOutbound(Outbound{
		To:      "ses-recipient@example.com",
		Subject: "AWS SES Test",
		Plain:   "SES plain text",
		HTML:    "<b>SES html</b>",
	})
	if err != nil {
		t.Fatalf("SendOutbound failed: %v", err)
	}

	if !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/") {
		t.Errorf("unexpected auth header: %s", authHeader)
	}
	if !strings.Contains(authHeader, "/us-west-2/ses/aws4_request") {
		t.Errorf("auth header missing credential scope: %s", authHeader)
	}
	if receivedBody["FromEmailAddress"] != "AWS Team <noreply@example.com>" {
		t.Errorf("unexpected FromEmailAddress: %v", receivedBody["FromEmailAddress"])
	}
}

func TestResendSender_Error(t *testing.T) {
	mockClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"message":"invalid api key"}`)),
			}, nil
		}),
	}

	sender, err := NewSender(ProviderConfig{
		Provider:   "resend",
		APIKey:     "bad_key",
		FromEmail:  "auth@example.com",
		BaseURL:    "https://api.resend.test",
		HTTPClient: mockClient,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = sender.SendOutbound(Outbound{To: "a@b.com", Subject: "Hi", Plain: "Body"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid api key") {
		t.Errorf("expected error message to contain 'invalid api key', got %v", err)
	}
}
