package tests

import (
	"testing"

	. "github.com/onsi/gomega"

	postgresRepo "github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
)

func TestInteractionService_LikePost(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	likeRepo := postgresRepo.NewLikeRepository(containers.DB)
	commentRepo := postgresRepo.NewCommentRepository(containers.DB)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, containers.Cache)

	// Given: User and post exist
	user := createTestUser(t, containers.DB, "likeuser", "like@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Post to Like", "This post will be liked")

	// When: User likes the post
	err := interactionService.LikePost(user.ID, post.ID)

	// Then: Like is created successfully
	Expect(err).NotTo(HaveOccurred())

	// Verify like was saved to database
	likes, err := likeRepo.FindByPostID(post.ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(likes).To(HaveLen(1))
	Expect(likes[0].UserID).To(Equal(user.ID))
	Expect(likes[0].PostID).To(Equal(post.ID))

	// Verify post like count was incremented
	updatedPost, err := postgresRepo.NewPostRepository(containers.DB).FindByID(post.ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(updatedPost.LikesCount).To(Equal(1))
}

func TestInteractionService_LikePostDuplicate(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	likeRepo := postgresRepo.NewLikeRepository(containers.DB)
	commentRepo := postgresRepo.NewCommentRepository(containers.DB)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, containers.Cache)

	// Given: User and post exist
	user := createTestUser(t, containers.DB, "duplicateuser", "duplicate@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Post to Like Twice", "This post will be liked twice")

	// When: User likes the post first time
	err1 := interactionService.LikePost(user.ID, post.ID)
	Expect(err1).NotTo(HaveOccurred())

	// When: User tries to like the same post again
	err2 := interactionService.LikePost(user.ID, post.ID)

	// Then: Second like fails due to unique constraint
	Expect(err2).To(HaveOccurred())
	Expect(err2.Error()).To(ContainSubstring("duplicate key"))

	// Verify only one like exists
	likes, err := likeRepo.FindByPostID(post.ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(likes).To(HaveLen(1))

	// Verify post like count is still 1 (not incremented twice)
	updatedPost, err := postgresRepo.NewPostRepository(containers.DB).FindByID(post.ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(updatedPost.LikesCount).To(Equal(1))
}

func TestInteractionService_LikePostNonExistentPost(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	likeRepo := postgresRepo.NewLikeRepository(containers.DB)
	commentRepo := postgresRepo.NewCommentRepository(containers.DB)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, containers.Cache)

	// Given: User exists but post doesn't
	user := createTestUser(t, containers.DB, "nonexistentuser", "nonexistent@example.com")
	nonExistentPostID := uint(99999)

	// When: User tries to like non-existent post
	err := interactionService.LikePost(user.ID, nonExistentPostID)

	// Then: Like fails due to foreign key constraint
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("foreign key"))
}

func TestInteractionService_LikePostNonExistentUser(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	likeRepo := postgresRepo.NewLikeRepository(containers.DB)
	commentRepo := postgresRepo.NewCommentRepository(containers.DB)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, containers.Cache)

	// Given: Post exists but user doesn't
	user := createTestUser(t, containers.DB, "postowner", "owner@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Post to Like", "This post will be liked")
	nonExistentUserID := uint(99999)

	// When: Non-existent user tries to like post
	err := interactionService.LikePost(nonExistentUserID, post.ID)

	// Then: Like fails due to foreign key constraint
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("foreign key"))
}

func TestInteractionService_CommentPost(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	likeRepo := postgresRepo.NewLikeRepository(containers.DB)
	commentRepo := postgresRepo.NewCommentRepository(containers.DB)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, containers.Cache)

	// Given: User and post exist
	user := createTestUser(t, containers.DB, "commentuser", "comment@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Post to Comment", "This post will be commented on")
	commentText := "This is a great post!"

	// When: User comments on the post
	err := interactionService.CommentPost(user.ID, post.ID, commentText)

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
	updatedPost, err := postgresRepo.NewPostRepository(containers.DB).FindByID(post.ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(updatedPost.CommentsCount).To(Equal(1))
}

func TestInteractionService_CommentPostEmptyText(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	likeRepo := postgresRepo.NewLikeRepository(containers.DB)
	commentRepo := postgresRepo.NewCommentRepository(containers.DB)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, containers.Cache)

	// Given: User and post exist
	user := createTestUser(t, containers.DB, "emptycomment", "empty@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Post to Comment", "This post will be commented on")

	// When: User tries to comment with empty text
	err := interactionService.CommentPost(user.ID, post.ID, "")

	// Then: Comment creation succeeds (database allows empty text)
	Expect(err).NotTo(HaveOccurred())

	// Verify comment was saved
	comments, err := commentRepo.FindByPostID(post.ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(comments).To(HaveLen(1))
	Expect(comments[0].Text).To(Equal(""))

	// Verify post comment count was incremented
	updatedPost, err := postgresRepo.NewPostRepository(containers.DB).FindByID(post.ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(updatedPost.CommentsCount).To(Equal(1))
}

func TestInteractionService_CommentPostLongText(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	likeRepo := postgresRepo.NewLikeRepository(containers.DB)
	commentRepo := postgresRepo.NewCommentRepository(containers.DB)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, containers.Cache)

	// Given: User and post exist
	user := createTestUser(t, containers.DB, "longcomment", "long@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Post to Comment", "This post will be commented on")

	// Given: Very long comment text (1000 characters)
	longCommentText := ""
	for i := 0; i < 100; i++ {
		longCommentText += "This is a very long comment that will test the database limits. "
	}

	// When: User comments with long text
	err := interactionService.CommentPost(user.ID, post.ID, longCommentText)

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
}

func TestInteractionService_CommentPostNonExistentPost(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	likeRepo := postgresRepo.NewLikeRepository(containers.DB)
	commentRepo := postgresRepo.NewCommentRepository(containers.DB)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, containers.Cache)

	// Given: User exists but post doesn't
	user := createTestUser(t, containers.DB, "nonexistentcomment", "nonexistent@example.com")
	nonExistentPostID := uint(99999)

	// When: User tries to comment on non-existent post
	err := interactionService.CommentPost(user.ID, nonExistentPostID, "This comment will fail")

	// Then: Comment fails due to foreign key constraint
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("foreign key"))
}

func TestInteractionService_MultipleUsersLikeSamePost(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	likeRepo := postgresRepo.NewLikeRepository(containers.DB)
	commentRepo := postgresRepo.NewCommentRepository(containers.DB)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, containers.Cache)

	// Given: Multiple users and one post
	user1 := createTestUser(t, containers.DB, "user1", "user1@example.com")
	user2 := createTestUser(t, containers.DB, "user2", "user2@example.com")
	user3 := createTestUser(t, containers.DB, "user3", "user3@example.com")
	post := createTestPost(t, containers.DB, user1.ID, "Popular Post", "This post will be liked by multiple users")

	// When: Multiple users like the same post
	err1 := interactionService.LikePost(user1.ID, post.ID)
	err2 := interactionService.LikePost(user2.ID, post.ID)
	err3 := interactionService.LikePost(user3.ID, post.ID)

	// Then: All likes succeed
	Expect(err1).NotTo(HaveOccurred())
	Expect(err2).NotTo(HaveOccurred())
	Expect(err3).NotTo(HaveOccurred())

	// Verify all likes were saved
	likes, err := likeRepo.FindByPostID(post.ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(likes).To(HaveLen(3))

	// Verify post like count is correct
	updatedPost, err := postgresRepo.NewPostRepository(containers.DB).FindByID(post.ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(updatedPost.LikesCount).To(Equal(3))
}

func TestInteractionService_MultipleCommentsOnSamePost(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	likeRepo := postgresRepo.NewLikeRepository(containers.DB)
	commentRepo := postgresRepo.NewCommentRepository(containers.DB)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, containers.Cache)

	// Given: Multiple users and one post
	user1 := createTestUser(t, containers.DB, "commenter1", "commenter1@example.com")
	user2 := createTestUser(t, containers.DB, "commenter2", "commenter2@example.com")
	user3 := createTestUser(t, containers.DB, "commenter3", "commenter3@example.com")
	post := createTestPost(t, containers.DB, user1.ID, "Discussion Post", "This post will have multiple comments")

	// When: Multiple users comment on the same post
	err1 := interactionService.CommentPost(user1.ID, post.ID, "First comment!")
	err2 := interactionService.CommentPost(user2.ID, post.ID, "Second comment!")
	err3 := interactionService.CommentPost(user3.ID, post.ID, "Third comment!")

	// Then: All comments succeed
	Expect(err1).NotTo(HaveOccurred())
	Expect(err2).NotTo(HaveOccurred())
	Expect(err3).NotTo(HaveOccurred())

	// Verify all comments were saved
	comments, err := commentRepo.FindByPostID(post.ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(comments).To(HaveLen(3))

	// Verify post comment count is correct
	updatedPost, err := postgresRepo.NewPostRepository(containers.DB).FindByID(post.ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(updatedPost.CommentsCount).To(Equal(3))
}
