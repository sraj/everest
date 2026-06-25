package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/auth"
)

func (h *Handler) handleVerify(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func (h *Handler) handleMe(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(auth.IntrospectUser)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "not authenticated",
		})
	}
	return c.JSON(fiber.Map{
		"sub":   user.Sub,
		"name":  user.Name,
		"email": user.Email,
	})
}

func (h *Handler) ownerFromContext(c *fiber.Ctx) string {
	if user, ok := c.Locals("user").(auth.IntrospectUser); ok {
		return user.Sub
	}
	return "00000000-0000-0000-0000-000000000001"
}

func sanitizeFilename(name string) string {
	if name == "" {
		return "Untitled Document"
	}
	result := make([]byte, 0, len(name))
	for _, b := range []byte(name) {
		if b == '\r' || b == '\n' || b == '"' {
			continue
		}
		result = append(result, b)
	}
	return string(result)
}
