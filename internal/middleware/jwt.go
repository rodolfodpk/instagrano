package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWT(jwtSecret string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(401).JSON(fiber.Map{"error": "missing token"})
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Debug logging
        fmt.Printf("JWT Secret: %s\n", jwtSecret)
        fmt.Printf("Token: %s\n", tokenString)

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("invalid signing method")
            }
            return []byte(jwtSecret), nil
        })

        if err != nil {
            fmt.Printf("JWT Parse Error: %v\n", err)
            return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
        }
        
        if !token.Valid {
            fmt.Printf("Token is not valid\n")
            return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
        }

        claims := token.Claims.(jwt.MapClaims)
        userID := uint(claims["user_id"].(float64))

        c.Locals("userID", userID)
        return c.Next()
    }
}
