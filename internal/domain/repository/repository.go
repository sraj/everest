package repository

import (
	"context"

	"github.com/sraj/everest/internal/domain/model"
)

// DocumentRepository defines the interface for document persistence
type DocumentRepository interface {
	Create(ctx context.Context, doc *model.Document) error
	GetByID(ctx context.Context, id string) (*model.Document, error)
	GetByOwnerID(ctx context.Context, ownerID string) ([]*model.Document, error)
	Update(ctx context.Context, doc *model.Document) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*model.Document, error)
}

// UserRepository defines the interface for user persistence
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
}

// DocumentPermissionRepository defines the interface for document permissions
type DocumentPermissionRepository interface {
	Create(ctx context.Context, perm *model.DocumentPermission) error
	GetByDocumentID(ctx context.Context, documentID string) ([]*model.DocumentPermission, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.DocumentPermission, error)
	Delete(ctx context.Context, id string) error
}

// ContentRepository defines the interface for document content storage (MinIO)
type ContentRepository interface {
	Save(ctx context.Context, key string, content []byte, contentType string) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}
