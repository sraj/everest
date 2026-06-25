package http

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/version"
)

func (h *Handler) healthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	resp := HealthResponse{
		Status:  "ok",
		Version: version.Version,
		Commit:  version.Commit,
		Checks:  make(map[string]string),
	}
	code := fiber.StatusOK

	for name, check := range h.healthChecks {
		if err := check(ctx); err != nil {
			resp.Checks[name] = err.Error()
			resp.Status = "degraded"
			code = fiber.StatusServiceUnavailable
		} else {
			resp.Checks[name] = "ok"
		}
	}

	return c.Status(code).JSON(resp)
}
