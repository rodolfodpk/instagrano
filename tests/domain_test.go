package tests

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"

	"github.com/rodolfodpk/instagrano/internal/domain"
)

func TestPost_CalculateScore(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Fresh post with high engagement", func(t *testing.T) {
		RegisterTestingT(t)

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
		Expect(score).To(BeNumerically(">", 300))
		Expect(score).To(BeNumerically("<", 500))
	})

	t.Run("Old post with high engagement", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: An old post with high engagement
		post := &domain.Post{
			LikesCount:    100,
			CommentsCount: 50,
			ViewsCount:    1000,
			CreatedAt:     time.Now().Add(-24 * time.Hour), // 24 hours old
		}

		// When: Calculate score
		score := post.CalculateScore()

		// Then: Score should be lower due to time decay
		Expect(score).To(BeNumerically(">", 0))
		Expect(score).To(BeNumerically("<", 50))
	})

	t.Run("Fresh post with no engagement", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: A fresh post with no engagement
		post := &domain.Post{
			LikesCount:    0,
			CommentsCount: 0,
			ViewsCount:    0,
			CreatedAt:     time.Now().Add(-1 * time.Hour), // 1 hour old
		}

		// When: Calculate score
		score := post.CalculateScore()

		// Then: Score should be zero
		Expect(score).To(Equal(0.0))
	})

	t.Run("Very old post with high engagement", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: A very old post with high engagement
		post := &domain.Post{
			LikesCount:    1000,
			CommentsCount: 500,
			ViewsCount:    10000,
			CreatedAt:     time.Now().Add(-168 * time.Hour), // 1 week old
		}

		// When: Calculate score
		score := post.CalculateScore()

		// Then: Score should be very low due to time decay
		Expect(score).To(BeNumerically(">", 0))
		Expect(score).To(BeNumerically("<", 10))
	})

	t.Run("Score comparison between fresh and old posts", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: Two posts with same engagement but different ages
		freshPost := &domain.Post{
			LikesCount:    100,
			CommentsCount: 50,
			ViewsCount:    1000,
			CreatedAt:     time.Now().Add(-1 * time.Hour), // 1 hour old
		}

		oldPost := &domain.Post{
			LikesCount:    100,
			CommentsCount: 50,
			ViewsCount:    1000,
			CreatedAt:     time.Now().Add(-24 * time.Hour), // 24 hours old
		}

		// When: Calculate scores
		freshScore := freshPost.CalculateScore()
		oldScore := oldPost.CalculateScore()

		// Then: Fresh post should score higher
		Expect(freshScore).To(BeNumerically(">", oldScore))
	})

	t.Run("Score with different engagement types", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: Posts with different engagement patterns
		likeHeavyPost := &domain.Post{
			LikesCount:    200,
			CommentsCount: 10,
			ViewsCount:    1000,
			CreatedAt:     time.Now().Add(-1 * time.Hour),
		}

		commentHeavyPost := &domain.Post{
			LikesCount:    50,
			CommentsCount: 100,
			ViewsCount:    1000,
			CreatedAt:     time.Now().Add(-1 * time.Hour),
		}

		viewHeavyPost := &domain.Post{
			LikesCount:    10,
			CommentsCount: 5,
			ViewsCount:    10000,
			CreatedAt:     time.Now().Add(-1 * time.Hour),
		}

		// When: Calculate scores
		likeScore := likeHeavyPost.CalculateScore()
		commentScore := commentHeavyPost.CalculateScore()
		viewScore := viewHeavyPost.CalculateScore()

		// Then: Comment-heavy should score highest (weight=3), then likes (weight=2), then views (weight=0.1)
		// Note: commentScore = 50*2 + 100*3 + 1000*0.1 = 100 + 300 + 100 = 500
		//       likeScore = 200*2 + 10*3 + 1000*0.1 = 400 + 30 + 100 = 530
		//       viewScore = 10*2 + 5*3 + 10000*0.1 = 20 + 15 + 1000 = 1035
		Expect(viewScore).To(BeNumerically(">", likeScore))
		Expect(likeScore).To(BeNumerically(">", commentScore))
	})

	t.Run("Edge case: post created in the future", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: A post created in the future (edge case)
		post := &domain.Post{
			LikesCount:    100,
			CommentsCount: 50,
			ViewsCount:    1000,
			CreatedAt:     time.Now().Add(1 * time.Hour), // 1 hour in the future
		}

		// When: Calculate score
		score := post.CalculateScore()

		// Then: Score should be higher than normal due to negative age
		Expect(score).To(BeNumerically(">", 250))
	})

	t.Run("Edge case: post created exactly now", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: A post created exactly now
		post := &domain.Post{
			LikesCount:    100,
			CommentsCount: 50,
			ViewsCount:    1000,
			CreatedAt:     time.Now(),
		}

		// When: Calculate score
		score := post.CalculateScore()

		// Then: Score should be very close to raw engagement score
		expectedEngagement := float64(100)*2.0 + float64(50)*3.0 + float64(1000)*0.1
		Expect(score).To(BeNumerically("~", expectedEngagement, 1.0))
	})
}

func TestUser_ValidatePassword(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Valid password", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: A user with a hashed password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		Expect(err).NotTo(HaveOccurred())

		user := &domain.User{
			Password: string(hashedPassword),
		}

		// When: Validate correct password
		err = user.ValidatePassword("password123")

		// Then: Should succeed
		Expect(err).NotTo(HaveOccurred())
	})

	t.Run("Invalid password", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: A user with a hashed password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		Expect(err).NotTo(HaveOccurred())

		user := &domain.User{
			Password: string(hashedPassword),
		}

		// When: Validate wrong password
		err = user.ValidatePassword("wrongpassword")

		// Then: Should fail
		Expect(err).To(HaveOccurred())
	})

	t.Run("Empty password", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: A user with a hashed password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		Expect(err).NotTo(HaveOccurred())

		user := &domain.User{
			Password: string(hashedPassword),
		}

		// When: Validate empty password
		err = user.ValidatePassword("")

		// Then: Should fail
		Expect(err).To(HaveOccurred())
	})

	t.Run("Malformed hash", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: A user with malformed hash
		user := &domain.User{
			Password: "invalid-hash",
		}

		// When: Validate password
		err := user.ValidatePassword("password123")

		// Then: Should fail
		Expect(err).To(HaveOccurred())
	})
}

func TestMediaType(t *testing.T) {
	RegisterTestingT(t)

	t.Run("MediaType constants", func(t *testing.T) {
		RegisterTestingT(t)

		// Then: Constants should have expected values
		Expect(string(domain.MediaTypeImage)).To(Equal("image"))
		Expect(string(domain.MediaTypeVideo)).To(Equal("video"))
	})

	t.Run("MediaType in Post struct", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: Posts with different media types
		imagePost := &domain.Post{
			MediaType: domain.MediaTypeImage,
		}

		videoPost := &domain.Post{
			MediaType: domain.MediaTypeVideo,
		}

		// Then: Media types should be correctly assigned
		Expect(imagePost.MediaType).To(Equal(domain.MediaTypeImage))
		Expect(videoPost.MediaType).To(Equal(domain.MediaTypeVideo))
	})
}

func TestScoringConstants(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Scoring weights", func(t *testing.T) {
		RegisterTestingT(t)

		// Then: Constants should have expected values
		Expect(domain.LikeWeight).To(Equal(2.0))
		Expect(domain.CommentWeight).To(Equal(3.0))
		Expect(domain.ViewWeight).To(Equal(0.1))
		Expect(domain.DecayRate).To(Equal(0.1))
	})

	t.Run("Weight impact on scoring", func(t *testing.T) {
		RegisterTestingT(t)

		// Given: Posts with same counts but different engagement types
		post := &domain.Post{
			CreatedAt: time.Now().Add(-1 * time.Hour),
		}

		// Test likes impact
		post.LikesCount = 100
		post.CommentsCount = 0
		post.ViewsCount = 0
		likeScore := post.CalculateScore()

		// Test comments impact
		post.LikesCount = 0
		post.CommentsCount = 100
		post.ViewsCount = 0
		commentScore := post.CalculateScore()

		// Test views impact
		post.LikesCount = 0
		post.CommentsCount = 0
		post.ViewsCount = 100
		viewScore := post.CalculateScore()

		// Then: Comments should have highest impact, then likes, then views
		Expect(commentScore).To(BeNumerically(">", likeScore))
		Expect(likeScore).To(BeNumerically(">", viewScore))
	})
}
