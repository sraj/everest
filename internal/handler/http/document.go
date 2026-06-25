package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/apperror"
	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/service"
)

func (h *Handler) listDocuments(c *fiber.Ctx) error {
	var q ListDocumentsQuery
	if err := bindQuery(c, &q); err != nil {
		return err
	}

	p := model.Page{Number: q.Page, Size: q.Size}
	if p.Number == 0 {
		p = model.DefaultPage()
	}

	ownerID := h.ownerFromContext(c)

	result, err := h.docService.List(c.Context(), p, ownerID)
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

	ownerID := h.ownerFromContext(c)

	doc, err := h.docService.Create(c.Context(), service.CreateDocumentInput{
		Title:       title,
		OwnerID:     ownerID,
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
	ownerID := h.ownerFromContext(c)

	doc, err := h.docService.GetByID(c.Context(), id, ownerID)
	if err != nil {
		return apperror.NotFound("document %s not found", id)
	}

	content, err := h.docService.GetContent(c.Context(), id, ownerID)
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
	ownerID := h.ownerFromContext(c)

	doc, err := h.docService.GetByID(c.Context(), id, ownerID)
	if err != nil {
		return apperror.NotFound("document %s not found", id)
	}

	content, err := h.docService.GetContent(c.Context(), id, ownerID)
	if err != nil {
		h.log.Error("failed to get document content", "error", err.Error())
		return apperror.Internal("failed to get document content")
	}

	c.Set("Content-Disposition", "attachment; filename=\""+sanitizeFilename(doc.Title)+".html\"")
	c.Set("Content-Type", "text/html")
	return c.Send(content)
}

func (h *Handler) getDocumentThumbnail(c *fiber.Ctx) error {
	id := c.Params("id")
	ownerID := h.ownerFromContext(c)

	thumbnail, err := h.docService.GetThumbnail(c.Context(), id, ownerID)
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
	}, h.ownerFromContext(c))
	if err != nil {
		h.log.Error("failed to update document", "error", err.Error())
		return apperror.Internal("failed to update document")
	}

	return c.JSON(toDocumentResponse(doc))
}

func (h *Handler) deleteDocument(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.docService.Delete(c.Context(), id, h.ownerFromContext(c)); err != nil {
		h.log.Error("failed to delete document", "error", err.Error())
		return apperror.Internal("failed to delete document")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
