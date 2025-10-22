package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/events"
	postgresRepo "github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
	"go.uber.org/zap"
)

// Helper function to create interaction service with all required dependencies
func createInteractionService() (*service.InteractionService, postgresRepo.LikeRepository, postgresRepo.CommentRepository, postgresRepo.PostRepository) {
	likeRepo := postgresRepo.NewLikeRepository(sharedContainers.DB)
	commentRepo := postgresRepo.NewCommentRepository(sharedContainers.DB)
	postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	eventPublisher := events.NewPublisher(sharedContainers.Cache, logger)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, postRepo, sharedContainers.Cache, eventPublisher, logger)
	return interactionService, likeRepo, commentRepo, postRepo
}

var _ = Describe("InteractionService", func() {
	Describe("LikePost", func() {
		It("should like post successfully", func() {
			// Given: Interaction service setup
			interactionService, likeRepo, _, postRepo := createInteractionService()

			// Given: User and post exist
			user := createTestUser(sharedContainers.DB, "likeuser", "like@example.com")
			post := createTestPost(sharedContainers.DB, user.ID, "Post to Like", "This post will be liked")

			// When: User likes the post
			_, _, err := interactionService.LikePost(user.ID, post.ID)

			// Then: Like is created successfully
			Expect(err).NotTo(HaveOccurred())

			// Verify like was saved to database
			likes, err := likeRepo.FindByPostID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(likes).To(HaveLen(1))
			Expect(likes[0].UserID).To(Equal(user.ID))
			Expect(likes[0].PostID).To(Equal(post.ID))

			// Verify post like count was incremented
			updatedPost, err := postRepo.FindByID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedPost.LikesCount).To(Equal(1))
		})

		It("should prevent duplicate likes", func() {
			// Given: Interaction service setup
			interactionService, likeRepo, _, postRepo := createInteractionService()

			// Given: User and post exist
			user := createTestUser(sharedContainers.DB, "duplicateuser", "duplicate@example.com")
			post := createTestPost(sharedContainers.DB, user.ID, "Post to Like Twice", "This post will be liked twice")

			// When: User likes the post first time
			_, _, err1 := interactionService.LikePost(user.ID, post.ID)
			Expect(err1).NotTo(HaveOccurred())

			// When: User tries to like the same post again
			_, _, err2 := interactionService.LikePost(user.ID, post.ID)

			// Then: Second like succeeds (it becomes an unlike due to toggle behavior)
			Expect(err2).NotTo(HaveOccurred())

			// Verify no likes exist (first like was undone)
			likes, err := likeRepo.FindByPostID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(likes).To(HaveLen(0))

			// Verify post like count is 0 (unliked)
			updatedPost, err := postRepo.FindByID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedPost.LikesCount).To(Equal(0))
		})

		It("should fail when liking non-existent post", func() {
			// Given: Interaction service setup
			interactionService, _, _, _ := createInteractionService()

			// Given: User exists but post doesn't
			user := createTestUser(sharedContainers.DB, "nonexistentuser", "nonexistent@example.com")
			nonExistentPostID := uint(99999)

			// When: User tries to like non-existent post
			_, _, err := interactionService.LikePost(user.ID, nonExistentPostID)

			// Then: Like fails due to foreign key constraint
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("foreign key"))
		})

		It("should fail when non-existent user likes post", func() {
			// Given: Interaction service setup
			interactionService, _, _, _ := createInteractionService()

			// Given: Post exists but user doesn't
			user := createTestUser(sharedContainers.DB, "postowner", "owner@example.com")
			post := createTestPost(sharedContainers.DB, user.ID, "Post to Like", "This post will be liked")
			nonExistentUserID := uint(99999)

			// When: Non-existent user tries to like post
			_, _, err := interactionService.LikePost(nonExistentUserID, post.ID)

			// Then: Like fails due to foreign key constraint
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("foreign key"))
		})

		It("should handle multiple users liking same post", func() {
			// Given: Interaction service setup
			interactionService, likeRepo, _, postRepo := createInteractionService()

			// Given: Multiple users and one post
			user1 := createTestUser(sharedContainers.DB, "user1", "user1@example.com")
			user2 := createTestUser(sharedContainers.DB, "user2", "user2@example.com")
			user3 := createTestUser(sharedContainers.DB, "user3", "user3@example.com")
			post := createTestPost(sharedContainers.DB, user1.ID, "Popular Post", "This post will be liked by multiple users")

			// When: Multiple users like the same post
			_, _, err1 := interactionService.LikePost(user1.ID, post.ID)
			_, _, err2 := interactionService.LikePost(user2.ID, post.ID)
			_, _, err3 := interactionService.LikePost(user3.ID, post.ID)

			// Then: All likes succeed
			Expect(err1).NotTo(HaveOccurred())
			Expect(err2).NotTo(HaveOccurred())
			Expect(err3).NotTo(HaveOccurred())

			// Verify all likes were saved
			likes, err := likeRepo.FindByPostID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(likes).To(HaveLen(3))

			// Verify post like count is correct
			updatedPost, err := postRepo.FindByID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedPost.LikesCount).To(Equal(3))
		})
	})

	Describe("CommentPost", func() {
		It("should comment on post successfully", func() {
			// Given: Interaction service setup
			interactionService, _, commentRepo, postRepo := createInteractionService()

			// Given: User and post exist
			user := createTestUser(sharedContainers.DB, "commentuser", "comment@example.com")
			post := createTestPost(sharedContainers.DB, user.ID, "Post to Comment", "This post will be commented on")
			commentText := "This is a great post!"

			// When: User comments on the post
			_, _, err := interactionService.CommentPost(user.ID, post.ID, commentText, user.Username)

			// Then: Comment is created successfully
			Expect(err).NotTo(HaveOccurred())

			// Verify comment was saved to database
			comments, err := commentRepo.FindByPostID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(comments).To(HaveLen(1))
			Expect(comments[0].UserID).To(Equal(user.ID))
			Expect(comments[0].PostID).To(Equal(post.ID))
			Expect(comments[0].Text).To(Equal(commentText))

			// Verify post comment count was incremented
			updatedPost, err := postRepo.FindByID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedPost.CommentsCount).To(Equal(1))
		})

		It("should allow empty comment text", func() {
			// Given: Interaction service setup
			interactionService, _, commentRepo, postRepo := createInteractionService()

			// Given: User and post exist
			user := createTestUser(sharedContainers.DB, "emptycomment", "empty@example.com")
			post := createTestPost(sharedContainers.DB, user.ID, "Post to Comment", "This post will be commented on")

			// When: User tries to comment with empty text
			_, _, err := interactionService.CommentPost(user.ID, post.ID, "", user.Username)

			// Then: Comment creation succeeds (database allows empty text)
			Expect(err).NotTo(HaveOccurred())

			// Verify comment was saved
			comments, err := commentRepo.FindByPostID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(comments).To(HaveLen(1))
			Expect(comments[0].Text).To(Equal(""))

			// Verify post comment count was incremented
			updatedPost, err := postRepo.FindByID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedPost.CommentsCount).To(Equal(1))
		})

		It("should handle long comment text", func() {
			// Given: Interaction service setup
			interactionService, _, commentRepo, _ := createInteractionService()

			// Given: User and post exist
			user := createTestUser(sharedContainers.DB, "longcomment", "long@example.com")
			post := createTestPost(sharedContainers.DB, user.ID, "Post to Comment", "This post will be commented on")

			// Given: Very long comment text (1000 characters)
			longCommentText := ""
			for i := 0; i < 100; i++ {
				longCommentText += "This is a very long comment that will test the database limits. "
			}

			// When: User comments with long text
			_, _, err := interactionService.CommentPost(user.ID, post.ID, longCommentText, user.Username)

			// Then: Comment is created successfully (if within DB limits)
			if err != nil {
				// If it fails, it should be due to length constraints
				Expect(err.Error()).To(ContainSubstring("too long"))
			} else {
				// If it succeeds, verify it was saved correctly
				comments, err := commentRepo.FindByPostID(post.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(comments).To(HaveLen(1))
				Expect(comments[0].Text).To(Equal(longCommentText))
			}
		})

		It("should fail when commenting on non-existent post", func() {
			// Given: Interaction service setup
			interactionService, _, _, _ := createInteractionService()

			// Given: User exists but post doesn't
			user := createTestUser(sharedContainers.DB, "nonexistentcomment", "nonexistent@example.com")
			nonExistentPostID := uint(99999)

			// When: User tries to comment on non-existent post
			_, _, err := interactionService.CommentPost(user.ID, nonExistentPostID, "This comment will fail", user.Username)

			// Then: Comment fails due to foreign key constraint
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("foreign key"))
		})

		It("should handle multiple users commenting on same post", func() {
			// Given: Interaction service setup
			interactionService, _, commentRepo, postRepo := createInteractionService()

			// Given: Multiple users and one post
			user1 := createTestUser(sharedContainers.DB, "commenter1", "commenter1@example.com")
			user2 := createTestUser(sharedContainers.DB, "commenter2", "commenter2@example.com")
			user3 := createTestUser(sharedContainers.DB, "commenter3", "commenter3@example.com")
			post := createTestPost(sharedContainers.DB, user1.ID, "Discussion Post", "This post will have multiple comments")

			// When: Multiple users comment on the same post
			_, _, err1 := interactionService.CommentPost(user1.ID, post.ID, "First comment!", user1.Username)
			_, _, err2 := interactionService.CommentPost(user2.ID, post.ID, "Second comment!", user2.Username)
			_, _, err3 := interactionService.CommentPost(user3.ID, post.ID, "Third comment!", user3.Username)

			// Then: All comments succeed
			Expect(err1).NotTo(HaveOccurred())
			Expect(err2).NotTo(HaveOccurred())
			Expect(err3).NotTo(HaveOccurred())

			// Verify all comments were saved
			comments, err := commentRepo.FindByPostID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(comments).To(HaveLen(3))

			// Verify post comment count is correct
			updatedPost, err := postRepo.FindByID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedPost.CommentsCount).To(Equal(3))
		})
	})
})
