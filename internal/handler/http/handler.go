package http

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/apperror"
	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/service"
	"github.com/sraj/everest/internal/version"
)

type HealthCheck func(ctx context.Context) error

type Handler struct {
	docService   service.DocumentService
	log          *slog.Logger
	healthChecks map[string]HealthCheck
}

func New(docService service.DocumentService, log *slog.Logger) *Handler {
	return &Handler{
		docService:   docService,
		log:          log,
		healthChecks: make(map[string]HealthCheck),
	}
}

func (h *Handler) AddHealthCheck(name string, check HealthCheck) {
	h.healthChecks[name] = check
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/health", h.healthCheck)
	app.Get("/api/docs/openapi.json", h.serveOpenAPI)

	api := app.Group("/api/v1")
	docs := api.Group("/documents")
	docs.Get("/", h.listDocuments)
	docs.Post("/", h.createDocument)
	docs.Get("/:id", h.getDocument)
	docs.Get("/:id/download", h.downloadDocument)
	docs.Get("/:id/thumbnail", h.getDocumentThumbnail)
	docs.Put("/:id", h.updateDocument)
	docs.Delete("/:id", h.deleteDocument)
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

func (h *Handler) listDocuments(c *fiber.Ctx) error {
	var q ListDocumentsQuery
	if err := bindQuery(c, &q); err != nil {
		return err
	}

	p := model.Page{Number: q.Page, Size: q.Size}
	if p.Number == 0 {
		p = model.DefaultPage()
	}

	result, err := h.docService.List(c.Context(), p)
	if err != nil {
		h.log.Error("failed to list documents", "error", err.Error())
		return apperror.Internal("failed to list documents")
	}

	items := make([]DocumentResponse, 0, len(result.Items))
	for _, doc := range result.Items {
		items = append(items, toDocumentResponse(doc))
	}

	return c.JSON(ListDocumentsResponse{
		Documents:  items,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	})
}

func (h *Handler) createDocument(c *fiber.Ctx) error {
	title, content, contentType := "Untitled Document", []byte{}, "text/html"

	if isMultipart(c) {
		body, err := readMultipartBody(c)
		if err != nil {
			return err
		}
		title, content, contentType = body.Title, body.Content, body.ContentType
	} else {
		var req CreateDocumentRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		if req.Title != "" {
			title = req.Title
		}
		if req.Content != "" {
			content = []byte(req.Content)
		}
	}

	doc, err := h.docService.Create(c.Context(), service.CreateDocumentInput{
		Title:       title,
		OwnerID:     "00000000-0000-0000-0000-000000000001",
		Content:     content,
		ContentType: contentType,
	})
	if err != nil {
		h.log.Error("failed to create document", "error", err.Error())
		return apperror.Internal("failed to create document")
	}

	resp := toDocumentResponse(doc)
	resp.ContentType = contentType
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) getDocument(c *fiber.Ctx) error {
	id := c.Params("id")

	doc, err := h.docService.GetByID(c.Context(), id)
	if err != nil {
		return apperror.NotFound("document %s not found", id)
	}

	content, err := h.docService.GetContent(c.Context(), id)
	if err != nil {
		h.log.Error("failed to get document content", "error", err.Error())
		content = []byte{}
	}

	resp := toDocumentResponse(doc)
	resp.Content = string(content)
	return c.JSON(resp)
}

func (h *Handler) downloadDocument(c *fiber.Ctx) error {
	id := c.Params("id")

	doc, err := h.docService.GetByID(c.Context(), id)
	if err != nil {
		return apperror.NotFound("document %s not found", id)
	}

	content, err := h.docService.GetContent(c.Context(), id)
	if err != nil {
		h.log.Error("failed to get document content", "error", err.Error())
		return apperror.Internal("failed to get document content")
	}

	c.Set("Content-Disposition", "attachment; filename=\""+doc.Title+".html\"")
	c.Set("Content-Type", "text/html")
	return c.Send(content)
}

func (h *Handler) getDocumentThumbnail(c *fiber.Ctx) error {
	id := c.Params("id")

	thumbnail, err := h.docService.GetThumbnail(c.Context(), id)
	if err != nil {
		return apperror.NotFound("document %s not found", id)
	}

	if thumbnail == nil {
		return c.SendStatus(fiber.StatusNoContent)
	}

	c.Set("Cache-Control", "public, max-age=60, must-revalidate")
	c.Set("Content-Type", "image/png")
	return c.Send(thumbnail)
}

func (h *Handler) updateDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	title, content := "", []byte{}

	if isMultipart(c) {
		body, err := readMultipartBody(c)
		if err != nil {
			return err
		}
		title = body.Title
		content = body.Content
	} else {
		var req UpdateDocumentRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		if req.Title != nil {
			title = *req.Title
		}
		if req.Content != nil {
			content = []byte(*req.Content)
		}
	}

	doc, err := h.docService.Update(c.Context(), service.UpdateDocumentInput{
		ID:      id,
		Title:   title,
		Content: content,
	})
	if err != nil {
		h.log.Error("failed to update document", "error", err.Error())
		return apperror.Internal("failed to update document")
	}

	return c.JSON(toDocumentResponse(doc))
}

func (h *Handler) deleteDocument(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.docService.Delete(c.Context(), id); err != nil {
		h.log.Error("failed to delete document", "error", err.Error())
		return apperror.Internal("failed to delete document")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) serveOpenAPI(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	if err := c.SendFile("api/gen/openapiv2/api.swagger.json"); err != nil {
		return apperror.NotFound("openapi spec not found — run: make proto")
	}
	return nil
}
