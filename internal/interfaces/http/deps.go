package http

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/identity"
)

type RouterDeps struct {
	Pool   *pgxpool.Pool
	Config config.Config
}

func postgresAuthWire(pool *pgxpool.Pool) AuthWire {
	return AuthWire{
		Users:         identity.NewUserRepository(pool),
		Roles:         identity.NewRoleRepository(pool),
		RefreshTokens: identity.NewRefreshTokenRepository(pool),
	}
}
