package http

import (
	"github.com/sraj/everest/internal/domain/model"
)

// DocumentResponse is the API representation of a document.
type DocumentResponse struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Content      string   `json:"content,omitempty"`
	ContentType  string   `json:"content_type,omitempty"`
	FileName     string   `json:"file_name,omitempty"`
	ThumbnailURL string   `json:"thumbnail_url,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// ListDocumentsResponse wraps a paginated document list.
type ListDocumentsResponse struct {
	Documents  []DocumentResponse `json:"documents"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// HealthResponse is the API representation of the health check.
type HealthResponse struct {
	Status  string            `json:"status"`
	Version string            `json:"version"`
	Commit  string            `json:"commit"`
	Checks  map[string]string `json:"checks"`
}

func toDocumentResponse(doc *model.Document) DocumentResponse {
	resp := DocumentResponse{
		ID:        doc.ID,
		Title:     doc.Title,
		Tags:      doc.Tags,
		CreatedAt: doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if doc.ThumbnailID != nil {
		resp.ThumbnailURL = "/api/v1/documents/" + doc.ID + "/thumbnail"
	}
	return resp
}
