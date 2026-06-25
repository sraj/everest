package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/apperror"
)

func (h *Handler) serveOpenAPI(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	if err := c.SendFile("api/gen/openapiv2/api.swagger.json"); err != nil {
		return apperror.NotFound("openapi spec not found — run: make proto")
	}
	return nil
}
