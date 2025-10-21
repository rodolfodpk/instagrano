package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	postgresRepo "github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
)

var _ = Describe("AuthService", func() {
	Describe("Register", func() {
		It("should register a new user successfully", func() {
			// Given: Auth service with test database
			userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
			authService := service.NewAuthService(userRepo, "test-secret")

			// When: Register a new user
			user, err := authService.Register("testuser", "test@example.com", "password123")

			// Then: Should register successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Username).To(Equal("testuser"))
			Expect(user.Email).To(Equal("test@example.com"))
			Expect(user.ID).To(BeNumerically(">", 0))
		})

		It("should reject duplicate username", func() {
			// Given: Auth service with existing user
			userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
			authService := service.NewAuthService(userRepo, "test-secret")

			// First registration
			_, err := authService.Register("testuser", "test@example.com", "password123")
			Expect(err).NotTo(HaveOccurred())

			// When: Try to register with same username
			_, err = authService.Register("testuser", "test2@example.com", "password123")

			// Then: Should reject duplicate username
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("username"))
		})

		It("should reject duplicate email", func() {
			// Given: Auth service with existing user
			userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
			authService := service.NewAuthService(userRepo, "test-secret")

			// First registration
			_, err := authService.Register("testuser", "test@example.com", "password123")
			Expect(err).NotTo(HaveOccurred())

			// When: Try to register with same email
			_, err = authService.Register("testuser2", "test@example.com", "password123")

			// Then: Should reject duplicate email
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("email"))
		})

		It("should reject invalid email format", func() {
			// Given: Auth service
			userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
			authService := service.NewAuthService(userRepo, "test-secret")

			// When: Register with invalid email (empty email)
			_, err := authService.Register("testuser", "", "password123")

			// Then: Should reject empty email
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid input"))
		})

		It("should reject short password", func() {
			// Given: Auth service
			userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
			authService := service.NewAuthService(userRepo, "test-secret")

			// When: Register with empty password
			_, err := authService.Register("testuser", "test@example.com", "")

			// Then: Should reject empty password
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid input"))
		})

		It("should reject empty username", func() {
			// Given: Auth service
			userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
			authService := service.NewAuthService(userRepo, "test-secret")

			// When: Register with empty username
			_, err := authService.Register("", "test@example.com", "password123")

			// Then: Should reject empty username
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid input"))
		})

		It("should hash password correctly", func() {
			// Given: Auth service
			userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
			authService := service.NewAuthService(userRepo, "test-secret")

			// When: Register a user
			user, err := authService.Register("testuser", "test@example.com", "password123")

			// Then: Password should be hashed
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Password).NotTo(Equal("password123"))
			Expect(len(user.Password)).To(BeNumerically(">", 50)) // bcrypt hash length
		})
	})

	Describe("Login", func() {
		It("should login with correct credentials", func() {
			// Given: Registered user
			userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
			authService := service.NewAuthService(userRepo, "test-secret")

			_, err := authService.Register("testuser", "test@example.com", "password123")
			Expect(err).NotTo(HaveOccurred())

			// When: Login with correct credentials
			user, token, err := authService.Login("test@example.com", "password123")

			// Then: Should login successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Username).To(Equal("testuser"))
			Expect(token).NotTo(BeEmpty())
		})

		It("should reject incorrect password", func() {
			// Given: Registered user
			userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
			authService := service.NewAuthService(userRepo, "test-secret")

			_, err := authService.Register("testuser", "test@example.com", "password123")
			Expect(err).NotTo(HaveOccurred())

			// When: Login with incorrect password
			_, _, err = authService.Login("test@example.com", "wrongpassword")

			// Then: Should reject login
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid credentials"))
		})

		It("should reject non-existent email", func() {
			// Given: Auth service
			userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
			authService := service.NewAuthService(userRepo, "test-secret")

			// When: Login with non-existent email
			_, _, err := authService.Login("nonexistent@example.com", "password123")

			// Then: Should reject login
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid credentials"))
		})

		It("should generate valid JWT token", func() {
			// Given: Registered user
			userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
			authService := service.NewAuthService(userRepo, "test-secret")

			_, err := authService.Register("testuser", "test@example.com", "password123")
			Expect(err).NotTo(HaveOccurred())

			// When: Login
			_, token, err := authService.Login("test@example.com", "password123")

			// Then: Should generate valid token
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeEmpty())
			Expect(len(token)).To(BeNumerically(">", 50)) // JWT tokens are longer
		})
	})
})
