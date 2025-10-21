package tests

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/pagination"
)

var _ = Describe("Cursor", func() {
	Describe("Encode", func() {
		It("should encode cursor to non-empty string", func() {
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
		})

		It("should produce different encodings for different cursors", func() {
			// Given: Two different cursors
			cursor1 := &pagination.Cursor{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ID:        123,
			}
			cursor2 := &pagination.Cursor{
				Timestamp: time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC),
				ID:        124,
			}

			// When: Encode both cursors
			encoded1 := cursor1.Encode()
			encoded2 := cursor2.Encode()

			// Then: Encodings should be different
			Expect(encoded1).NotTo(Equal(encoded2))
		})
	})

	Describe("Decode", func() {
		It("should decode valid cursor string", func() {
			// Given: A valid encoded cursor
			originalCursor := &pagination.Cursor{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ID:        123,
			}
			encoded := originalCursor.Encode()

			// When: Decode the cursor
			decodedCursor, err := pagination.DecodeCursor(encoded)

			// Then: Should decode successfully and match original
			Expect(err).NotTo(HaveOccurred())
			Expect(decodedCursor.Timestamp.Unix()).To(Equal(originalCursor.Timestamp.Unix()))
			Expect(decodedCursor.ID).To(Equal(originalCursor.ID))
		})

		It("should handle invalid cursor string", func() {
			// Given: An invalid cursor string
			invalidCursor := "invalid-cursor-string"

			// When: Decode the invalid cursor
			decodedCursor, err := pagination.DecodeCursor(invalidCursor)

			// Then: Should return error
			Expect(err).To(HaveOccurred())
			Expect(decodedCursor).To(BeNil())
		})

		It("should handle empty cursor string", func() {
			// Given: An empty cursor string
			emptyCursor := ""

			// When: Decode the empty cursor
			decodedCursor, err := pagination.DecodeCursor(emptyCursor)

			// Then: Should return nil cursor with no error (as per implementation)
			Expect(err).NotTo(HaveOccurred())
			Expect(decodedCursor).To(BeNil())
		})

		It("should round-trip encode and decode", func() {
			// Given: A cursor with specific values
			originalCursor := &pagination.Cursor{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ID:        123,
			}

			// When: Encode and then decode
			encoded := originalCursor.Encode()
			decodedCursor, err := pagination.DecodeCursor(encoded)

			// Then: Should match original exactly
			Expect(err).NotTo(HaveOccurred())
			Expect(decodedCursor.Timestamp.Unix()).To(Equal(originalCursor.Timestamp.Unix()))
			Expect(decodedCursor.ID).To(Equal(originalCursor.ID))
		})
	})

	Describe("IsEmpty", func() {
		It("should return true for empty cursor", func() {
			// Given: An empty cursor
			cursor := &pagination.Cursor{}

			// When: Check if empty
			isEmpty := cursor.Timestamp.IsZero() && cursor.ID == 0

			// Then: Should be true
			Expect(isEmpty).To(BeTrue())
		})

		It("should return false for non-empty cursor", func() {
			// Given: A non-empty cursor
			cursor := &pagination.Cursor{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ID:        123,
			}

			// When: Check if empty
			isEmpty := cursor.Timestamp.IsZero() && cursor.ID == 0

			// Then: Should be false
			Expect(isEmpty).To(BeFalse())
		})
	})
})