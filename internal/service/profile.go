package service

import (
	"context"
	"log/slog"

	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/store"
)

// ProfileService manages app-specific user profile data.
type ProfileService struct {
	store store.Store
	log   *slog.Logger
}

func NewProfileService(st store.Store, log *slog.Logger) *ProfileService {
	return &ProfileService{store: st, log: log}
}

// Get returns the profile for the given user, creating one if it doesn't exist.
func (s *ProfileService) Get(ctx context.Context, userID string) (*model.UserProfile, error) {
	return s.store.Profile().Get(ctx, userID)
}

// Update modifies the app-specific profile fields (nickname, preferences).
func (s *ProfileService) Update(ctx context.Context, userID string, nickname string, prefs model.Prefs) (*model.UserProfile, error) {
	profile := &model.UserProfile{
		UserID:      userID,
		Nickname:    nickname,
		Preferences: prefs,
	}
	if err := s.store.Profile().Upsert(ctx, profile); err != nil {
		s.log.Error("failed to upsert profile", "error", err.Error())
		return nil, err
	}
	return profile, nil
}
