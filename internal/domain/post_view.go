package domain

import "time"

type PostView struct {
	ID              uint       `json:"id"`
	UserID          uint       `json:"user_id"`
	PostID          uint       `json:"post_id"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`
}
