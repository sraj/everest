package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/apperror"
	"github.com/sraj/everest/internal/domain/model"
)

func (h *Handler) getProfile(c *fiber.Ctx) error {
	userID := h.ownerFromContext(c)
	profile, err := h.profileService.Get(c.Context(), userID)
	if err != nil {
		h.log.Error("failed to get profile", "error", err.Error())
		return apperror.Internal("failed to get profile")
	}
	return c.JSON(profile)
}

func (h *Handler) updateProfile(c *fiber.Ctx) error {
	userID := h.ownerFromContext(c)

	var req struct {
		Nickname    string       `json:"nickname"`
		Preferences model.Prefs  `json:"preferences"`
	}
	if err := c.BodyParser(&req); err != nil {
		return apperror.BadRequest("invalid request body")
	}

	profile, err := h.profileService.Update(c.Context(), userID, req.Nickname, req.Preferences)
	if err != nil {
		h.log.Error("failed to update profile", "error", err.Error())
		return apperror.Internal("failed to update profile")
	}
	return c.JSON(profile)
}
