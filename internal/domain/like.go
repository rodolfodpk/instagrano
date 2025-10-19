package domain

import "time"

type Like struct {
    ID        uint      `json:"id"`
    UserID    uint      `json:"user_id"`
    PostID    uint      `json:"post_id"`
    CreatedAt time.Time `json:"created_at"`
}
