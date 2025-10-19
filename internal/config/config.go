package config

import "os"

type Config struct {
    DatabaseURL string
    JWTSecret   string
    Port        string
}

func Load() *Config {
    return &Config{
        DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/instagrano?sslmode=disable"),
        JWTSecret:   getEnv("JWT_SECRET", "dev-secret"),
        Port:        getEnv("PORT", "3000"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
