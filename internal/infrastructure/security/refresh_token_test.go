package security_test

import (
	"testing"

	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
)

func TestGenerateRefreshToken_uniqueAndHashable(t *testing.T) {
	plain1, hash1, err := tokens.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	plain2, hash2, _ := tokens.GenerateRefreshToken()
	if plain1 == plain2 {
		t.Fatal("tokens should be unique")
	}
	if tokens.HashRefreshToken(plain1) != hash1 {
		t.Fatal("hash mismatch")
	}
	_ = hash2
}
