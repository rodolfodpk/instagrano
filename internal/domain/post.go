package domain

import "time"

type Post struct {
    ID            uint      `json:"id"`
    UserID        uint      `json:"user_id"`
    Title         string    `json:"title"`
    Caption       string    `json:"caption"`
    MediaType     MediaType `json:"media_type"`
    MediaURL      string    `json:"media_url"`
    LikesCount    int       `json:"likes_count"`
    CommentsCount int       `json:"comments_count"`
    ViewsCount    int       `json:"views_count"`
    Score         float64   `json:"score"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

type MediaType string

const (
    MediaTypeImage MediaType = "image"
    MediaTypeVideo MediaType = "video"
)
