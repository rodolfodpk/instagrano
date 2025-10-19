package handlers

import (
	"io"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// @title Instagrano API
// @version 1.0
// @description A simple Instagram-like API built with Go and Fiber
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /

// @Summary Get home message
// @Description Returns a welcome message for Instagrano
// @Tags general
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router / [get]

var users, posts, comments = make(map[string]User), []Post{}, make(map[int][]Comment)

func Home(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "🚀 Instagrano!"})
}

// @Summary Register a new user
// @Description Register a new user with username, email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object true "Registration data"
// @Success 200 {object} map[string]interface{}
// @Router /register [post]
func Register(c *fiber.Ctx) error {
	var r struct{ Username, Email, Password string }
	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if r.Username == "" || r.Email == "" || r.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Username, email, and password are required"})
	}
	if _, exists := users[r.Email]; exists {
		return c.Status(409).JSON(fiber.Map{"error": "User already exists"})
	}
	users[r.Email] = User{ID: uint(len(users) + 1), Username: r.Username, Email: r.Email}
	return c.JSON(fiber.Map{"user": users[r.Email], "token": "instagrano-" + r.Username})
}

// @Summary Login user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object true "Login data"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /login [post]
func Login(c *fiber.Ctx) error {
	var r struct{ Email, Password string }
	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if r.Email == "" || r.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email and password are required"})
	}
	u, ok := users[r.Email]
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}
	return c.JSON(fiber.Map{"user": u, "token": "instagrano-" + u.Username})
}

// @Summary Upload image
// @Description Upload an image file
// @Tags posts
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file"
// @Param caption formData string false "Image caption"
// @Success 200 {object} map[string]interface{}
// @Router /upload [post]
func Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer f.Close()

	_, err = io.ReadAll(f)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read file"})
	}

	url := "https://s3.instagrano.com/" + file.Filename
	post := Post{ID: len(posts) + 1, UserID: 1, ImageURL: url, Caption: c.FormValue("caption")}
	posts = append(posts, post)
	return c.JSON(fiber.Map{"post_id": post.ID, "image_url": url})
}

// @Summary Get feed
// @Description Get all posts in the feed
// @Tags posts
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /feed [get]
func Feed(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"posts": posts})
}

// @Summary Like a post
// @Description Like a post by ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]int
// @Failure 404 {object} map[string]string
// @Router /posts/{id}/like [post]
func Like(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	for i := range posts {
		if posts[i].ID == id {
			posts[i].LikesCount++
			return c.JSON(fiber.Map{"likes": posts[i].LikesCount})
		}
	}
	return c.Status(404).JSON(fiber.Map{"error": "Not found"})
}

// @Summary Add comment to post
// @Description Add a comment to a post by ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param request body object true "Comment data"
// @Success 200 {object} map[string]interface{}
// @Router /posts/{id}/comments [post]
func AddComment(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	var r struct{ Text string }
	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if r.Text == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Comment text is required"})
	}

	comments[id] = append(comments[id], Comment{ID: len(comments[id]) + 1, Text: r.Text})
	return c.JSON(fiber.Map{"comments": comments[id]})
}

type User struct {
	ID              uint
	Username, Email string
}

type Post struct {
	ID                int
	UserID            uint
	ImageURL, Caption string
	LikesCount        int
}

type Comment struct {
	ID   int
	Text string
}
