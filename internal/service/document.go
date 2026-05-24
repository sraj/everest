package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/domain/repository"
)

// DocumentService handles document business logic
type DocumentService struct {
	docRepo      repository.DocumentRepository
	contentRepo  repository.ContentRepository
	thumbnailSvc *ThumbnailService
	log          *slog.Logger
}

// NewDocumentService creates a new document service
func NewDocumentService(
	docRepo repository.DocumentRepository,
	contentRepo repository.ContentRepository,
	thumbnailSvc *ThumbnailService,
	log *slog.Logger,
) *DocumentService {
	return &DocumentService{
		docRepo:      docRepo,
		contentRepo:  contentRepo,
		thumbnailSvc: thumbnailSvc,
		log:          log,
	}
}

// CreateDocumentInput represents input for creating a document
type CreateDocumentInput struct {
	Title       string
	OwnerID     string
	Content     []byte
	ContentType string
}

// CreateDocument creates a new document
func (s *DocumentService) CreateDocument(ctx context.Context, input CreateDocumentInput) (*model.Document, error) {
	id := uuid.New().String()
	contentID := uuid.New().String()
	now := time.Now()

	// Save content to MinIO
	contentType := input.ContentType
	if contentType == "" {
		contentType = "text/html"
	}
	if err := s.contentRepo.Save(ctx, contentID, input.Content, contentType); err != nil {
		s.log.Error("failed to save document content", "error", err)
		return nil, err
	}

	// Create document record
	doc := &model.Document{
		ID:        id,
		Title:     input.Title,
		OwnerID:   input.OwnerID,
		ContentID: contentID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Generate thumbnail asynchronously (don't block document creation)
	if s.thumbnailSvc != nil && len(input.Content) > 0 {
		go s.generateAndSaveThumbnail(context.Background(), doc.ID, input.Content)
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		s.log.Error("failed to create document", "error", err)
		// Cleanup content on failure
		_ = s.contentRepo.Delete(ctx, contentID)
		return nil, err
	}

	return doc, nil
}

// GetDocument retrieves a document by ID
func (s *DocumentService) GetDocument(ctx context.Context, id string) (*model.Document, error) {
	return s.docRepo.GetByID(ctx, id)
}

// GetDocumentContent retrieves document content
func (s *DocumentService) GetDocumentContent(ctx context.Context, id string) ([]byte, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.contentRepo.Get(ctx, doc.ContentID)
}

// UpdateDocumentInput represents input for updating a document
type UpdateDocumentInput struct {
	ID      string
	Title   string
	Content []byte
}

// UpdateDocument updates an existing document
func (s *DocumentService) UpdateDocument(ctx context.Context, input UpdateDocumentInput) (*model.Document, error) {
	doc, err := s.docRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Update content in MinIO
	if input.Content != nil {
		if err := s.contentRepo.Save(ctx, doc.ContentID, input.Content, "text/html"); err != nil {
			s.log.Error("failed to update document content", "error", err)
			return nil, err
		}

		// Regenerate thumbnail asynchronously when content changes
		if s.thumbnailSvc != nil && len(input.Content) > 0 {
			go s.generateAndSaveThumbnail(context.Background(), doc.ID, input.Content)
		}
	}

	// Update document record
	if input.Title != "" {
		doc.Title = input.Title
	}
	doc.UpdatedAt = time.Now()

	if err := s.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}

	return doc, nil
}

// DeleteDocument deletes a document
func (s *DocumentService) DeleteDocument(ctx context.Context, id string) error {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete content from MinIO
	if err := s.contentRepo.Delete(ctx, doc.ContentID); err != nil {
		s.log.Error("failed to delete document content", "error", err)
		// Continue with document deletion
	}

	// Delete thumbnail from MinIO
	if doc.ThumbnailID != nil && *doc.ThumbnailID != "" {
		if err := s.contentRepo.Delete(ctx, *doc.ThumbnailID); err != nil {
			s.log.Error("failed to delete document thumbnail", "error", err)
			// Continue with document deletion
		}
	}

	return s.docRepo.Delete(ctx, id)
}

// ListDocuments lists documents with pagination.
// If ownerID is non-empty, only documents owned by that user are returned.
func (s *DocumentService) ListDocuments(ctx context.Context, ownerID string, limit, offset int) ([]*model.Document, error) {
	if ownerID != "" {
		return s.docRepo.GetByOwnerID(ctx, ownerID)
	}
	return s.docRepo.List(ctx, limit, offset)
}

// GetDocumentThumbnail retrieves document thumbnail
func (s *DocumentService) GetDocumentThumbnail(ctx context.Context, id string) ([]byte, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if doc.ThumbnailID == nil || *doc.ThumbnailID == "" {
		return nil, nil // No thumbnail available
	}

	return s.contentRepo.Get(ctx, *doc.ThumbnailID)
}

// generateAndSaveThumbnail generates a thumbnail and saves it to storage
func (s *DocumentService) generateAndSaveThumbnail(ctx context.Context, docID string, content []byte) {
	if s.thumbnailSvc == nil {
		return
	}

	// Generate thumbnail
	thumbnail, err := s.thumbnailSvc.GenerateFromHTML(ctx, content)
	if err != nil {
		s.log.Error("failed to generate thumbnail", "error", err, "doc_id", docID)
		return
	}

	// Get the document to check if it still exists
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		s.log.Error("document not found for thumbnail", "error", err, "doc_id", docID)
		return
	}

	// Generate thumbnail ID or reuse existing one
	thumbnailID := ""
	if doc.ThumbnailID != nil {
		thumbnailID = *doc.ThumbnailID
	}
	if thumbnailID == "" {
		thumbnailID = uuid.New().String()
	}

	// Save thumbnail to MinIO
	if err := s.contentRepo.Save(ctx, thumbnailID, thumbnail, "image/png"); err != nil {
		s.log.Error("failed to save thumbnail", "error", err, "doc_id", docID)
		return
	}

	// Update document with thumbnail ID
	doc.ThumbnailID = &thumbnailID
	doc.UpdatedAt = time.Now()
	if err := s.docRepo.Update(ctx, doc); err != nil {
		s.log.Error("failed to update document with thumbnail", "error", err, "doc_id", docID)
		return
	}

	s.log.Info("thumbnail generated and saved", "doc_id", docID, "thumbnail_id", thumbnailID)
}
