package dto

import (
	"time"

	"github.com/rodolfodpk/instagrano/internal/domain"
)

type PostResponse struct {
	ID            uint             `json:"id"`
	UserID        uint             `json:"user_id"`
	Username      string           `json:"username"`
	Title         string           `json:"title"`
	Caption       string           `json:"caption"`
	MediaType     domain.MediaType `json:"media_type"`
	MediaURL      string           `json:"media_url"`
	LikesCount    int              `json:"likes_count"`
	CommentsCount int              `json:"comments_count"`
	ViewsCount    int              `json:"views_count"`
	Score         float64          `json:"score"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

func ToPostResponse(post *domain.Post) *PostResponse {
	return &PostResponse{
		ID:            post.ID,
		UserID:        post.UserID,
		Username:      post.Username,
		Title:         post.Title,
		Caption:       post.Caption,
		MediaType:     post.MediaType,
		MediaURL:      post.MediaURL,
		LikesCount:    post.LikesCount,
		CommentsCount: post.CommentsCount,
		ViewsCount:    post.ViewsCount,
		Score:         post.Score,
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
	}
}
