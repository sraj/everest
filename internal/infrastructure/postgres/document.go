package postgres

import (
	"context"

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

func (r *documentRepository) List(ctx context.Context, limit, offset int) ([]*model.Document, error) {
	var docs []*model.Document
	err := r.db.Select("id", "title", "owner_id", "content_id", "thumbnail_id", "created_at", "updated_at").
		From("documents").
		OrderBy("updated_at", dbx.DESC).
		Paginate(dbx.Page{Number: offset/limit + 1, Size: limit}).
		All(ctx, &docs)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func thumbnailVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
