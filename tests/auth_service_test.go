package tests

import (
	"testing"

	. "github.com/onsi/gomega"

	postgresRepo "github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
)

func TestAuthService_Register(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	userRepo := postgresRepo.NewUserRepository(containers.DB)
	authService := service.NewAuthService(userRepo, "test-secret")

	// When: Register new user
	user, err := authService.Register("newuser", "new@example.com", "password123")

	// Then: Registration succeeds
	Expect(err).NotTo(HaveOccurred())
	Expect(user.Username).To(Equal("newuser"))
	Expect(user.Email).To(Equal("new@example.com"))
	Expect(user.Password).NotTo(Equal("password123")) // Should be hashed
	Expect(user.Password).To(HavePrefix("$2a$")) // bcrypt hash
}

func TestAuthService_Login(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	userRepo := postgresRepo.NewUserRepository(containers.DB)
	authService := service.NewAuthService(userRepo, "test-secret")

	// Given: User exists
	_, err := authService.Register("loginuser", "login@example.com", "password123")
	Expect(err).NotTo(HaveOccurred())

	// When: Login with correct credentials
	user, token, err := authService.Login("login@example.com", "password123")

	// Then: Login succeeds
	Expect(err).NotTo(HaveOccurred())
	Expect(token).NotTo(BeEmpty())
	Expect(user.Email).To(Equal("login@example.com"))

	// Verify JWT format (header.payload.signature)
	Expect(token).To(MatchRegexp(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$`))
}

func TestAuthService_LoginWithWrongPassword(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	userRepo := postgresRepo.NewUserRepository(containers.DB)
	authService := service.NewAuthService(userRepo, "test-secret")

	// Given: User exists
	authService.Register("wrongpass", "wrong@example.com", "correctpass")

	// When: Login with wrong password
	_, _, err := authService.Login("wrong@example.com", "wrongpass")

	// Then: Login fails
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("invalid credentials"))
}

func TestAuthService_LoginWithNonExistentUser(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	userRepo := postgresRepo.NewUserRepository(containers.DB)
	authService := service.NewAuthService(userRepo, "test-secret")

	// When: Login with non-existent user
	_, _, err := authService.Login("nonexistent@example.com", "password123")

	// Then: Login fails
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("invalid credentials"))
}

func TestAuthService_RegisterDuplicateEmail(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	userRepo := postgresRepo.NewUserRepository(containers.DB)
	authService := service.NewAuthService(userRepo, "test-secret")

	// Given: User already exists
	_, err := authService.Register("user1", "duplicate@example.com", "password123")
	Expect(err).NotTo(HaveOccurred())

	// When: Try to register with same email
	_, err = authService.Register("user2", "duplicate@example.com", "password456")

	// Then: Registration fails
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("duplicate"))
}

func TestAuthService_RegisterDuplicateUsername(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	userRepo := postgresRepo.NewUserRepository(containers.DB)
	authService := service.NewAuthService(userRepo, "test-secret")

	// Given: User already exists
	_, err := authService.Register("duplicateuser", "user1@example.com", "password123")
	Expect(err).NotTo(HaveOccurred())

	// When: Try to register with same username
	_, err = authService.Register("duplicateuser", "user2@example.com", "password456")

	// Then: Registration fails
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("duplicate"))
}

func TestAuthService_RegisterEmptyFields(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	userRepo := postgresRepo.NewUserRepository(containers.DB)
	authService := service.NewAuthService(userRepo, "test-secret")

	// When: Register with empty username
	_, err := authService.Register("", "test@example.com", "password123")
	Expect(err).To(HaveOccurred())

	// When: Register with empty email
	_, err = authService.Register("testuser", "", "password123")
	Expect(err).To(HaveOccurred())

	// When: Register with empty password
	_, err = authService.Register("testuser", "test@example.com", "")
	Expect(err).To(HaveOccurred())
}

func TestAuthService_JWTTokenValidation(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	userRepo := postgresRepo.NewUserRepository(containers.DB)
	authService := service.NewAuthService(userRepo, "test-secret")

	// Given: User exists and logs in
	_, err := authService.Register("jwtuser", "jwt@example.com", "password123")
	Expect(err).NotTo(HaveOccurred())

	user, token, err := authService.Login("jwt@example.com", "password123")
	Expect(err).NotTo(HaveOccurred())

	// Then: Token contains user information
	Expect(token).NotTo(BeEmpty())
	Expect(user.Email).To(Equal("jwt@example.com"))
	Expect(user.Username).To(Equal("jwtuser"))
}
