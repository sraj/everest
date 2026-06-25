package store

import (
	"context"

	"github.com/sraj/everest/internal/domain/model"
)

// UserProfileStore defines the interface for user profile persistence.
type UserProfileStore interface {
	Get(ctx context.Context, userID string) (*model.UserProfile, error)
	Upsert(ctx context.Context, profile *model.UserProfile) error
}
