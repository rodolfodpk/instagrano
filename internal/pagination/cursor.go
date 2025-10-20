package pagination

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Cursor represents a pagination cursor
type Cursor struct {
	Timestamp time.Time
	ID        uint
}

// Encode encodes a cursor to a base64 string
func (c *Cursor) Encode() string {
	cursorStr := fmt.Sprintf("%d_%d", c.Timestamp.Unix(), c.ID)
	return base64.StdEncoding.EncodeToString([]byte(cursorStr))
}

// Decode decodes a base64 string to a cursor
func DecodeCursor(cursorStr string) (*Cursor, error) {
	if cursorStr == "" {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}

	parts := strings.Split(string(decoded), "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cursor format")
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp in cursor: %w", err)
	}

	id, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid id in cursor: %w", err)
	}

	return &Cursor{
		Timestamp: time.Unix(timestamp, 0),
		ID:        uint(id),
	}, nil
}

// FeedResult represents the result of a feed query with pagination
type FeedResult struct {
	Posts     []interface{} `json:"posts"`
	NextCursor string       `json:"next_cursor,omitempty"`
	HasMore   bool          `json:"has_more"`
}
