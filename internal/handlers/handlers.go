package handlers

import (
	"github.com/gofiber/fiber/v2"
	"io"
	"strconv"
)

var users, posts, comments = make(map[string]User), []Post{}, make(map[int][]Comment)

func Home(c *fiber.Ctx) { 
	c.JSON(fiber.Map{"message": "🚀 Instagrano!"}) 
}

func Register(c *fiber.Ctx) error {
	var r struct{ Username, Email, Password string }
	c.BodyParser(&r)
	users[r.Email] = User{ID: uint(len(users) + 1), Username: r.Username, Email: r.Email}
	return c.JSON(fiber.Map{"user": users[r.Email], "token": "instagrano-" + r.Username})
}

func Login(c *fiber.Ctx) error {
	var r struct{ Email, Password string }
	c.BodyParser(&r)
	u, ok := users[r.Email]
	if !ok { 
		return c.Status(401).JSON(fiber.Map{"error": "Invalid"}) 
	}
	return c.JSON(fiber.Map{"user": u, "token": "instagrano-" + u.Username})
}

func Upload(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	file, _ := c.FormFile("file")
	f, _ := file.Open()
	data, _ := io.ReadAll(f)
	url := "https://s3.instagrano.com/" + file.Filename
	post := Post{ID: len(posts) + 1, UserID: 1, ImageURL: url, Caption: c.FormValue("caption")}
	posts = append(posts, post)
	return c.JSON(fiber.Map{"post_id": post.ID, "image_url": url})
}

func Feed(c *fiber.Ctx) error { 
	return c.JSON(fiber.Map{"posts": posts}) 
}

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

func Comment(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var r struct{ Text string }
	c.BodyParser(&r)
	comments[id] = append(comments[id], Comment{ID: len(comments[id]) + 1, Text: r.Text})
	return c.JSON(fiber.Map{"comments": comments[id]})
}

type User struct { 
	ID uint
	Username, Email string 
}

type Post struct { 
	ID int
	UserID uint
	ImageURL, Caption string
	LikesCount int 
}

type Comment struct { 
	ID int
	Text string 
}
