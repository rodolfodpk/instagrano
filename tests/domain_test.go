package tests

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"

	"github.com/rodolfodpk/instagrano/internal/domain"
)

var _ = Describe("Post", func() {
	Describe("CalculateScore", func() {
		It("should calculate high score for fresh post with high engagement", func() {
			// Given: A fresh post with high engagement
			post := &domain.Post{
				LikesCount:    100,
				CommentsCount: 50,
				ViewsCount:    1000,
				CreatedAt:     time.Now().Add(-1 * time.Hour), // 1 hour old
			}

			// When: Calculate score
			score := post.CalculateScore()

			// Then: Score should be high due to fresh content and engagement
			Expect(score).To(BeNumerically(">", 0))
			Expect(score).To(BeNumerically(">", 50)) // Should be high due to engagement
		})

		It("should calculate lower score for old post", func() {
			// Given: An old post
			post := &domain.Post{
				LikesCount:    100,
				CommentsCount: 50,
				ViewsCount:    1000,
				CreatedAt:     time.Now().Add(-24 * time.Hour), // 1 day old
			}

			// When: Calculate score
			score := post.CalculateScore()

			// Then: Score should be lower due to age
			Expect(score).To(BeNumerically(">", 0))
			Expect(score).To(BeNumerically("<", 50)) // Should be lower due to age
		})

		It("should calculate score based on engagement ratio", func() {
			// Given: Posts with different engagement ratios
			highEngagementPost := &domain.Post{
				LikesCount:    1000,
				CommentsCount: 200,
				ViewsCount:    1000, // High engagement ratio
				CreatedAt:     time.Now().Add(-1 * time.Hour),
			}

			lowEngagementPost := &domain.Post{
				LikesCount:    10,
				CommentsCount: 2,
				ViewsCount:    1000, // Low engagement ratio
				CreatedAt:     time.Now().Add(-1 * time.Hour),
			}

			// When: Calculate scores
			highScore := highEngagementPost.CalculateScore()
			lowScore := lowEngagementPost.CalculateScore()

			// Then: High engagement post should have higher score
			Expect(highScore).To(BeNumerically(">", lowScore))
		})

		It("should handle zero values gracefully", func() {
			// Given: A post with zero engagement
			post := &domain.Post{
				LikesCount:    0,
				CommentsCount: 0,
				ViewsCount:    0,
				CreatedAt:     time.Now().Add(-1 * time.Hour),
			}

			// When: Calculate score
			score := post.CalculateScore()

			// Then: Should not panic and return a valid score
			Expect(score).To(BeNumerically(">=", 0))
		})
	})
})

var _ = Describe("User", func() {
	Describe("ValidatePassword", func() {
		It("should validate correct password", func() {
			// Given: User with hashed password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			Expect(err).NotTo(HaveOccurred())

			user := &domain.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: string(hashedPassword),
			}

			// When: Validate correct password
			err = user.ValidatePassword("password123")

			// Then: Should validate successfully
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject incorrect password", func() {
			// Given: User with hashed password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			Expect(err).NotTo(HaveOccurred())

			user := &domain.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: string(hashedPassword),
			}

			// When: Validate incorrect password
			err = user.ValidatePassword("wrongpassword")

			// Then: Should reject password
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(bcrypt.ErrMismatchedHashAndPassword))
		})

		It("should handle empty password", func() {
			// Given: User with hashed password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			Expect(err).NotTo(HaveOccurred())

			user := &domain.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: string(hashedPassword),
			}

			// When: Validate empty password
			err = user.ValidatePassword("")

			// Then: Should reject empty password
			Expect(err).To(HaveOccurred())
		})

		It("should handle malformed password hash", func() {
			// Given: User with malformed password hash
			user := &domain.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "invalid-hash",
			}

			// When: Validate password
			err := user.ValidatePassword("password123")

			// Then: Should return error
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("User Creation", func() {
		It("should create user with valid data", func() {
			// Given: Valid user data
			user := &domain.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			}

			// Then: User should be created successfully
			Expect(user.Username).To(Equal("testuser"))
			Expect(user.Email).To(Equal("test@example.com"))
			Expect(user.Password).To(Equal("password123"))
		})

		It("should handle empty fields", func() {
			// Given: User with empty fields
			user := &domain.User{
				Username: "",
				Email:    "",
				Password: "",
			}

			// Then: User should be created (validation happens at service level)
			Expect(user.Username).To(Equal(""))
			Expect(user.Email).To(Equal(""))
			Expect(user.Password).To(Equal(""))
		})
	})
})

var _ = Describe("Comment", func() {
	Describe("Comment Creation", func() {
		It("should create comment with valid data", func() {
			// Given: Valid comment data
			comment := &domain.Comment{
				Text:   "This is a valid comment",
				UserID: 1,
				PostID: 1,
			}

			// Then: Comment should be created successfully
			Expect(comment.Text).To(Equal("This is a valid comment"))
			Expect(comment.UserID).To(Equal(uint(1)))
			Expect(comment.PostID).To(Equal(uint(1)))
		})

		It("should handle empty text", func() {
			// Given: Comment with empty text
			comment := &domain.Comment{
				Text:   "",
				UserID: 1,
				PostID: 1,
			}

			// Then: Comment should be created (validation happens at service level)
			Expect(comment.Text).To(Equal(""))
			Expect(comment.UserID).To(Equal(uint(1)))
			Expect(comment.PostID).To(Equal(uint(1)))
		})

		It("should handle zero IDs", func() {
			// Given: Comment with zero IDs
			comment := &domain.Comment{
				Text:   "Valid text",
				UserID: 0,
				PostID: 0,
			}

			// Then: Comment should be created (validation happens at service level)
			Expect(comment.Text).To(Equal("Valid text"))
			Expect(comment.UserID).To(Equal(uint(0)))
			Expect(comment.PostID).To(Equal(uint(0)))
		})
	})
})