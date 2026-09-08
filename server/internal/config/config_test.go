package config

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

func TestLoadRequiresMFAEncryptionKeyInProduction(t *testing.T) {
	t.Setenv("SOOAUTH_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("REDIS_URL", "redis://example")
	t.Setenv("MFA_ENCRYPTION_KEY", "")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "MFA_ENCRYPTION_KEY") {
		t.Fatalf("expected production MFA key error, got %v", err)
	}
}

func TestLoadDecodesMFAEncryptionKey(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	t.Setenv("SOOAUTH_ENV", "development")
	t.Setenv("MFA_ENCRYPTION_KEY", base64.RawStdEncoding.EncodeToString(key))

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if string(cfg.MFAEncryptionKey) != string(key) {
		t.Fatalf("decoded unexpected MFA key")
	}
}

func TestLoadDecodesHexMFAEncryptionKey(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	t.Setenv("SOOAUTH_ENV", "development")
	t.Setenv("MFA_ENCRYPTION_KEY", hex.EncodeToString(key))

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if string(cfg.MFAEncryptionKey) != string(key) {
		t.Fatalf("decoded unexpected hex MFA key")
	}
}
