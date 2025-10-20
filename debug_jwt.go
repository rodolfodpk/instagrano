package main

import (
    "fmt"
    "log"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

func main() {
    jwtSecret := "super-secret-key-for-testing"
    
    // Generate token
    claims := jwt.MapClaims{
        "user_id": 4,
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(jwtSecret))
    if err != nil {
        log.Fatal("Error generating token:", err)
    }
    
    fmt.Printf("Generated token: %s\n", tokenString)
    
    // Parse token
    parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("invalid signing method")
        }
        return []byte(jwtSecret), nil
    })
    
    if err != nil {
        log.Fatal("Error parsing token:", err)
    }
    
    if !parsedToken.Valid {
        log.Fatal("Token is not valid")
    }
    
    claims = parsedToken.Claims.(jwt.MapClaims)
    userID := claims["user_id"].(float64)
    
    fmt.Printf("Token is valid! User ID: %v\n", userID)
}
