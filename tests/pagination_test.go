package tests

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/pagination"
)

func TestCursor_Encode(t *testing.T) {
	RegisterTestingT(t)

	// Given: A cursor with specific values
	cursor := &pagination.Cursor{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		ID:        123,
	}

	// When: Encode the cursor
	encoded := cursor.Encode()

	// Then: Should produce a non-empty string
	Expect(encoded).NotTo(BeEmpty())
	Expect(len(encoded)).To(BeNumerically(">", 0))
}

func TestCursor_Decode(t *testing.T) {
	RegisterTestingT(t)

	// Given: A cursor with specific values
	originalCursor := &pagination.Cursor{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		ID:        123,
	}

	// When: Encode and then decode
	encoded := originalCursor.Encode()
	decodedCursor, err := pagination.DecodeCursor(encoded)

	// Then: Should decode successfully and match original values
	Expect(err).NotTo(HaveOccurred())
	Expect(decodedCursor).NotTo(BeNil())
	// Use Unix timestamp comparison to avoid timezone issues
	Expect(decodedCursor.Timestamp.Unix()).To(Equal(originalCursor.Timestamp.Unix()))
	Expect(decodedCursor.ID).To(Equal(originalCursor.ID))
}

func TestCursor_DecodeInvalid(t *testing.T) {
	RegisterTestingT(t)

	// Given: Invalid encoded cursor
	invalidEncoded := "invalid-cursor-string"

	// When: Try to decode invalid cursor
	decodedCursor, err := pagination.DecodeCursor(invalidEncoded)

	// Then: Should return an error
	Expect(err).To(HaveOccurred())
	Expect(decodedCursor).To(BeNil())
}

func TestCursor_DecodeEmpty(t *testing.T) {
	RegisterTestingT(t)

	// Given: Empty cursor string
	emptyCursor := ""

	// When: Try to decode empty cursor
	decodedCursor, err := pagination.DecodeCursor(emptyCursor)

	// Then: Should return nil cursor and nil error (current behavior)
	Expect(err).To(BeNil())
	Expect(decodedCursor).To(BeNil())
}

func TestCursor_RoundTrip(t *testing.T) {
	RegisterTestingT(t)

	// Test multiple cursor values
	testCases := []struct {
		name      string
		timestamp time.Time
		id        uint
	}{
		{
			name:      "zero values",
			timestamp: time.Time{},
			id:        0,
		},
		{
			name:      "small values",
			timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			id:        1,
		},
		{
			name:      "large values",
			timestamp: time.Date(2030, 12, 31, 23, 59, 59, 999999999, time.UTC),
			id:        999999,
		},
		{
			name:      "current time",
			timestamp: time.Now().UTC(),
			id:        42,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RegisterTestingT(t)

			// Given: A cursor with test values
			originalCursor := &pagination.Cursor{
				Timestamp: tc.timestamp,
				ID:        tc.id,
			}

			// When: Encode and decode
			encoded := originalCursor.Encode()
			decodedCursor, err := pagination.DecodeCursor(encoded)

			// Then: Should round-trip successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(decodedCursor).NotTo(BeNil())
			// Use Unix timestamp comparison to avoid timezone issues
			Expect(decodedCursor.Timestamp.Unix()).To(Equal(originalCursor.Timestamp.Unix()))
			Expect(decodedCursor.ID).To(Equal(originalCursor.ID))
		})
	}
}

func TestCursor_Uniqueness(t *testing.T) {
	RegisterTestingT(t)

	// Given: Multiple cursors with different values
	cursor1 := &pagination.Cursor{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		ID:        123,
	}
	cursor2 := &pagination.Cursor{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 1, 0, time.UTC), // 1 second later
		ID:        123,
	}
	cursor3 := &pagination.Cursor{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		ID:        124, // Different ID
	}

	// When: Encode all cursors
	encoded1 := cursor1.Encode()
	encoded2 := cursor2.Encode()
	encoded3 := cursor3.Encode()

	// Then: All encoded cursors should be unique
	Expect(encoded1).NotTo(Equal(encoded2))
	Expect(encoded1).NotTo(Equal(encoded3))
	Expect(encoded2).NotTo(Equal(encoded3))
}

func TestFeedResult_Structure(t *testing.T) {
	RegisterTestingT(t)

	// Given: A feed result
	result := &pagination.FeedResult{
		Posts:      []interface{}{"post1", "post2"},
		NextCursor: "next-cursor",
		HasMore:    true,
	}

	// Then: Should have correct structure
	Expect(result.Posts).To(HaveLen(2))
	Expect(result.NextCursor).To(Equal("next-cursor"))
	Expect(result.HasMore).To(BeTrue())
}

func TestFeedResult_EmptyFeed(t *testing.T) {
	RegisterTestingT(t)

	// Given: An empty feed result
	result := &pagination.FeedResult{
		Posts:      []interface{}{},
		NextCursor: "",
		HasMore:    false,
	}

	// Then: Should represent empty feed correctly
	Expect(result.Posts).To(BeEmpty())
	Expect(result.NextCursor).To(BeEmpty())
	Expect(result.HasMore).To(BeFalse())
}
