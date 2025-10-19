package domain

import (
    "math"
    "time"
)

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

const (
    LikeWeight    = 2.0
    CommentWeight = 3.0
    ViewWeight    = 0.1
    DecayRate     = 0.1
)

func (p *Post) CalculateScore() float64 {
    ageHours := time.Since(p.CreatedAt).Hours()
    engagementScore := float64(p.LikesCount)*LikeWeight +
        float64(p.CommentsCount)*CommentWeight +
        float64(p.ViewsCount)*ViewWeight
    timeDecay := math.Exp(-DecayRate * ageHours)
    return engagementScore * timeDecay
}
