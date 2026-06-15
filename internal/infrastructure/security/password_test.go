package security_test

import (
	"testing"

	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/security"
)

func TestBcryptHasher_HashAndCompare(t *testing.T) {
	h := security.NewBcryptHasher()
	hash, err := h.Hash("Secret12")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "Secret12" {
		t.Fatal("hash should not equal plaintext")
	}
	if err := h.Compare(hash, "Secret12"); err != nil {
		t.Fatalf("compare valid: %v", err)
	}
	if err := h.Compare(hash, "WrongPass1"); err == nil {
		t.Fatal("expected mismatch error")
	}
}
