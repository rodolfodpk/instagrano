package dto

type FeedResponse struct {
	Posts      []*PostResponse `json:"posts"`
	NextCursor string          `json:"next_cursor"`
	HasMore    bool            `json:"has_more"`
}
