package tests

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/config"
)

func TestConfig_Load(t *testing.T) {
	RegisterTestingT(t)

	// Given: Environment variables are set
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb")
	os.Setenv("S3_ENDPOINT", "http://test-s3:4566")
	os.Setenv("S3_BUCKET", "test-bucket")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("PORT", "8080")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "console")
	os.Setenv("DEFAULT_PAGE_SIZE", "10")
	os.Setenv("MAX_PAGE_SIZE", "50")
	os.Setenv("REDIS_ADDR", "localhost:6380")
	os.Setenv("REDIS_PASSWORD", "testpass")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("CACHE_TTL", "10m")

	defer func() {
		// Clean up environment variables
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("S3_ENDPOINT")
		os.Unsetenv("S3_BUCKET")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("PORT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
		os.Unsetenv("DEFAULT_PAGE_SIZE")
		os.Unsetenv("MAX_PAGE_SIZE")
		os.Unsetenv("REDIS_ADDR")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("CACHE_TTL")
	}()

	// When: Load configuration
	cfg := config.Load()

	// Then: Configuration should be loaded from environment
	Expect(cfg.DatabaseURL).To(Equal("postgres://test:test@localhost:5432/testdb"))
	Expect(cfg.S3Endpoint).To(Equal("http://test-s3:4566"))
	Expect(cfg.S3Bucket).To(Equal("test-bucket"))
	Expect(cfg.JWTSecret).To(Equal("test-secret"))
	Expect(cfg.Port).To(Equal("8080"))
	Expect(cfg.LogLevel).To(Equal("debug"))
	Expect(cfg.LogFormat).To(Equal("console"))
	Expect(cfg.DefaultPageSize).To(Equal(10))
	Expect(cfg.MaxPageSize).To(Equal(50))
	Expect(cfg.RedisAddr).To(Equal("localhost:6380"))
	Expect(cfg.RedisPassword).To(Equal("testpass"))
	Expect(cfg.RedisDB).To(Equal(1))
	Expect(cfg.CacheTTL).To(Equal(10 * time.Minute))
}

func TestConfig_LoadWithDefaults(t *testing.T) {
	RegisterTestingT(t)

	// Given: No environment variables are set
	// Clean up any existing environment variables
	envVars := []string{
		"DATABASE_URL", "S3_ENDPOINT", "S3_BUCKET", "JWT_SECRET", "PORT",
		"LOG_LEVEL", "LOG_FORMAT", "DEFAULT_PAGE_SIZE", "MAX_PAGE_SIZE",
		"REDIS_ADDR", "REDIS_PASSWORD", "REDIS_DB", "CACHE_TTL",
	}
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	// When: Load configuration
	cfg := config.Load()

	// Then: Configuration should use default values
	Expect(cfg.DatabaseURL).To(Equal("postgres://postgres:postgres@localhost:5433/instagrano?sslmode=disable"))
	Expect(cfg.S3Endpoint).To(Equal("http://localhost:4566"))
	Expect(cfg.S3Bucket).To(Equal("instagrano-media"))
	Expect(cfg.JWTSecret).To(Equal("dev-secret"))
	Expect(cfg.Port).To(Equal("3000"))
	Expect(cfg.LogLevel).To(Equal("info"))
	Expect(cfg.LogFormat).To(Equal("json"))
	Expect(cfg.DefaultPageSize).To(Equal(20))
	Expect(cfg.MaxPageSize).To(Equal(100))
	Expect(cfg.RedisAddr).To(Equal("localhost:6379"))
	Expect(cfg.RedisPassword).To(Equal(""))
	Expect(cfg.RedisDB).To(Equal(0))
	Expect(cfg.CacheTTL).To(Equal(5 * time.Minute))
}

func TestConfig_GetZapLevel(t *testing.T) {
	RegisterTestingT(t)

	// Test different log levels
	testCases := []struct {
		level    string
		expected string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
		{"invalid", "info"}, // Should default to info
	}

	for _, tc := range testCases {
		// Given: Config with specific log level
		cfg := &config.Config{LogLevel: tc.level}

		// When: Get Zap level
		zapLevel := cfg.GetZapLevel()

		// Then: Should return correct level
		Expect(zapLevel.String()).To(Equal(tc.expected))
	}
}

func TestConfig_InvalidDuration(t *testing.T) {
	RegisterTestingT(t)

	// Given: Invalid duration in environment
	os.Setenv("CACHE_TTL", "invalid-duration")
	defer os.Unsetenv("CACHE_TTL")

	// When: Load configuration
	cfg := config.Load()

	// Then: Should use default duration
	Expect(cfg.CacheTTL).To(Equal(5 * time.Minute))
}

func TestConfig_InvalidInt(t *testing.T) {
	RegisterTestingT(t)

	// Given: Invalid integer in environment
	os.Setenv("DEFAULT_PAGE_SIZE", "not-a-number")
	defer os.Unsetenv("DEFAULT_PAGE_SIZE")

	// When: Load configuration
	cfg := config.Load()

	// Then: Should use default integer
	Expect(cfg.DefaultPageSize).To(Equal(20))
}

func TestConfig_PartialEnvironment(t *testing.T) {
	RegisterTestingT(t)

	// Given: Only some environment variables are set
	os.Setenv("JWT_SECRET", "custom-secret")
	os.Setenv("PORT", "9999")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("PORT")
	}()

	// When: Load configuration
	cfg := config.Load()

	// Then: Should use custom values where set, defaults elsewhere
	Expect(cfg.JWTSecret).To(Equal("custom-secret"))
	Expect(cfg.Port).To(Equal("9999"))
	Expect(cfg.DatabaseURL).To(Equal("postgres://postgres:postgres@localhost:5433/instagrano?sslmode=disable")) // Default
	Expect(cfg.LogLevel).To(Equal("info"))                                                                      // Default
}
