package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/sraj/everest/internal/apperror"
	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/pkg/jobs"
	"github.com/sraj/everest/internal/store"
)

var htmlPolicy = bluemonday.UGCPolicy()

// DocumentService defines the interface for document business logic
type DocumentService interface {
	Create(ctx context.Context, input CreateDocumentInput) (*model.Document, error)
	GetByID(ctx context.Context, id string) (*model.Document, error)
	GetContent(ctx context.Context, id string) ([]byte, error)
	Update(ctx context.Context, input UpdateDocumentInput) (*model.Document, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page model.Page) (*model.PageResult, error)
	GetThumbnail(ctx context.Context, id string) ([]byte, error)
}

type documentService struct {
	store        store.Store
	thumbnailSvc ThumbnailService
	taggerSvc    TaggerService
	pool         *jobs.Pool
	log          *slog.Logger
}

// NewDocumentService creates a new document service.
func NewDocumentService(st store.Store, thumbnailSvc ThumbnailService, taggerSvc TaggerService, pool *jobs.Pool, log *slog.Logger) DocumentService {
	return &documentService{
		store:        st,
		thumbnailSvc: thumbnailSvc,
		taggerSvc:    taggerSvc,
		pool:         pool,
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

// Create creates a new document
func (s *documentService) Create(ctx context.Context, input CreateDocumentInput) (*model.Document, error) {
	id := uuid.New().String()
	contentID := uuid.New().String()
	now := time.Now()

	// Save content to MinIO
	contentType := input.ContentType
	if contentType == "" {
		contentType = "text/html"
	}
	sanitized := htmlPolicy.SanitizeBytes(input.Content)
	if err := s.store.Content().Save(ctx, contentID, sanitized, contentType); err != nil {
		s.log.Error("failed to save document content", "error", err.Error())
		return nil, err
	}

	// Create document record
	doc := &model.Document{
		ID:        id,
		Title:     input.Title,
		OwnerID:   input.OwnerID,
		ContentID: contentID,
		Tags:      model.Tags{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.store.Document().Create(ctx, doc); err != nil {
		s.log.Error("failed to create document", "error", err.Error())
		_ = s.store.Content().Delete(ctx, contentID)
		return nil, err
	}

	// Generate thumbnail asynchronously after document exists in DB
	if s.thumbnailSvc != nil && len(input.Content) > 0 {
		s.pool.Submit(func(ctx context.Context) error {
			s.generateAndSaveThumbnail(ctx, doc.ID, input.Content)
			return nil
		})
	}

	// Generate tags asynchronously
	if s.taggerSvc != nil {
		s.pool.Submit(func(ctx context.Context) error {
			s.generateAndSaveTags(ctx, doc.ID, input.Content)
			return nil
		})
	}

	return doc, nil
}

// GetByID retrieves a document by ID
func (s *documentService) GetByID(ctx context.Context, id string) (*model.Document, error) {
	doc, err := s.store.Document().GetByID(ctx, id)
	if err != nil {
		return nil, s.translateError(err, id)
	}
	return doc, nil
}

// GetContent retrieves document content
func (s *documentService) GetContent(ctx context.Context, id string) ([]byte, error) {
	doc, err := s.store.Document().GetByID(ctx, id)
	if err != nil {
		return nil, s.translateError(err, id)
	}

	return s.store.Content().Get(ctx, doc.ContentID)
}

// UpdateDocumentInput represents input for updating a document
type UpdateDocumentInput struct {
	ID      string
	Title   string
	Content []byte
}

// Update updates an existing document
func (s *documentService) Update(ctx context.Context, input UpdateDocumentInput) (*model.Document, error) {
	doc, err := s.store.Document().GetByID(ctx, input.ID)
	if err != nil {
		return nil, s.translateError(err, input.ID)
	}

	// Update content in MinIO
	if input.Content != nil {
		sanitized := htmlPolicy.SanitizeBytes(input.Content)
		if err := s.store.Content().Save(ctx, doc.ContentID, sanitized, "text/html"); err != nil {
			s.log.Error("failed to update document content", "error", err.Error())
			return nil, err
		}

		// Regenerate thumbnail asynchronously when content changes
		if s.thumbnailSvc != nil && len(input.Content) > 0 {
			s.pool.Submit(func(ctx context.Context) error {
				s.generateAndSaveThumbnail(ctx, doc.ID, input.Content)
				return nil
			})
		}

		// Regenerate tags when content changes
		if s.taggerSvc != nil {
			s.pool.Submit(func(ctx context.Context) error {
				s.generateAndSaveTags(ctx, doc.ID, input.Content)
				return nil
			})
		}
	}

	// Update document record
	if input.Title != "" {
		doc.Title = input.Title
	}
	doc.UpdatedAt = time.Now()

	if err := s.store.Document().Update(ctx, doc); err != nil {
		return nil, s.translateError(err, input.ID)
	}

	return doc, nil
}

// Delete deletes a document
func (s *documentService) Delete(ctx context.Context, id string) error {
	doc, err := s.store.Document().GetByID(ctx, id)
	if err != nil {
		return s.translateError(err, id)
	}

	// Delete content from MinIO
	if err := s.store.Content().Delete(ctx, doc.ContentID); err != nil {
		s.log.Error("failed to delete document content", "error", err.Error())
		// Continue with document deletion
	}

	// Delete thumbnail from MinIO
	if doc.ThumbnailID != nil && *doc.ThumbnailID != "" {
		if err := s.store.Content().Delete(ctx, *doc.ThumbnailID); err != nil {
			s.log.Error("failed to delete document thumbnail", "error", err.Error())
			// Continue with document deletion
		}
	}

	if err := s.store.Document().Delete(ctx, id); err != nil {
		return s.translateError(err, id)
	}
	return nil
}

// List lists documents with pagination
func (s *documentService) List(ctx context.Context, page model.Page) (*model.PageResult, error) {
	return s.store.Document().List(ctx, page)
}

// GetThumbnail retrieves document thumbnail
func (s *documentService) GetThumbnail(ctx context.Context, id string) ([]byte, error) {
	doc, err := s.store.Document().GetByID(ctx, id)
	if err != nil {
		return nil, s.translateError(err, id)
	}

	if doc.ThumbnailID == nil || *doc.ThumbnailID == "" {
		return nil, nil // No thumbnail available
	}

	return s.store.Content().Get(ctx, *doc.ThumbnailID)
}

// generateAndSaveThumbnail generates a thumbnail and saves it to storage
func (s *documentService) generateAndSaveThumbnail(ctx context.Context, docID string, content []byte) {
	if s.thumbnailSvc == nil {
		return
	}

	// Generate thumbnail
	thumbnail, err := s.thumbnailSvc.GenerateFromHTML(ctx, content)
	if err != nil {
		s.log.Error("failed to generate thumbnail", "error", err.Error(), "doc_id", docID)
		return
	}

	// Get the document to check if it still exists
	doc, err := s.store.Document().GetByID(ctx, docID)
	if err != nil {
		s.log.Error("document not found for thumbnail", "error", err.Error(), "doc_id", docID)
		return
	}

	// Generate thumbnail ID or reuse existing one
	thumbnailID := doc.ThumbnailID
	if thumbnailID == nil {
		id := uuid.New().String()
		thumbnailID = &id
	}

	// Save thumbnail to MinIO
	if err := s.store.Content().Save(ctx, *thumbnailID, thumbnail, "image/png"); err != nil {
		s.log.Error("failed to save thumbnail", "error", err.Error(), "doc_id", docID)
		return
	}

	// Update document with thumbnail ID
	doc.ThumbnailID = thumbnailID
	doc.UpdatedAt = time.Now()
	if err := s.store.Document().Update(ctx, doc); err != nil {
		s.log.Error("failed to update document with thumbnail", "error", err.Error(), "doc_id", docID)
		return
	}

	s.log.Info("thumbnail generated and saved", "doc_id", docID, "thumbnail_id", thumbnailID)
}

func (s *documentService) generateAndSaveTags(ctx context.Context, docID string, content []byte) {
	tags, err := s.taggerSvc.Generate(ctx, content)
	if err != nil {
		s.log.Error("failed to generate tags", "error", err.Error(), "doc_id", docID)
		return
	}
	if len(tags) == 0 {
		return
	}

	doc, err := s.store.Document().GetByID(ctx, docID)
	if err != nil {
		s.log.Error("document not found for tags", "error", err.Error(), "doc_id", docID)
		return
	}

	doc.Tags = tags
	doc.UpdatedAt = time.Now()
	if err := s.store.Document().Update(ctx, doc); err != nil {
		s.log.Error("failed to save tags", "error", err.Error(), "doc_id", docID)
		return
	}

	s.log.Info("tags generated", "doc_id", docID, "tags", tags)
}

func (s *documentService) translateError(err error, id string) error {
	var nfErr store.ErrNotFound
	if errors.As(err, &nfErr) {
		return apperror.NotFound("document %s not found", nfErr.ID)
	}
	var cfErr store.ErrConflict
	if errors.As(err, &cfErr) {
		return apperror.Conflict("document %s: %s already exists", cfErr.Field, cfErr.Resource)
	}
	return err
}
