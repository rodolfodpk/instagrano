package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rodolfodpk/instagrano/internal/dto"
	"github.com/rodolfodpk/instagrano/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body object{username=string,email=string,password=string} true "Registration details"
// @Success      200  {object}  object{user=domain.User}
// @Failure      400  {object}  object{error=string}
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	user, err := h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	response := dto.AuthResponse{
		User: dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Token: "", // No token on registration
	}
	return c.JSON(response)
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body object{email=string,password=string} true "Login credentials"
// @Success      200  {object}  object{user=domain.User,token=string}
// @Failure      401  {object}  object{error=string}
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	user, token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	response := dto.AuthResponse{
		User: dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Token: token,
	}
	return c.JSON(response)
}

// GetMe godoc
// @Summary      Get current user
// @Description  Get the current authenticated user's information
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{user=dto.UserResponse}
// @Failure      401  {object}  object{error=string}
// @Router       /auth/me [get]
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	response := dto.AuthResponse{
		User: dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Token: "", // Don't return token for security
	}
	return c.JSON(response)
}
