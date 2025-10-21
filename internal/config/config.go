package config

import (
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	DatabaseURL     string
	S3Endpoint      string
	S3Bucket        string
	JWTSecret       string
	Port            string
	LogLevel        string
	LogFormat       string
	DefaultPageSize int
	MaxPageSize     int
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	CacheTTL        time.Duration
}

func Load() *Config {
	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/instagrano?sslmode=disable"),
		S3Endpoint:      getEnv("S3_ENDPOINT", "http://localhost:4566"),
		S3Bucket:        getEnv("S3_BUCKET", "instagrano-media"),
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret"),
		Port:            getEnv("PORT", "8080"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		LogFormat:       getEnv("LOG_FORMAT", "json"),
		DefaultPageSize: getEnvInt("DEFAULT_PAGE_SIZE", 20),
		MaxPageSize:     getEnvInt("MAX_PAGE_SIZE", 100),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         getEnvInt("REDIS_DB", 0),
		CacheTTL:        getDurationEnv("CACHE_TTL", 5*time.Minute),
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

func getEnvInt(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
