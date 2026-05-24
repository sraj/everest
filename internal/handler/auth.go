package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/infrastructure/zitadel"
)

// handleVerify validates a Bearer token and returns user info.
//
//	POST /auth/verify
//	Authorization: Bearer <access_token>
func (h *Handler) handleVerify(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(zitadel.IntrospectUser)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "not authenticated",
		})
	}
	return c.JSON(user)
}

// handleMe returns the current user info from the token.
//
//	GET /auth/me
//	Authorization: Bearer <access_token>
func (h *Handler) handleMe(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(zitadel.IntrospectUser)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "not authenticated",
		})
	}
	return c.JSON(user)
}
