package tokens_test

import (
	"testing"

	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
)

func TestHashRefreshToken_deterministic(t *testing.T) {
	h1 := tokens.HashRefreshToken("abc")
	h2 := tokens.HashRefreshToken("abc")
	if h1 != h2 {
		t.Fatal("hash should be deterministic")
	}
}
