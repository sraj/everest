package http

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/apperror"
	"github.com/sraj/everest/internal/auth"
	"github.com/sraj/everest/internal/service"
)

type HealthCheck func(ctx context.Context) error

type Handler struct {
	docService     service.DocumentService
	profileService *service.ProfileService
	log            *slog.Logger
	healthChecks   map[string]HealthCheck
	authMiddleware fiber.Handler
	bffHandler     *auth.BFFHandler
}

func New(docService service.DocumentService, log *slog.Logger, authMiddleware ...fiber.Handler) *Handler {
	h := &Handler{
		docService:   docService,
		log:          log,
		healthChecks: make(map[string]HealthCheck),
	}
	if len(authMiddleware) > 0 {
		h.authMiddleware = authMiddleware[0]
	}
	return h
}

func (h *Handler) SetProfileService(ps *service.ProfileService) {
	h.profileService = ps
}

func (h *Handler) SetBFFHandler(bff *auth.BFFHandler) {
	h.bffHandler = bff
}

func (h *Handler) AddHealthCheck(name string, check HealthCheck) {
	h.healthChecks[name] = check
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/health", h.healthCheck)
	app.Get("/api/docs/openapi.json", h.serveOpenAPI)

	if h.bffHandler != nil {
		h.bffHandler.RegisterRoutes(app)
	}

	if h.authMiddleware != nil && h.bffHandler == nil {
		auth := app.Group("/auth")
		auth.Post("/verify", h.authMiddleware, h.handleVerify)
		auth.Get("/me", h.authMiddleware, h.handleMe)
	}

	api := app.Group("/api/v1")
	docs := api.Group("/documents")
	if h.authMiddleware != nil {
		docs.Use(h.authMiddleware)
	}
	docs.Get("/", h.listDocuments)
	docs.Post("/", h.createDocument)
	docs.Get("/:id", h.getDocument)
	docs.Get("/:id/download", h.downloadDocument)
	docs.Put("/:id", h.updateDocument)
	docs.Delete("/:id", h.deleteDocument)

	api.Get("/profile", h.getProfile)
	api.Put("/profile", h.updateProfile)
}

func ErrorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		var body any

		switch e := err.(type) {
		case *fiber.Error:
			code = e.Code
			body = fiber.Map{"kind": "internal", "message": e.Message}
		case *apperror.AppError:
			code = e.Status
			body = e
		default:
			body = fiber.Map{
				"kind":    "internal",
				"message": err.Error(),
			}
		}

		log.Error("request error",
			"error", err.Error(),
			"status", code,
			"path", c.Path(),
		)

		return c.Status(code).JSON(body)
	}
}
