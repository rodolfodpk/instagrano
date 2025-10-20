package config

import (
	"os"
	"go.uber.org/zap"
)

type Config struct {
	DatabaseURL string
	S3Endpoint  string
	S3Bucket    string
	JWTSecret   string
	Port        string
	LogLevel    string
	LogFormat   string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/instagrano?sslmode=disable"),
		S3Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:4566"),
		S3Bucket:    getEnv("S3_BUCKET", "instagrano-media"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret"),
		Port:        getEnv("PORT", "3000"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		LogFormat:   getEnv("LOG_FORMAT", "json"),
	}
}

func (c *Config) GetZapLevel() zap.AtomicLevel {
	switch c.LogLevel {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
