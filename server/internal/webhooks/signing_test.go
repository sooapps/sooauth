package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
)

func TestSignature(t *testing.T) {
	secret := "whsec_test"
	body := []byte(`{"type":"user.login"}`)
	timestamp := int64(1700000000)
	got := Signature(secret, timestamp, body)
	parts := strings.Split(got, ",")
	if len(parts) != 2 || parts[0] != "t=1700000000" {
		t.Fatalf("unexpected signature header: %q", got)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strconv.FormatInt(timestamp, 10) + "."))
	_, _ = mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))
	if parts[1] != "v1="+want {
		t.Fatalf("signature mismatch: got %q want %q", parts[1], "v1="+want)
	}
}
