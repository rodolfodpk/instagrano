package config

import "os"

type Config struct {
	DatabaseURL string
	S3Endpoint  string
	S3Bucket    string
	JWTSecret   string
	Port        string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/instagrano?sslmode=disable"),
		S3Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:4566"),
		S3Bucket:    getEnv("S3_BUCKET", "instagrano-media"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret"),
		Port:        getEnv("PORT", "3000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
