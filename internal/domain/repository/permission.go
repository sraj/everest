package repository

import (
	"context"

	"github.com/sraj/everest/internal/domain/model"
)

// DocumentPermissionRepository defines the interface for document permissions
type DocumentPermissionRepository interface {
	Create(ctx context.Context, perm *model.DocumentPermission) error
	GetByDocumentID(ctx context.Context, documentID string) ([]*model.DocumentPermission, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.DocumentPermission, error)
	Delete(ctx context.Context, id string) error
}
