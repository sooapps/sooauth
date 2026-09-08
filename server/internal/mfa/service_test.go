package mfa

import (
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func TestValidCodeUsesTOTP(t *testing.T) {
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "test", AccountName: "user@example.com", Digits: otp.DigitsSix})
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(key.Secret(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !validCode(key.Secret(), code) {
		t.Fatal("expected generated code to validate")
	}
	if validCode(key.Secret(), "not-a-code") {
		t.Fatal("unexpected invalid code validation")
	}
}

func TestBackupCodesAreUniqueAndHashed(t *testing.T) {
	codes, hashes, err := backupCodes()
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != 10 || len(hashes) != 10 {
		t.Fatalf("expected ten backup codes, got %d and %d", len(codes), len(hashes))
	}
	seen := make(map[string]bool, len(codes))
	for i, code := range codes {
		if seen[code] || hashes[i] == code {
			t.Fatal("backup codes must be unique and stored as hashes")
		}
		seen[code] = true
	}
}
