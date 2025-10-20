package tests

import (
	"testing"

	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rodolfodpk/instagrano/internal/logger"
)

func TestLogger_New(t *testing.T) {
	RegisterTestingT(t)

	// Test JSON format
	jsonLogger := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")
	Expect(jsonLogger).NotTo(BeNil())

	// Test console format
	consoleLogger := logger.New(zap.NewAtomicLevelAt(zap.DebugLevel), "console")
	Expect(consoleLogger).NotTo(BeNil())

	// Test invalid format (should default to JSON)
	invalidLogger := logger.New(zap.NewAtomicLevelAt(zap.WarnLevel), "invalid")
	Expect(invalidLogger).NotTo(BeNil())
}

func TestLogger_WithRequestID(t *testing.T) {
	RegisterTestingT(t)

	// Given: A logger instance
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

	// When: Add request ID
	loggerWithID := log.WithRequestID("req-123")

	// Then: Should return a new logger instance
	Expect(loggerWithID).NotTo(BeNil())
	Expect(loggerWithID).NotTo(Equal(log)) // Should be a new instance
}

func TestLogger_WithUserID(t *testing.T) {
	RegisterTestingT(t)

	// Given: A logger instance
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

	// When: Add user ID
	loggerWithUserID := log.WithUserID(123)

	// Then: Should return a new logger instance
	Expect(loggerWithUserID).NotTo(BeNil())
	Expect(loggerWithUserID).NotTo(Equal(log)) // Should be a new instance
}

func TestLogger_WithField(t *testing.T) {
	RegisterTestingT(t)

	// Given: A logger instance
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

	// When: Add a field
	loggerWithField := log.WithField("key", "value")

	// Then: Should return a new logger instance
	Expect(loggerWithField).NotTo(BeNil())
	Expect(loggerWithField).NotTo(Equal(log)) // Should be a new instance
}

func TestLogger_WithFields(t *testing.T) {
	RegisterTestingT(t)

	// Given: A logger instance
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

	// When: Add multiple fields
	fields := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
		"field3": true,
	}
	loggerWithFields := log.WithFields(fields)

	// Then: Should return a new logger instance
	Expect(loggerWithFields).NotTo(BeNil())
	Expect(loggerWithFields).NotTo(Equal(log)) // Should be a new instance
}

func TestLogger_Chaining(t *testing.T) {
	RegisterTestingT(t)

	// Given: A logger instance
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

	// When: Chain multiple operations
	chainedLogger := log.
		WithRequestID("req-456").
		WithUserID(789).
		WithField("operation", "test").
		WithFields(map[string]interface{}{
			"additional": "data",
		})

	// Then: Should return a new logger instance
	Expect(chainedLogger).NotTo(BeNil())
	Expect(chainedLogger).NotTo(Equal(log)) // Should be a new instance
}

func TestLogger_DifferentLevels(t *testing.T) {
	RegisterTestingT(t)

	// Test all log levels
	levels := []zapcore.Level{
		zap.DebugLevel,
		zap.InfoLevel,
		zap.WarnLevel,
		zap.ErrorLevel,
	}

	for _, level := range levels {
		// Given: Logger with specific level
		log := logger.New(zap.NewAtomicLevelAt(level), "json")

		// Then: Should create logger successfully
		Expect(log).NotTo(BeNil())
	}
}

func TestLogger_DifferentFormats(t *testing.T) {
	RegisterTestingT(t)

	// Test different formats
	formats := []string{"json", "console", "invalid"}

	for _, format := range formats {
		// Given: Logger with specific format
		log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), format)

		// Then: Should create logger successfully (invalid format should default to JSON)
		Expect(log).NotTo(BeNil())
	}
}
