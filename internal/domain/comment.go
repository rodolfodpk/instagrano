package domain

import "time"

type Comment struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	PostID    uint      `json:"post_id"`
	Text      string    `json:"text"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}
