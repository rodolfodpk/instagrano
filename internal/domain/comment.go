package domain

import "time"

type Comment struct {
    ID        uint      `json:"id"`
    UserID    uint      `json:"user_id"`
    PostID    uint      `json:"post_id"`
    Text      string    `json:"text"`
    CreatedAt time.Time `json:"created_at"`
}
