package service

import (
	"time"

	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/repository/postgres"
)

type PostViewService struct {
	viewRepo postgres.PostViewRepository
}

func NewPostViewService(viewRepo postgres.PostViewRepository) *PostViewService {
	return &PostViewService{
		viewRepo: viewRepo,
	}
}

func (s *PostViewService) StartView(userID, postID uint) (*domain.PostView, error) {
	view := &domain.PostView{
		UserID:    userID,
		PostID:    postID,
		StartedAt: time.Now(),
	}

	if err := s.viewRepo.StartView(view); err != nil {
		return nil, err
	}

	// Increment views count (best effort - don't fail if this fails)
	if err := s.viewRepo.IncrementPostViewsCount(postID); err != nil {
		// Log error but don't return it - view tracking is more important than counter
	}

	return view, nil
}

func (s *PostViewService) EndView(userID, postID uint, startedAt, endedAt time.Time) error {
	// Best effort - don't fail if view not found (user might have refreshed page)
	return s.viewRepo.EndView(userID, postID, startedAt, endedAt)
}
