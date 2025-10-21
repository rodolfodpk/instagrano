package tests

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
)

func TestPostViewService_StartView(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	viewRepo := postgres.NewPostViewRepository(containers.DB)
	viewService := service.NewPostViewService(viewRepo)

	// Create test user and post
	user := createTestUser(t, containers.DB, "testuser", "test@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Test Post", "Test Caption")

	// When: Start view tracking
	view, err := viewService.StartView(user.ID, post.ID)

	// Then: Should create view record and increment views_count
	Expect(err).NotTo(HaveOccurred())
	Expect(view).NotTo(BeNil())
	Expect(view.UserID).To(Equal(user.ID))
	Expect(view.PostID).To(Equal(post.ID))
	Expect(view.StartedAt).NotTo(BeZero())
	Expect(view.EndedAt).To(BeNil())
	Expect(view.DurationSeconds).To(BeNil())

	// Verify views_count was incremented
	var viewsCount int
	err = containers.DB.QueryRow("SELECT views_count FROM posts WHERE id = $1", post.ID).Scan(&viewsCount)
	Expect(err).NotTo(HaveOccurred())
	Expect(viewsCount).To(Equal(1))
}

func TestPostViewService_EndView(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	viewRepo := postgres.NewPostViewRepository(containers.DB)
	viewService := service.NewPostViewService(viewRepo)

	// Create test user and post
	user := createTestUser(t, containers.DB, "testuser", "test@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Test Post", "Test Caption")

	// Start view tracking first
	view, err := viewService.StartView(user.ID, post.ID)
	Expect(err).NotTo(HaveOccurred())
	startedAt := view.StartedAt
	endedAt := time.Now()

	// When: End view tracking
	err = viewService.EndView(user.ID, post.ID, startedAt, endedAt)

	// Then: Should update view record with duration
	Expect(err).NotTo(HaveOccurred())

	// Verify view record was created/updated
	var viewRecord domain.PostView
	err = containers.DB.QueryRow(`
		SELECT id, user_id, post_id, started_at, ended_at, duration_seconds 
		FROM post_views 
		WHERE user_id = $1 AND post_id = $2 AND started_at = $3`,
		user.ID, post.ID, startedAt).Scan(
		&viewRecord.ID, &viewRecord.UserID, &viewRecord.PostID, &viewRecord.StartedAt, &viewRecord.EndedAt, &viewRecord.DurationSeconds)

	Expect(err).NotTo(HaveOccurred())
	Expect(viewRecord.UserID).To(Equal(user.ID))
	Expect(viewRecord.PostID).To(Equal(post.ID))
	Expect(viewRecord.EndedAt).NotTo(BeNil())
	Expect(*viewRecord.DurationSeconds).To(BeNumerically(">=", 0)) // Should be >= 0 seconds
}

func TestPostViewService_MultipleViews(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	viewRepo := postgres.NewPostViewRepository(containers.DB)
	viewService := service.NewPostViewService(viewRepo)

	// Create test user and post
	user := createTestUser(t, containers.DB, "testuser", "test@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Test Post", "Test Caption")

	// When: User views the same post multiple times
	view1, err1 := viewService.StartView(user.ID, post.ID)
	time.Sleep(100 * time.Millisecond) // Small delay
	view2, err2 := viewService.StartView(user.ID, post.ID)

	// Then: Both views should be created successfully
	Expect(err1).NotTo(HaveOccurred())
	Expect(err2).NotTo(HaveOccurred())
	Expect(view1.ID).NotTo(Equal(view2.ID)) // Different view records

	// Verify views_count was incremented twice
	var viewsCount int
	err := containers.DB.QueryRow("SELECT views_count FROM posts WHERE id = $1", post.ID).Scan(&viewsCount)
	Expect(err).NotTo(HaveOccurred())
	Expect(viewsCount).To(Equal(2))
}

func TestPostViewRepository_IncrementViewsCount(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	viewRepo := postgres.NewPostViewRepository(containers.DB)

	// Create test user and post
	user := createTestUser(t, containers.DB, "testuser", "test@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Test Post", "Test Caption")

	// When: Increment views count multiple times
	err1 := viewRepo.IncrementPostViewsCount(post.ID)
	err2 := viewRepo.IncrementPostViewsCount(post.ID)
	err3 := viewRepo.IncrementPostViewsCount(post.ID)

	// Then: All increments should succeed
	Expect(err1).NotTo(HaveOccurred())
	Expect(err2).NotTo(HaveOccurred())
	Expect(err3).NotTo(HaveOccurred())

	// Verify views_count was incremented correctly
	var viewsCount int
	err := containers.DB.QueryRow("SELECT views_count FROM posts WHERE id = $1", post.ID).Scan(&viewsCount)
	Expect(err).NotTo(HaveOccurred())
	Expect(viewsCount).To(Equal(3))
}

func TestPostViewService_EndViewNonExistent(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	viewRepo := postgres.NewPostViewRepository(containers.DB)
	viewService := service.NewPostViewService(viewRepo)

	// Create test user and post
	user := createTestUser(t, containers.DB, "testuser", "test@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Test Post", "Test Caption")

	// When: Try to end a view that was never started
	startedAt := time.Now().Add(-5 * time.Second)
	endedAt := time.Now()
	err := viewService.EndView(user.ID, post.ID, startedAt, endedAt)

	// Then: Should handle gracefully (not fail)
	Expect(err).NotTo(HaveOccurred()) // Service should not fail for missing views
}
