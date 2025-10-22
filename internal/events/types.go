package events

// EventType represents the type of event that occurred
type EventType string

const (
	EventTypeNewPost       EventType = "new_post"
	EventTypePostLiked     EventType = "post_liked"
	EventTypePostCommented EventType = "post_commented"
	EventTypePostDeleted   EventType = "post_deleted"
	EventTypeConnected     EventType = "connected"
	EventTypeHeartbeat     EventType = "heartbeat"
)

// Event represents a real-time event that can be broadcast to clients
type Event struct {
	Type              EventType   `json:"type"`
	PostID            uint        `json:"post_id"`
	TriggeredByUserID uint        `json:"triggered_by_user_id"`
	Data              interface{} `json:"data"`
	Timestamp         int64       `json:"timestamp"`
}

// NewPostData contains the post information for new_post events
type NewPostData struct {
	Post interface{} `json:"post"`
}

// PostInteractionData contains interaction counts for like/comment events
type PostInteractionData struct {
	LikesCount    int      `json:"likes_count"`
	CommentsCount int      `json:"comments_count"`
	Comment       *Comment `json:"comment,omitempty"` // Include for post_commented events
}

// Comment represents a comment in events
type Comment struct {
	ID        uint   `json:"id"`
	Text      string `json:"text"`
	Username  string `json:"username"`
	UserID    uint   `json:"user_id"`
	CreatedAt string `json:"created_at"`
}
