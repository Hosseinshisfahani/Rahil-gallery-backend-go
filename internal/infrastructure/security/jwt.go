package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
	"github.com/rahil-gallery/rahil-gallery-server/internal/pkg/tokens"
)

type accessClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Role  string `json:"role"`
}

type JWTProvider struct {
	secret    []byte
	accessTTL time.Duration
}

func NewJWTProvider(cfg config.Config) JWTProvider {
	return JWTProvider{
		secret:    []byte(cfg.JWTAccessSecret),
		accessTTL: cfg.JWTAccessTTL,
	}
}

func (p JWTProvider) IssueAccess(claims tokens.AccessClaims) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(p.accessTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
		Email: claims.Email,
		Role:  claims.Role,
	})

	signed, err := token.SignedString(p.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}

	return signed, expiresAt, nil
}

func (p JWTProvider) ParseAccess(tokenString string) (tokens.AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return p.secret, nil
	})
	if err != nil {
		return tokens.AccessClaims{}, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return tokens.AccessClaims{}, fmt.Errorf("invalid token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return tokens.AccessClaims{}, fmt.Errorf("invalid subject: %w", err)
	}

	return tokens.AccessClaims{
		UserID: shared.ID(userID),
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}
