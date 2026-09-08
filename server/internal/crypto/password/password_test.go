package password_test

import (
	"testing"

	"github.com/sooapps/sooauth/server/internal/crypto/password"
)

func TestHashAndVerify(t *testing.T) {
	encoded, err := password.Hash("correct-horse-battery")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := password.Verify("correct-horse-battery", encoded)
	if err != nil || !ok {
		t.Fatalf("expected match, ok=%v err=%v", ok, err)
	}
	ok, err = password.Verify("wrong-password", encoded)
	if err != nil || ok {
		t.Fatalf("expected mismatch, ok=%v err=%v", ok, err)
	}
}

func TestWeakPasswordRejected(t *testing.T) {
	_, err := password.Hash("short")
	if err != password.ErrWeakPassword {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}
