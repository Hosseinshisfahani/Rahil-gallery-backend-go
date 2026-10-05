package handler

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/repository"
	"github.com/rahil-gallery/rahil-gallery-server/internal/service"
)

type AppDeps struct {
	Pool      *pgxpool.Pool
	Config    config.Config
	Tokens    service.JWTProvider
	Auth      *service.AuthService
	Customers *repository.Repository
	Customer  *service.CustomerService
	Bulk      *service.BulkService
	Jobs      *repository.JobRepository
}
