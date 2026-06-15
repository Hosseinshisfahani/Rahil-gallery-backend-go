package identity

import (
	"context"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type UserRepository interface {
	FindByID(ctx context.Context, id shared.ID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	SoftDelete(ctx context.Context, id shared.ID) error
}

type RoleRepository interface {
	FindByID(ctx context.Context, id shared.ID) (*Role, error)
	FindByName(ctx context.Context, name string) (*Role, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id shared.ID) error
}

type UserAddressRepository interface {
	ListByUserID(ctx context.Context, userID shared.ID) ([]UserAddress, error)
	Create(ctx context.Context, addr *UserAddress) error
	Update(ctx context.Context, addr *UserAddress) error
	Delete(ctx context.Context, id shared.ID) error
}
