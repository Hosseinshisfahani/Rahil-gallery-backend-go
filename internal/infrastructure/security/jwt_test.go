package security_test

import (
	"testing"
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/security"
	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
	"github.com/google/uuid"
)

func TestJWTProvider_IssueAndParse(t *testing.T) {
	cfg := config.Config{
		JWTAccessSecret: "test-secret-key-at-least-32-chars-long",
		JWTAccessTTL:    time.Minute,
	}
	p := security.NewJWTProvider(cfg)
	userID := uuid.New()

	token, exp, err := p.IssueAccess(tokens.AccessClaims{
		UserID: userID,
		Email:  "jwt@example.com",
		Role:   "customer",
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if time.Until(exp) <= 0 {
		t.Fatal("expiry should be in the future")
	}

	claims, err := p.ParseAccess(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != userID || claims.Email != "jwt@example.com" || claims.Role != "customer" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestJWTProvider_ParseInvalid(t *testing.T) {
	cfg := config.Config{JWTAccessSecret: "secret", JWTAccessTTL: time.Minute}
	p := security.NewJWTProvider(cfg)
	_, err := p.ParseAccess("not.a.jwt")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
