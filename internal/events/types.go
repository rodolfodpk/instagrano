package events

// EventType represents the type of event that occurred
type EventType string

const (
	EventTypeNewPost       EventType = "new_post"
	EventTypePostLiked     EventType = "post_liked"
	EventTypePostCommented EventType = "post_commented"
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
	LikesCount    int `json:"likes_count"`
	CommentsCount int `json:"comments_count"`
}
