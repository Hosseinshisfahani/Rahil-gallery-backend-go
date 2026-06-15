package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	appcatalog "github.com/rahil-gallery/rahil-gallery-server/internal/application/catalog"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

func mapProductError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, shared.ErrInvalidInput):
		return productBadRequest(c, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, shared.ErrNotFound):
		return productNotFound(c, "NOT_FOUND", err.Error())
	case errors.Is(err, shared.ErrConflict):
		return productConflict(c, "CONFLICT", err.Error())
	case errors.Is(err, appcatalog.ErrSKURequired), errors.Is(err, appcatalog.ErrNameRequired),
		errors.Is(err, appcatalog.ErrCategoryRequired), errors.Is(err, appcatalog.ErrInvalidStatus),
		errors.Is(err, appcatalog.ErrInvalidJewelry), errors.Is(err, appcatalog.ErrInvalidMetal),
		errors.Is(err, appcatalog.ErrInvalidGemstone), errors.Is(err, appcatalog.ErrVariantRequired),
		errors.Is(err, appcatalog.ErrImageURLRequired), errors.Is(err, appcatalog.ErrNegativePrice),
		errors.Is(err, appcatalog.ErrNegativeInventory):
		return productBadRequest(c, "VALIDATION_ERROR", err.Error())
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{"code": "INTERNAL_ERROR", "message": "something went wrong"},
		})
	}
}

func productBadRequest(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"error": fiber.Map{"code": code, "message": msg},
	})
}

func productNotFound(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": fiber.Map{"code": code, "message": msg},
	})
}

func productConflict(c *fiber.Ctx, code, msg string) error {
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{
		"error": fiber.Map{"code": code, "message": msg},
	})
}
