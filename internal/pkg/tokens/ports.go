package tokens

import (
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type AccessClaims struct {
	UserID shared.ID
	Email  string
	Role   string
}

type TokenProvider interface {
	IssueAccess(claims AccessClaims) (token string, expiresAt time.Time, err error)
	ParseAccess(token string) (AccessClaims, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}
