package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/store"
	"github.com/sraj/everest/pkg/dbx"
)

type userProfileStore struct {
	db *dbx.DB
}

func NewUserProfileStore(db *dbx.DB) store.UserProfileStore {
	return &userProfileStore{db: db}
}

func (r *userProfileStore) Get(ctx context.Context, userID string) (*model.UserProfile, error) {
	var profile model.UserProfile
	err := r.db.Select("user_id", "nickname", "preferences", "created_at", "updated_at").
		From("user_profiles").
		Where(dbx.Cond.Eq("user_id", userID)).
		One(ctx, &profile)
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}
	return &profile, nil
}

func (r *userProfileStore) Upsert(ctx context.Context, profile *model.UserProfile) error {
	now := time.Now()
	prefsJSON, _ := json.Marshal(profile.Preferences)

	_, err := r.db.RawExec(ctx, `
		INSERT INTO user_profiles (user_id, nickname, preferences, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			nickname = EXCLUDED.nickname,
			preferences = EXCLUDED.preferences,
			updated_at = EXCLUDED.updated_at`,
		profile.UserID, profile.Nickname, prefsJSON, now, now,
	)
	if err != nil {
		return fmt.Errorf("upsert user profile: %w", err)
	}
	profile.CreatedAt = now
	profile.UpdatedAt = now
	return nil
}
