package tests

import (
	"encoding/json"
	"net/http/httptest"
	"time"

	"github.com/gofiber/fiber/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/events"
)

var _ = Describe("Server-Sent Events", func() {
	var (
		app   *fiber.App
		token string
		user  *domain.User
	)

	BeforeEach(func() {
		app, _, _ = setupTestApp()

		// Create test user and get token
		user = createTestUser(sharedContainers.DB, "sseuser", "sse@example.com")
		var err error
		token, err = createTestJWT(user.ID)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Authentication", func() {
		It("should reject requests without token", func() {
			// When: Connect to SSE without token
			req := httptest.NewRequest("GET", "/api/events/stream", nil)
			resp, err := app.Test(req)

			// Then: Should return 401 Unauthorized
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(401))
		})

		It("should reject requests with invalid token", func() {
			// When: Connect to SSE with invalid token
			req := httptest.NewRequest("GET", "/api/events/stream?token=invalid", nil)
			resp, err := app.Test(req)

			// Then: Should return 401 Unauthorized
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(401))
		})

		It("should accept requests with valid token", func() {
			// When: Connect to SSE with valid token
			eventCh, cleanup := connectSSE(app, token)
			defer cleanup()

			// Then: Should receive connection event
			event := waitForSSEEvent(eventCh, string(string(events.EventTypeConnected)), 5*time.Second)
			Expect(event.Type).To(Equal(string(string(events.EventTypeConnected))))
		})
	})

	Describe("Event Publishing", func() {
		It("should publish new post events", func() {
			// Given: SSE connection
			eventCh, cleanup := connectSSE(app, token)
			defer cleanup()

			// Wait for connection event
			waitForSSEEvent(eventCh, string(events.EventTypeConnected), 5*time.Second)

			// When: Create a new post
			_ = createTestPostWithSSE(app, token, "SSE Test Post", "This is a test post for SSE")

			// Then: Should receive new post event
			event := waitForSSEEvent(eventCh, string(events.EventTypeNewPost), 5*time.Second)
			Expect(event.Type).To(Equal(string(events.EventTypeNewPost)))

			// Parse event data
			eventData := parseSSEEventData(event.Data)
			Expect(eventData.PostID).To(BeNumerically(">", 0))
			Expect(eventData.TriggeredByUserID).To(Equal(user.ID))
		})

		It("should publish post liked events", func() {
			// Given: SSE connection and a post
			eventCh, cleanup := connectSSE(app, token)
			defer cleanup()

			// Wait for connection event
			waitForSSEEvent(eventCh, string(events.EventTypeConnected), 5*time.Second)

			// Create a post first
			postData := createTestPostWithSSE(app, token, "Like Test Post", "This post will be liked")
			postID := uint(postData["id"].(float64))

			// When: Like the post
			likeTestPostWithSSE(app, token, postID)

			// Then: Should receive post liked event
			event := waitForSSEEvent(eventCh, string(events.EventTypePostLiked), 5*time.Second)
			Expect(event.Type).To(Equal(string(events.EventTypePostLiked)))

			// Parse event data
			eventData := parseSSEEventData(event.Data)
			Expect(eventData.PostID).To(Equal(postID))
			Expect(eventData.TriggeredByUserID).To(Equal(user.ID))
		})

		It("should publish post commented events", func() {
			// Given: SSE connection and a post
			eventCh, cleanup := connectSSE(app, token)
			defer cleanup()

			// Wait for connection event
			waitForSSEEvent(eventCh, string(events.EventTypeConnected), 5*time.Second)

			// Create a post first
			postData := createTestPostWithSSE(app, token, "Comment Test Post", "This post will be commented on")
			postID := uint(postData["id"].(float64))

			// When: Comment on the post
			commentTestPostWithSSE(app, token, postID, "This is a test comment")

			// Then: Should receive post commented event
			event := waitForSSEEvent(eventCh, string(events.EventTypePostCommented), 5*time.Second)
			Expect(event.Type).To(Equal(string(events.EventTypePostCommented)))

			// Parse event data
			eventData := parseSSEEventData(event.Data)
			Expect(eventData.PostID).To(Equal(postID))
			Expect(eventData.TriggeredByUserID).To(Equal(user.ID))
		})
	})

	Describe("Event Filtering", func() {
		It("should not send events triggered by the same user", func() {
			// Given: SSE connection
			eventCh, cleanup := connectSSE(app, token)
			defer cleanup()

			// Wait for connection event
			waitForSSEEvent(eventCh, string(events.EventTypeConnected), 5*time.Second)

			// When: Create a post (this should trigger an event)
			_ = createTestPostWithSSE(app, token, "Self Filter Test", "This should not trigger an event for the same user")

			// Then: Should NOT receive new post event (filtered out)
			select {
			case event := <-eventCh:
				// If we receive an event, it should not be a new_post event from the same user
				if event.Type == string(events.EventTypeNewPost) {
					eventData := parseSSEEventData(event.Data)
					Expect(eventData.TriggeredByUserID).NotTo(Equal(user.ID))
				}
			case <-time.After(2 * time.Second):
				// Timeout is expected - no event should be received
			}
		})
	})

	Describe("Multi-User Scenario", func() {
		It("should handle multiple users receiving events", func() {
			// Given: Two users
			user2 := createTestUser(sharedContainers.DB, "sseuser2", "sse2@example.com")
			token2, err := createTestJWT(user2.ID)
			Expect(err).NotTo(HaveOccurred())

			// Create SSE connections for both users
			eventCh1, cleanup1 := connectSSE(app, token)
			defer cleanup1()
			eventCh2, cleanup2 := connectSSE(app, token2)
			defer cleanup2()

			// Wait for both connections
			waitForSSEEvent(eventCh1, string(events.EventTypeConnected), 5*time.Second)
			waitForSSEEvent(eventCh2, string(events.EventTypeConnected), 5*time.Second)

			// When: User 1 creates a post
			_ = createTestPostWithSSE(app, token, "Multi User Test", "This post should be seen by user 2")

			// Then: User 2 should receive the new post event
			event := waitForSSEEvent(eventCh2, string(events.EventTypeNewPost), 5*time.Second)
			Expect(event.Type).To(Equal(string(events.EventTypeNewPost)))

			eventData := parseSSEEventData(event.Data)
			Expect(eventData.TriggeredByUserID).To(Equal(user.ID)) // Triggered by user 1

			// And: User 1 should NOT receive the event (filtered out)
			select {
			case event := <-eventCh1:
				if event.Type == string(events.EventTypeNewPost) {
					Fail("User 1 should not receive their own new post event")
				}
			case <-time.After(1 * time.Second):
				// Expected - no event for user 1
			}
		})
	})

	Describe("Event Data Structure", func() {
		It("should have correct event structure", func() {
			// Given: SSE connection
			eventCh, cleanup := connectSSE(app, token)
			defer cleanup()

			// Wait for connection event
			event := waitForSSEEvent(eventCh, string(events.EventTypeConnected), 5*time.Second)

			// Then: Event should have correct structure
			Expect(event.Type).To(Equal(string(events.EventTypeConnected)))
			Expect(event.Data).NotTo(BeEmpty())

			// Parse the event data
			var eventData map[string]interface{}
			err := json.Unmarshal(event.Data, &eventData)
			Expect(err).NotTo(HaveOccurred())
			Expect(eventData).To(HaveKey("message"))
		})

		It("should include heartbeat events", func() {
			// Given: SSE connection
			eventCh, cleanup := connectSSE(app, token)
			defer cleanup()

			// Wait for connection event
			waitForSSEEvent(eventCh, string(events.EventTypeConnected), 5*time.Second)

			// When: Wait for heartbeat (sent every 15 seconds)
			event := waitForSSEEvent(eventCh, string(events.EventTypeHeartbeat), 20*time.Second)

			// Then: Should receive heartbeat event
			Expect(event.Type).To(Equal(string(events.EventTypeHeartbeat)))

			var eventData map[string]interface{}
			err := json.Unmarshal(event.Data, &eventData)
			Expect(err).NotTo(HaveOccurred())
			Expect(eventData).To(HaveKey("timestamp"))
		})
	})
})
