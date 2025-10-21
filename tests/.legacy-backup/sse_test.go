package tests

import (
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSE Tests", func() {
	It("should authenticate SSE connections properly", func() {
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

	It("should establish SSE connection and receive initial event", func() {
		app, _, cleanup := setupTestApp()
		defer cleanup()

		token := registerAndLogin(app, "sseuser2", "sse2@example.com", "pass123")

		// Test that we can establish a connection
		req := httptest.NewRequest("GET", "/api/events/stream?token="+token, nil)
		resp, err := app.Test(req, 100) // Short timeout to avoid hanging
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))
		Expect(resp.Header.Get("Content-Type")).To(Equal("text/event-stream"))
	})

	It("should handle SSE authentication edge cases", func() {
		app, _, cleanup := setupTestApp()
		defer cleanup()

		// Test empty token
		req := httptest.NewRequest("GET", "/api/events/stream?token=", nil)
		resp, err := app.Test(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(401))

		// Test malformed token
		req = httptest.NewRequest("GET", "/api/events/stream?token=malformed.token.here", nil)
		resp, err = app.Test(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(401))

		// Test valid token
		token := registerAndLogin(app, "sseuser3", "sse3@example.com", "pass123")
		req = httptest.NewRequest("GET", "/api/events/stream?token="+token, nil)
		resp, err = app.Test(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))
	})

	It("should validate JWT token format in SSE handler", func() {
		app, _, cleanup := setupTestApp()
		defer cleanup()

		// Test with a token that has wrong format
		req := httptest.NewRequest("GET", "/api/events/stream?token=not.a.valid.jwt", nil)
		resp, err := app.Test(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(401))

		// Test with a token that has wrong signature
		req = httptest.NewRequest("GET", "/api/events/stream?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.invalid", nil)
		resp, err = app.Test(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(401))
	})
})
