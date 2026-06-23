package store

import (
	"context"

	"github.com/sraj/everest/internal/domain/model"
)

// DocumentStore defines the interface for document persistence.
type DocumentStore interface {
	Create(ctx context.Context, doc *model.Document) error
	GetByID(ctx context.Context, id string) (*model.Document, error)
	GetByOwnerID(ctx context.Context, ownerID string) ([]*model.Document, error)
	Update(ctx context.Context, doc *model.Document) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page model.Page) (*model.PageResult, error)
	Count(ctx context.Context) (int, error)
}
