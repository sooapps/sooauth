package encryption

import (
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	cipher, err := New([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatal(err)
	}
	protected, err := cipher.Encrypt("JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatal(err)
	}
	if !IsEncrypted(protected) || strings.Contains(protected, "JBSWY3DPEHPK3PXP") {
		t.Fatal("secret was not stored as an encrypted envelope")
	}
	plain, err := cipher.Decrypt(protected)
	if err != nil || plain != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("decrypt = %q, %v", plain, err)
	}
}

func TestDecryptRejectsTampering(t *testing.T) {
	cipher, err := New([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatal(err)
	}
	protected, err := cipher.Encrypt("secret")
	if err != nil {
		t.Fatal(err)
	}
	chars := []byte(protected)
	if chars[10] == 'A' {
		chars[10] = 'B'
	} else {
		chars[10] = 'A'
	}
	protected = string(chars)
	if _, err := cipher.Decrypt(protected); err != ErrInvalidCiphertext {
		t.Fatalf("expected tampering error, got %v", err)
	}
}
