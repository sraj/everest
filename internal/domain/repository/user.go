package repository

import (
	"context"

	"github.com/sraj/everest/internal/domain/model"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
}
