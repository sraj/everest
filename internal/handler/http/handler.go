package http

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"time"

	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/apperror"
	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/service"
	"github.com/sraj/everest/internal/version"
)

// HealthCheck is a function that checks a dependency's health.
type HealthCheck func(ctx context.Context) error

// Handler holds dependencies for HTTP handlers
type Handler struct {
	docService   service.DocumentService
	log          *slog.Logger
	healthChecks map[string]HealthCheck
}

// New creates a new handler with dependencies
func New(docService service.DocumentService, log *slog.Logger) *Handler {
	return &Handler{
		docService:   docService,
		log:          log,
		healthChecks: make(map[string]HealthCheck),
	}
}

// AddHealthCheck registers a named dependency health check.
func (h *Handler) AddHealthCheck(name string, check HealthCheck) {
	h.healthChecks[name] = check
}

// RegisterRoutes registers all HTTP routes
func (h *Handler) RegisterRoutes(app *fiber.App) {
	// Health check
	app.Get("/health", h.healthCheck)

	// API documentation
	app.Get("/api/docs/openapi.json", h.serveOpenAPI)

	// API v1 routes
	api := app.Group("/api/v1")

	// Document routes
	docs := api.Group("/documents")
	docs.Get("/", h.listDocuments)
	docs.Post("/", h.createDocument)
	docs.Get("/:id", h.getDocument)
	docs.Get("/:id/download", h.downloadDocument)
	docs.Get("/:id/thumbnail", h.getDocumentThumbnail)
	docs.Put("/:id", h.updateDocument)
	docs.Delete("/:id", h.deleteDocument)
}

// ErrorHandler returns a custom error handler
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

// Health check handler
func (h *Handler) healthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	status := "ok"
	code := fiber.StatusOK
	checks := make(map[string]string)

	for name, check := range h.healthChecks {
		if err := check(ctx); err != nil {
			checks[name] = err.Error()
			status = "degraded"
			code = fiber.StatusServiceUnavailable
		} else {
			checks[name] = "ok"
		}
	}

	return c.Status(code).JSON(fiber.Map{
		"status":  status,
		"version": version.Version,
		"commit":  version.Commit,
		"checks":  checks,
	})
}

// DocumentResponse represents the API response for a document
type DocumentResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Content      string `json:"content,omitempty"`
	ContentType  string `json:"content_type,omitempty"`
	FileName     string `json:"file_name,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Document handlers
func (h *Handler) listDocuments(c *fiber.Ctx) error {
	ctx := c.Context()

	page := model.DefaultPage()
	if p, err := strconv.Atoi(c.Query("page", "1")); err == nil && p > 0 {
		page.Number = p
	}
	if s, err := strconv.Atoi(c.Query("size", "20")); err == nil && s > 0 && s <= 100 {
		page.Size = s
	}

	result, err := h.docService.List(ctx, page)
	if err != nil {
		h.log.Error("failed to list documents", "error", err)
		return apperror.Internal("failed to list documents")
	}

	responses := make([]DocumentResponse, 0, len(result.Items))
	for _, doc := range result.Items {
		resp := DocumentResponse{
			ID:        doc.ID,
			Title:     doc.Title,
			CreatedAt: doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if doc.ThumbnailID != nil {
			resp.ThumbnailURL = "/api/v1/documents/" + doc.ID + "/thumbnail"
		}
		responses = append(responses, resp)
	}

	return c.JSON(fiber.Map{
		"documents":   responses,
		"total":       result.Total,
		"page":        result.Page,
		"page_size":   result.PageSize,
		"total_pages": result.TotalPages,
	})
}

func (h *Handler) createDocument(c *fiber.Ctx) error {
	contentType := string(c.Request().Header.ContentType())

	var title string
	var content []byte
	var fileContentType string

	// Handle multipart form upload
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Get metadata from form fields
		title = c.FormValue("title")
		if title == "" {
			title = "Untitled Document"
		}

		// Get file from form
		file, err := c.FormFile("file")
		if err != nil {
			// No file uploaded, check for content field
			content = []byte(c.FormValue("content"))
			fileContentType = "text/html"
		} else {
			// Read file content
			f, err := file.Open()
			if err != nil {
				return apperror.BadRequest("failed to open uploaded file")
			}
			defer f.Close()

			content, err = io.ReadAll(f)
			if err != nil {
				return apperror.BadRequest("failed to read uploaded file")
			}

			fileContentType = file.Header.Get("Content-Type")
			if fileContentType == "" {
				fileContentType = "application/octet-stream"
			}

			// Use filename as title if title not provided
			if title == "Untitled Document" && file.Filename != "" {
				title = file.Filename
			}
		}
	} else {
		// Handle JSON request
		var req struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}
		if err := c.BodyParser(&req); err != nil {
			h.log.Error("failed to parse request body", "error", err, "body", string(c.Body()))
			return apperror.BadRequest("invalid request body")
		}
		title = req.Title
		if title == "" {
			title = "Untitled Document"
		}
		content = []byte(req.Content)
		fileContentType = "text/html"
	}

	doc, err := h.docService.Create(c.Context(), service.CreateDocumentInput{
		Title:       title,
		OwnerID:     "00000000-0000-0000-0000-000000000001", // TODO: Get from auth
		Content:     content,
		ContentType: fileContentType,
	})
	if err != nil {
		h.log.Error("failed to create document", "error", err)
		return apperror.Internal("failed to create document")
	}

	return c.Status(fiber.StatusCreated).JSON(DocumentResponse{
		ID:          doc.ID,
		Title:       doc.Title,
		ContentType: fileContentType,
		CreatedAt:   doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *Handler) getDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := c.Context()

	doc, err := h.docService.GetByID(ctx, id)
	if err != nil {
		return apperror.NotFound("document %s not found", id)
	}

	content, err := h.docService.GetContent(ctx, id)
	if err != nil {
		h.log.Error("failed to get document content", "error", err)
		content = []byte{}
	}

	return c.JSON(DocumentResponse{
		ID:        doc.ID,
		Title:     doc.Title,
		Content:   string(content),
		CreatedAt: doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *Handler) downloadDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := c.Context()

	doc, err := h.docService.GetByID(ctx, id)
	if err != nil {
		return apperror.NotFound("document %s not found", id)
	}

	content, err := h.docService.GetContent(ctx, id)
	if err != nil {
		h.log.Error("failed to get document content", "error", err)
		return apperror.Internal("failed to get document content")
	}

	// Set headers for file download
	c.Set("Content-Disposition", "attachment; filename=\""+doc.Title+".html\"")
	c.Set("Content-Type", "text/html")

	return c.Send(content)
}

func (h *Handler) getDocumentThumbnail(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := c.Context()

	thumbnail, err := h.docService.GetThumbnail(ctx, id)
	if err != nil {
		return apperror.NotFound("document %s not found", id)
	}

	if thumbnail == nil {
		// Return a 204 No Content if thumbnail doesn't exist yet
		return c.SendStatus(fiber.StatusNoContent)
	}

	// Set caching headers (thumbnails rarely change)
	c.Set("Cache-Control", "public, max-age=3600")
	c.Set("Content-Type", "image/png")

	return c.Send(thumbnail)
}

func (h *Handler) updateDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	contentType := string(c.Request().Header.ContentType())

	var title string
	var content []byte

	// Handle multipart form upload
	if strings.HasPrefix(contentType, "multipart/form-data") {
		title = c.FormValue("title")

		// Get file from form
		file, err := c.FormFile("file")
		if err != nil {
			// No file uploaded, check for content field
			contentStr := c.FormValue("content")
			if contentStr != "" {
				content = []byte(contentStr)
			}
		} else {
			// Read file content
			f, err := file.Open()
			if err != nil {
				return apperror.BadRequest("failed to open uploaded file")
			}
			defer f.Close()

			content, err = io.ReadAll(f)
			if err != nil {
				return apperror.BadRequest("failed to read uploaded file")
			}
		}
	} else {
		// Handle JSON request
		var req struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}
		if err := c.BodyParser(&req); err != nil {
			return apperror.BadRequest("invalid request body")
		}
		title = req.Title
		content = []byte(req.Content)
	}

	doc, err := h.docService.Update(c.Context(), service.UpdateDocumentInput{
		ID:      id,
		Title:   title,
		Content: content,
	})
	if err != nil {
		h.log.Error("failed to update document", "error", err)
		return apperror.Internal("failed to update document")
	}

	return c.JSON(DocumentResponse{
		ID:        doc.ID,
		Title:     doc.Title,
		CreatedAt: doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *Handler) deleteDocument(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.docService.Delete(c.Context(), id); err != nil {
		h.log.Error("failed to delete document", "error", err)
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
