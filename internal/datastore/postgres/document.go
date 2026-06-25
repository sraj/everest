package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/store"
	"github.com/sraj/everest/pkg/dbx"
)

type documentStore struct {
	db *dbx.DB
}

// NewDocumentStore creates a new PostgreSQL-backed DocumentStore.
func NewDocumentStore(db *dbx.DB) store.DocumentStore {
	return &documentStore{db: db}
}

func (r *documentStore) Create(ctx context.Context, doc *model.Document) error {
	_, err := r.db.Insert("documents").
		Columns("id", "title", "owner_id", "content_id", "thumbnail_id", "created_at", "updated_at").
		Values(doc.ID, doc.Title, doc.OwnerID, doc.ContentID, nullString(doc.ThumbnailID), doc.CreatedAt, doc.UpdatedAt).
		Exec(ctx)
	if err != nil {
		if dbx.IsUniqueViolation(err) {
			return store.ErrConflict{Resource: "document", Field: "id", Err: err}
		}
		return fmt.Errorf("create document: %w", err)
	}
	return nil
}

func (r *documentStore) GetByID(ctx context.Context, id string) (*model.Document, error) {
	var doc model.Document
	err := r.db.Select("id", "title", "owner_id", "content_id", "thumbnail_id", "created_at", "updated_at").
		From("documents").
		Where(dbx.Cond.Eq("id", id)).
		One(ctx, &doc)
	if err != nil {
		if errors.Is(err, dbx.ErrNotFound) {
			return nil, store.ErrNotFound{Resource: "document", ID: id}
		}
		return nil, fmt.Errorf("get document: %w", err)
	}
	return &doc, nil
}

func (r *documentStore) GetByOwnerID(ctx context.Context, ownerID string) ([]*model.Document, error) {
	var docs []*model.Document
	err := r.db.Select("id", "title", "owner_id", "content_id", "thumbnail_id", "created_at", "updated_at").
		From("documents").
		Where(dbx.Cond.Eq("owner_id", ownerID)).
		OrderBy("updated_at", dbx.DESC).
		All(ctx, &docs)
	if err != nil {
		return nil, fmt.Errorf("get documents by owner: %w", err)
	}
	return docs, nil
}

func (r *documentStore) Update(ctx context.Context, doc *model.Document) error {
	err := r.db.Update("documents").
		Set("title", doc.Title).
		Set("thumbnail_id", nullString(doc.ThumbnailID)).
		Set("updated_at", doc.UpdatedAt).
		Where(dbx.Cond.Eq("id", doc.ID)).
		ExecMustAffect(ctx)
	if err != nil {
		if errors.Is(err, dbx.ErrNoRows) {
			return store.ErrNotFound{Resource: "document", ID: doc.ID}
		}
		if dbx.IsUniqueViolation(err) {
			return store.ErrConflict{Resource: "document", Field: "title", Err: err}
		}
		return fmt.Errorf("update document: %w", err)
	}
	return nil
}

func (r *documentStore) Delete(ctx context.Context, id string) error {
	err := r.db.Delete("documents").
		Where(dbx.Cond.Eq("id", id)).
		ExecMustAffect(ctx)
	if err != nil {
		if errors.Is(err, dbx.ErrNoRows) {
			return store.ErrNotFound{Resource: "document", ID: id}
		}
		return fmt.Errorf("delete document: %w", err)
	}
	return nil
}

func (r *documentStore) List(ctx context.Context, page model.Page) (*model.PageResult, error) {
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

func (r *documentStore) Count(ctx context.Context) (int, error) {
	return r.db.Select("id").From("documents").Count(ctx)
}

func (r *documentStore) ListByOwner(ctx context.Context, ownerID string, page model.Page) (*model.PageResult, error) {
	dbPage := dbx.Page{Number: page.Number, Size: page.Size}

	total, err := r.db.Select("id").From("documents").
		Where(dbx.Cond.Eq("owner_id", ownerID)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("list documents count by owner: %w", err)
	}

	var docs []*model.Document
	q := r.db.Select("id", "title", "owner_id", "content_id", "thumbnail_id", "created_at", "updated_at").
		From("documents").
		Where(dbx.Cond.Eq("owner_id", ownerID)).
		OrderBy("updated_at", dbx.DESC).
		Paginate(dbPage)
	if err := q.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("list documents by owner: %w", err)
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

func nullString(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}
