package tests

import (
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/rodolfodpk/instagrano/internal/events"
)

func TestSSEAuthentication(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
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
	token := registerAndLogin(t, app, "sseuser1", "sse1@example.com", "pass123")
	req = httptest.NewRequest("GET", "/api/events/stream?token="+token, nil)
	resp, err = app.Test(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
}

func TestSSEConnection(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	token := registerAndLogin(t, app, "sseuser2", "sse2@example.com", "pass123")

	// Connect to SSE
	eventCh, cleanupSSE := connectSSE(t, app, token)
	defer cleanupSSE()

	// Should receive connected event
	connectedEvent := waitForSSEEvent(t, eventCh, "connected", 5*time.Second)
	Expect(connectedEvent.Type).To(Equal("connected"))

	// Should receive heartbeat events
	heartbeatEvent := waitForSSEEvent(t, eventCh, "heartbeat", 35*time.Second)
	Expect(heartbeatEvent.Type).To(Equal("heartbeat"))
}

func TestSSENewPostEvent(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Register two users
	token1 := registerAndLogin(t, app, "sseuser3", "sse3@example.com", "pass123")
	token2 := registerAndLogin(t, app, "sseuser4", "sse4@example.com", "pass123")

	// Connect both users to SSE
	eventCh1, cleanup1 := connectSSE(t, app, token1)
	defer cleanup1()
	eventCh2, cleanup2 := connectSSE(t, app, token2)
	defer cleanup2()

	// Wait for both connections to be established
	waitForSSEEvent(t, eventCh1, "connected", 5*time.Second)
	waitForSSEEvent(t, eventCh2, "connected", 5*time.Second)

	// User 1 creates a post
	postData := createTestPostWithSSE(t, app, token1, "SSE Test Post", "Testing real-time updates!")

	// User 2 should receive the new_post event
	newPostEvent := waitForSSEEvent(t, eventCh2, "new_post", 5*time.Second)
	event := parseSSEEventData(t, newPostEvent.Data)

	Expect(event.Type).To(Equal(events.EventTypeNewPost))
	Expect(event.PostID).To(Equal(uint(postData["id"].(float64))))
	Expect(event.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64))))

	// User 1 should NOT receive their own event (self-filtering)
	select {
	case <-eventCh1:
		t.Fatal("User 1 should not receive their own post event")
	case <-time.After(2 * time.Second):
		// This is expected - no event should be received
	}
}

func TestSSELikeEvent(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Register two users
	token1 := registerAndLogin(t, app, "sseuser5", "sse5@example.com", "pass123")
	token2 := registerAndLogin(t, app, "sseuser6", "sse6@example.com", "pass123")

	// Connect both users to SSE
	eventCh1, cleanup1 := connectSSE(t, app, token1)
	defer cleanup1()
	eventCh2, cleanup2 := connectSSE(t, app, token2)
	defer cleanup2()

	// Wait for both connections to be established
	waitForSSEEvent(t, eventCh1, "connected", 5*time.Second)
	waitForSSEEvent(t, eventCh2, "connected", 5*time.Second)

	// User 1 creates a post
	postData := createTestPostWithSSE(t, app, token1, "Like Test Post", "Testing like events!")
	postID := uint(postData["id"].(float64))

	// User 2 should receive the new_post event
	waitForSSEEvent(t, eventCh2, "new_post", 5*time.Second)

	// User 2 likes the post
	likeTestPostWithSSE(t, app, token2, postID)

	// User 1 should receive the post_liked event
	likedEvent := waitForSSEEvent(t, eventCh1, "post_liked", 5*time.Second)
	event := parseSSEEventData(t, likedEvent.Data)

	Expect(event.Type).To(Equal(events.EventTypePostLiked))
	Expect(event.PostID).To(Equal(postID))
	Expect(event.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64)))) // User 2's ID

	// User 2 should NOT receive their own like event
	select {
	case <-eventCh2:
		t.Fatal("User 2 should not receive their own like event")
	case <-time.After(2 * time.Second):
		// This is expected - no event should be received
	}
}

func TestSSECommentEvent(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Register two users
	token1 := registerAndLogin(t, app, "sseuser7", "sse7@example.com", "pass123")
	token2 := registerAndLogin(t, app, "sseuser8", "sse8@example.com", "pass123")

	// Connect both users to SSE
	eventCh1, cleanup1 := connectSSE(t, app, token1)
	defer cleanup1()
	eventCh2, cleanup2 := connectSSE(t, app, token2)
	defer cleanup2()

	// Wait for both connections to be established
	waitForSSEEvent(t, eventCh1, "connected", 5*time.Second)
	waitForSSEEvent(t, eventCh2, "connected", 5*time.Second)

	// User 1 creates a post
	postData := createTestPostWithSSE(t, app, token1, "Comment Test Post", "Testing comment events!")
	postID := uint(postData["id"].(float64))

	// User 2 should receive the new_post event
	waitForSSEEvent(t, eventCh2, "new_post", 5*time.Second)

	// User 2 comments on the post
	commentTestPostWithSSE(t, app, token2, postID, "Great post!")

	// User 1 should receive the post_commented event
	commentedEvent := waitForSSEEvent(t, eventCh1, "post_commented", 5*time.Second)
	event := parseSSEEventData(t, commentedEvent.Data)

	Expect(event.Type).To(Equal(events.EventTypePostCommented))
	Expect(event.PostID).To(Equal(postID))
	Expect(event.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64)))) // User 2's ID

	// User 2 should NOT receive their own comment event
	select {
	case <-eventCh2:
		t.Fatal("User 2 should not receive their own comment event")
	case <-time.After(2 * time.Second):
		// This is expected - no event should be received
	}
}

func TestSSEMultiUserScenario(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Register three users
	token1 := registerAndLogin(t, app, "sseuser9", "sse9@example.com", "pass123")
	token2 := registerAndLogin(t, app, "sseuser10", "sse10@example.com", "pass123")
	token3 := registerAndLogin(t, app, "sseuser11", "sse11@example.com", "pass123")

	// Connect all users to SSE
	eventCh1, cleanup1 := connectSSE(t, app, token1)
	defer cleanup1()
	eventCh2, cleanup2 := connectSSE(t, app, token2)
	defer cleanup2()
	eventCh3, cleanup3 := connectSSE(t, app, token3)
	defer cleanup3()

	// Wait for all connections to be established
	waitForSSEEvent(t, eventCh1, "connected", 5*time.Second)
	waitForSSEEvent(t, eventCh2, "connected", 5*time.Second)
	waitForSSEEvent(t, eventCh3, "connected", 5*time.Second)

	// User 1 creates a post
	postData := createTestPostWithSSE(t, app, token1, "Multi-User Test Post", "Testing with multiple users!")
	postID := uint(postData["id"].(float64))

	// Users 2 and 3 should receive the new_post event
	waitForSSEEvent(t, eventCh2, "new_post", 5*time.Second)
	waitForSSEEvent(t, eventCh3, "new_post", 5*time.Second)

	// User 1 should NOT receive their own event
	select {
	case <-eventCh1:
		t.Fatal("User 1 should not receive their own post event")
	case <-time.After(2 * time.Second):
		// Expected
	}

	// User 2 likes the post
	likeTestPostWithSSE(t, app, token2, postID)

	// Users 1 and 3 should receive the post_liked event
	waitForSSEEvent(t, eventCh1, "post_liked", 5*time.Second)
	waitForSSEEvent(t, eventCh3, "post_liked", 5*time.Second)

	// User 2 should NOT receive their own like event
	select {
	case <-eventCh2:
		t.Fatal("User 2 should not receive their own like event")
	case <-time.After(2 * time.Second):
		// Expected
	}

	// User 3 comments on the post
	commentTestPostWithSSE(t, app, token3, postID, "Awesome post!")

	// Users 1 and 2 should receive the post_commented event
	waitForSSEEvent(t, eventCh1, "post_commented", 5*time.Second)
	waitForSSEEvent(t, eventCh2, "post_commented", 5*time.Second)

	// User 3 should NOT receive their own comment event
	select {
	case <-eventCh3:
		t.Fatal("User 3 should not receive their own comment event")
	case <-time.After(2 * time.Second):
		// Expected
	}
}

func TestSSEEventDataStructure(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	token1 := registerAndLogin(t, app, "sseuser12", "sse12@example.com", "pass123")
	token2 := registerAndLogin(t, app, "sseuser13", "sse13@example.com", "pass123")

	eventCh1, cleanup1 := connectSSE(t, app, token1)
	defer cleanup1()
	eventCh2, cleanup2 := connectSSE(t, app, token2)
	defer cleanup2()

	waitForSSEEvent(t, eventCh1, "connected", 5*time.Second)
	waitForSSEEvent(t, eventCh2, "connected", 5*time.Second)

	// Test new_post event structure
	postData := createTestPostWithSSE(t, app, token1, "Data Structure Test", "Testing event data structure!")
	postID := uint(postData["id"].(float64))

	newPostEvent := waitForSSEEvent(t, eventCh2, "new_post", 5*time.Second)
	event := parseSSEEventData(t, newPostEvent.Data)

	Expect(event.Type).To(Equal(events.EventTypeNewPost))
	Expect(event.PostID).To(Equal(postID))
	Expect(event.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64))))
	Expect(event.Timestamp).To(BeNumerically(">", 0))

	// Test like event structure
	likeTestPostWithSSE(t, app, token2, postID)
	likedEvent := waitForSSEEvent(t, eventCh1, "post_liked", 5*time.Second)
	likeEvent := parseSSEEventData(t, likedEvent.Data)

	Expect(likeEvent.Type).To(Equal(events.EventTypePostLiked))
	Expect(likeEvent.PostID).To(Equal(postID))
	Expect(likeEvent.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64)))) // User 2's ID

	// Test comment event structure
	commentTestPostWithSSE(t, app, token2, postID, "Test comment")
	commentedEvent := waitForSSEEvent(t, eventCh1, "post_commented", 5*time.Second)
	commentEvent := parseSSEEventData(t, commentedEvent.Data)

	Expect(commentEvent.Type).To(Equal(events.EventTypePostCommented))
	Expect(commentEvent.PostID).To(Equal(postID))
	Expect(commentEvent.TriggeredByUserID).To(Equal(uint(postData["user_id"].(float64)))) // User 2's ID
}
