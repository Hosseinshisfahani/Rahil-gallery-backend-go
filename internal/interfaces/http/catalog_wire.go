package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	appcatalog "github.com/rahil-gallery/rahil-gallery-server/internal/application/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/catalog"
	catalogpg "github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/persistence/postgres/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http/handler"
)

func postgresCatalogWire(pool *pgxpool.Pool) catalog.Repository {
	return catalogpg.NewRepository(pool)
}

func RegisterCatalogRoutes(router fiber.Router, pool *pgxpool.Pool) {
	svc := appcatalog.NewService(postgresCatalogWire(pool))
	h := handler.NewCatalogHandler(svc)

	router.Get("/categories", h.ListCategories)
	router.Get("/categories/:slug", h.GetCategory)

	router.Get("/collections", h.ListCollections)
	router.Get("/collections/:slug", h.GetCollection)
	router.Get("/collections/:slug/products", h.ListCollectionProducts)

	router.Get("/products", h.ListProducts)
	router.Get("/products/:slug", h.GetProduct)
}
