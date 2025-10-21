package tests

import (
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/events"
)

var _ = Describe("Server-Sent Events", func() {
	Describe("Authentication", func() {
		It("should handle authentication correctly", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Test missing token
			req := httptest.NewRequest("GET", "/api/events/stream", nil)
			resp, err := app.Test(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(401))

			// Test invalid token
			req = httptest.NewRequest("GET", "/api/events/stream?token=invalid", nil)
			resp, err = app.Test(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(401))

			// Test valid token
			token := registerAndLogin(app, "sseuser1", "sse1@example.com", "pass123")
			req = httptest.NewRequest("GET", "/api/events/stream?token="+token, nil)
			resp, err = app.Test(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})

	Describe("Connection", func() {
		It("should establish SSE connection and receive events", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			token := registerAndLogin(app, "sseuser2", "sse2@example.com", "pass123")

			// When: Connect to SSE
			eventCh, cleanupSSE := connectSSE(app, token)
			defer cleanupSSE()

			// Then: Should receive connected event
			connectedEvent := waitForSSEEvent(eventCh, "connected", 5*time.Second)
			Expect(connectedEvent.Type).To(Equal("connected"))

			// And: Should receive heartbeat events
			heartbeatEvent := waitForSSEEvent(eventCh, "heartbeat", 35*time.Second)
			Expect(heartbeatEvent.Type).To(Equal("heartbeat"))
		})
	})

	Describe("New Post Events", func() {
		It("should broadcast new post events to other users", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: Two users registered
			token1 := registerAndLogin(app, "sseuser3", "sse3@example.com", "pass123")
			token2 := registerAndLogin(app, "sseuser4", "sse4@example.com", "pass123")

			// Given: Both users connected to SSE
			eventCh1, cleanup1 := connectSSE(app, token1)
			defer cleanup1()
			eventCh2, cleanup2 := connectSSE(app, token2)
			defer cleanup2()

			// Wait for both connections to be established
			waitForSSEEvent(eventCh1, "connected", 5*time.Second)
			waitForSSEEvent(eventCh2, "connected", 5*time.Second)

			// When: User 1 creates a post
			postData := createTestPostWithSSE(app, token1, "SSE Test Post", "Testing real-time updates!")

			// Then: User 2 should receive the new_post event
			newPostEvent := waitForSSEEvent(eventCh2, "new_post", 5*time.Second)
			event := parseSSEEventData(newPostEvent.Data)

			Expect(event.Type).To(Equal(string(events.EventTypeNewPost)))
			Expect(event.PostID).To(Equal(uint(postData["id"].(float64))))
			Expect(event.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64))))

			// And: User 1 should NOT receive their own event (self-filtering)
			Consistently(eventCh1, 2*time.Second).ShouldNot(Receive())
		})
	})

	Describe("Like Events", func() {
		It("should broadcast like events to other users", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: Two users registered
			token1 := registerAndLogin(app, "sseuser5", "sse5@example.com", "pass123")
			token2 := registerAndLogin(app, "sseuser6", "sse6@example.com", "pass123")

			// Given: Both users connected to SSE
			eventCh1, cleanup1 := connectSSE(app, token1)
			defer cleanup1()
			eventCh2, cleanup2 := connectSSE(app, token2)
			defer cleanup2()

			// Wait for both connections to be established
			waitForSSEEvent(eventCh1, "connected", 5*time.Second)
			waitForSSEEvent(eventCh2, "connected", 5*time.Second)

			// Given: User 1 creates a post
			postData := createTestPostWithSSE(app, token1, "Like Test Post", "Testing like events!")
			postID := uint(postData["id"].(float64))

			// User 2 should receive the new_post event
			waitForSSEEvent(eventCh2, "new_post", 5*time.Second)

			// When: User 2 likes the post
			likeTestPostWithSSE(app, token2, postID)

			// Then: User 1 should receive the post_liked event
			likedEvent := waitForSSEEvent(eventCh1, "post_liked", 5*time.Second)
			event := parseSSEEventData(likedEvent.Data)

			Expect(event.Type).To(Equal(string(events.EventTypePostLiked)))
			Expect(event.PostID).To(Equal(postID))
			Expect(event.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64)))) // User 2's ID

			// And: User 2 should NOT receive their own like event
			Consistently(eventCh2, 2*time.Second).ShouldNot(Receive())
		})
	})

	Describe("Comment Events", func() {
		It("should broadcast comment events to other users", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: Two users registered
			token1 := registerAndLogin(app, "sseuser7", "sse7@example.com", "pass123")
			token2 := registerAndLogin(app, "sseuser8", "sse8@example.com", "pass123")

			// Given: Both users connected to SSE
			eventCh1, cleanup1 := connectSSE(app, token1)
			defer cleanup1()
			eventCh2, cleanup2 := connectSSE(app, token2)
			defer cleanup2()

			// Wait for both connections to be established
			waitForSSEEvent(eventCh1, "connected", 5*time.Second)
			waitForSSEEvent(eventCh2, "connected", 5*time.Second)

			// Given: User 1 creates a post
			postData := createTestPostWithSSE(app, token1, "Comment Test Post", "Testing comment events!")
			postID := uint(postData["id"].(float64))

			// User 2 should receive the new_post event
			waitForSSEEvent(eventCh2, "new_post", 5*time.Second)

			// When: User 2 comments on the post
			commentTestPostWithSSE(app, token2, postID, "Great post!")

			// Then: User 1 should receive the post_commented event
			commentedEvent := waitForSSEEvent(eventCh1, "post_commented", 5*time.Second)
			event := parseSSEEventData(commentedEvent.Data)

			Expect(event.Type).To(Equal(string(events.EventTypePostCommented)))
			Expect(event.PostID).To(Equal(postID))
			Expect(event.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64)))) // User 2's ID

			// And: User 2 should NOT receive their own comment event
			Consistently(eventCh2, 2*time.Second).ShouldNot(Receive())
		})
	})

	Describe("Multi-User Scenarios", func() {
		It("should handle complex multi-user interactions", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: Three users registered
			token1 := registerAndLogin(app, "sseuser9", "sse9@example.com", "pass123")
			token2 := registerAndLogin(app, "sseuser10", "sse10@example.com", "pass123")
			token3 := registerAndLogin(app, "sseuser11", "sse11@example.com", "pass123")

			// Given: All users connected to SSE
			eventCh1, cleanup1 := connectSSE(app, token1)
			defer cleanup1()
			eventCh2, cleanup2 := connectSSE(app, token2)
			defer cleanup2()
			eventCh3, cleanup3 := connectSSE(app, token3)
			defer cleanup3()

			// Wait for all connections to be established
			waitForSSEEvent(eventCh1, "connected", 5*time.Second)
			waitForSSEEvent(eventCh2, "connected", 5*time.Second)
			waitForSSEEvent(eventCh3, "connected", 5*time.Second)

			// When: User 1 creates a post
			postData := createTestPostWithSSE(app, token1, "Multi-User Test Post", "Testing with multiple users!")
			postID := uint(postData["id"].(float64))

			// Then: Users 2 and 3 should receive the new_post event
			waitForSSEEvent(eventCh2, "new_post", 5*time.Second)
			waitForSSEEvent(eventCh3, "new_post", 5*time.Second)

			// And: User 1 should NOT receive their own event
			Consistently(eventCh1, 2*time.Second).ShouldNot(Receive())

			// When: User 2 likes the post
			likeTestPostWithSSE(app, token2, postID)

			// Then: Users 1 and 3 should receive the post_liked event
			waitForSSEEvent(eventCh1, "post_liked", 5*time.Second)
			waitForSSEEvent(eventCh3, "post_liked", 5*time.Second)

			// And: User 2 should NOT receive their own like event
			Consistently(eventCh2, 2*time.Second).ShouldNot(Receive())

			// When: User 3 comments on the post
			commentTestPostWithSSE(app, token3, postID, "Awesome post!")

			// Then: Users 1 and 2 should receive the post_commented event
			waitForSSEEvent(eventCh1, "post_commented", 5*time.Second)
			waitForSSEEvent(eventCh2, "post_commented", 5*time.Second)

			// And: User 3 should NOT receive their own comment event
			Consistently(eventCh3, 2*time.Second).ShouldNot(Receive())
		})
	})

	Describe("Event Data Structure", func() {
		It("should maintain correct event data structure", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			token1 := registerAndLogin(app, "sseuser12", "sse12@example.com", "pass123")
			token2 := registerAndLogin(app, "sseuser13", "sse13@example.com", "pass123")

			eventCh1, cleanup1 := connectSSE(app, token1)
			defer cleanup1()
			eventCh2, cleanup2 := connectSSE(app, token2)
			defer cleanup2()

			waitForSSEEvent(eventCh1, "connected", 5*time.Second)
			waitForSSEEvent(eventCh2, "connected", 5*time.Second)

			// Test new_post event structure
			postData := createTestPostWithSSE(app, token1, "Data Structure Test", "Testing event data structure!")
			postID := uint(postData["id"].(float64))

			newPostEvent := waitForSSEEvent(eventCh2, "new_post", 5*time.Second)
			event := parseSSEEventData(newPostEvent.Data)

			Expect(event.Type).To(Equal(string(events.EventTypeNewPost)))
			Expect(event.PostID).To(Equal(postID))
			Expect(event.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64))))
			Expect(event.Timestamp).To(BeNumerically(">", 0))

			// Test like event structure
			likeTestPostWithSSE(app, token2, postID)
			likedEvent := waitForSSEEvent(eventCh1, "post_liked", 5*time.Second)
			likeEvent := parseSSEEventData(likedEvent.Data)

			Expect(likeEvent.Type).To(Equal(string(events.EventTypePostLiked)))
			Expect(likeEvent.PostID).To(Equal(postID))
			Expect(likeEvent.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64)))) // User 2's ID

			// Test comment event structure
			commentTestPostWithSSE(app, token2, postID, "Test comment")
			commentedEvent := waitForSSEEvent(eventCh1, "post_commented", 5*time.Second)
			commentEvent := parseSSEEventData(commentedEvent.Data)

			Expect(commentEvent.Type).To(Equal(string(events.EventTypePostCommented)))
			Expect(commentEvent.PostID).To(Equal(postID))
			Expect(commentEvent.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64)))) // User 2's ID
		})
	})
})