package postgres

import (
	"context"
	"fmt"

	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/domain/repository"
	"github.com/sraj/everest/pkg/dbx"
)

type documentRepository struct {
	db *dbx.DB
}

// NewDocumentRepository creates a new PostgreSQL document repository
func NewDocumentRepository(db *dbx.DB) repository.DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(ctx context.Context, doc *model.Document) error {
	_, err := r.db.Insert("documents").
		Columns("id", "title", "owner_id", "content_id", "thumbnail_id", "created_at", "updated_at").
		Values(doc.ID, doc.Title, doc.OwnerID, doc.ContentID, nullString(thumbnailVal(doc.ThumbnailID)), doc.CreatedAt, doc.UpdatedAt).
		Exec(ctx)
	return err
}

func (r *documentRepository) GetByID(ctx context.Context, id string) (*model.Document, error) {
	var doc model.Document
	err := r.db.Select("id", "title", "owner_id", "content_id", "thumbnail_id", "created_at", "updated_at").
		From("documents").
		Where(dbx.Cond.Eq("id", id)).
		One(ctx, &doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *documentRepository) GetByOwnerID(ctx context.Context, ownerID string) ([]*model.Document, error) {
	var docs []*model.Document
	err := r.db.Select("id", "title", "owner_id", "content_id", "thumbnail_id", "created_at", "updated_at").
		From("documents").
		Where(dbx.Cond.Eq("owner_id", ownerID)).
		OrderBy("updated_at", dbx.DESC).
		All(ctx, &docs)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *documentRepository) Update(ctx context.Context, doc *model.Document) error {
	err := r.db.Update("documents").
		Set("title", doc.Title).
		Set("thumbnail_id", nullString(thumbnailVal(doc.ThumbnailID))).
		Set("updated_at", doc.UpdatedAt).
		Where(dbx.Cond.Eq("id", doc.ID)).
		ExecMustAffect(ctx)
	return err
}

func (r *documentRepository) Delete(ctx context.Context, id string) error {
	err := r.db.Delete("documents").
		Where(dbx.Cond.Eq("id", id)).
		ExecMustAffect(ctx)
	return err
}

func (r *documentRepository) List(ctx context.Context, page model.Page) (*model.PageResult, error) {
	dbPage := dbx.Page{Number: page.Number, Size: page.Size}

	total, err := r.db.Select("id").From("documents").Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("list documents count: %w", err)
	}

	var docs []*model.Document
	q := r.db.Select("id", "title", "owner_id", "content_id", "thumbnail_id", "created_at", "updated_at").
		From("documents").
		OrderBy("updated_at", dbx.DESC).
		Paginate(dbPage)
	if err := q.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}

	totalPages := 0
	if dbPage.Size > 0 {
		totalPages = (total + dbPage.Size - 1) / dbPage.Size
	}
	if dbPage.Number < 1 {
		dbPage.Number = 1
	}

	return &model.PageResult{
		Items:      docs,
		Total:      total,
		Page:       dbPage.Number,
		PageSize:   dbPage.Size,
		TotalPages: totalPages,
	}, nil
}

func (r *documentRepository) Count(ctx context.Context) (int, error) {
	return r.db.Select("id").From("documents").Count(ctx)
}

func nullString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func thumbnailVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
