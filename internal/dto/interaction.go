package dto

type CommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=500"`
}

type LikeResponse struct {
	PostID     uint `json:"post_id"`
	LikesCount int  `json:"likes_count"`
}

type CommentResponse struct {
	PostID        uint `json:"post_id"`
	CommentsCount int  `json:"comments_count"`
}
