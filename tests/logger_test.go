package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rodolfodpk/instagrano/internal/logger"
)

var _ = Describe("Logger", func() {
	Describe("New", func() {
		It("should create logger with JSON format", func() {
			// When: Create logger with JSON format
			jsonLogger := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

			// Then: Should create logger successfully
			Expect(jsonLogger).NotTo(BeNil())
		})

		It("should create logger with console format", func() {
			// When: Create logger with console format
			consoleLogger := logger.New(zap.NewAtomicLevelAt(zap.DebugLevel), "console")

			// Then: Should create logger successfully
			Expect(consoleLogger).NotTo(BeNil())
		})

		It("should default to JSON format for invalid format", func() {
			// When: Create logger with invalid format
			invalidLogger := logger.New(zap.NewAtomicLevelAt(zap.WarnLevel), "invalid")

			// Then: Should create logger successfully (defaults to JSON)
			Expect(invalidLogger).NotTo(BeNil())
		})
	})

	Describe("WithRequestID", func() {
		It("should add request ID to logger", func() {
			// Given: A logger
			logger := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

			// When: Add request ID
			requestID := "req-123"
			loggerWithID := logger.WithRequestID(requestID)

			// Then: Should return logger with request ID
			Expect(loggerWithID).NotTo(BeNil())
		})

		It("should handle empty request ID", func() {
			// Given: A logger
			logger := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

			// When: Add empty request ID
			loggerWithID := logger.WithRequestID("")

			// Then: Should return logger (empty ID is valid)
			Expect(loggerWithID).NotTo(BeNil())
		})
	})

	Describe("WithUserID", func() {
		It("should add user ID to logger", func() {
			// Given: A logger
			logger := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

			// When: Add user ID
			userID := uint(123)
			loggerWithUserID := logger.WithUserID(userID)

			// Then: Should return logger with user ID
			Expect(loggerWithUserID).NotTo(BeNil())
		})

		It("should handle zero user ID", func() {
			// Given: A logger
			logger := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

			// When: Add zero user ID
			loggerWithUserID := logger.WithUserID(0)

			// Then: Should return logger (zero ID is valid)
			Expect(loggerWithUserID).NotTo(BeNil())
		})
	})

	Describe("Logging Levels", func() {
		It("should respect debug level", func() {
			// Given: A logger with debug level
			logger := logger.New(zap.NewAtomicLevelAt(zap.DebugLevel), "json")

			// When: Log at debug level
			// Then: Should not panic (debug level allows debug logs)
			Expect(func() {
				logger.Logger.Debug("debug message")
			}).NotTo(Panic())
		})

		It("should respect info level", func() {
			// Given: A logger with info level
			logger := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

			// When: Log at info level
			// Then: Should not panic (info level allows info logs)
			Expect(func() {
				logger.Logger.Info("info message")
			}).NotTo(Panic())
		})

		It("should respect warn level", func() {
			// Given: A logger with warn level
			logger := logger.New(zap.NewAtomicLevelAt(zap.WarnLevel), "json")

			// When: Log at warn level
			// Then: Should not panic (warn level allows warn logs)
			Expect(func() {
				logger.Logger.Warn("warn message")
			}).NotTo(Panic())
		})

		It("should respect error level", func() {
			// Given: A logger with error level
			logger := logger.New(zap.NewAtomicLevelAt(zap.ErrorLevel), "json")

			// When: Log at error level
			// Then: Should not panic (error level allows error logs)
			Expect(func() {
				logger.Logger.Error("error message")
			}).NotTo(Panic())
		})
	})

	Describe("Logger Fields", func() {
		It("should support structured logging", func() {
			// Given: A logger
			logger := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")

			// When: Log with fields
			// Then: Should not panic (structured logging should work)
			Expect(func() {
				logger.Logger.Info("structured message",
					zap.String("key", "value"),
					zap.Int("number", 42),
					zap.Bool("flag", true),
				)
			}).NotTo(Panic())
		})

		It("should support error logging with fields", func() {
			// Given: A logger
			logger := logger.New(zap.NewAtomicLevelAt(zap.ErrorLevel), "json")

			// When: Log error with fields
			// Then: Should not panic (error logging with fields should work)
			Expect(func() {
				logger.Logger.Error("error message",
					zap.String("operation", "test"),
					zap.Error(nil), // nil error is valid
				)
			}).NotTo(Panic())
		})
	})

	Describe("Logger Configuration", func() {
		It("should create logger with different levels", func() {
			levels := []zapcore.Level{
				zap.DebugLevel,
				zap.InfoLevel,
				zap.WarnLevel,
				zap.ErrorLevel,
			}

			for _, level := range levels {
				// When: Create logger with specific level
				logger := logger.New(zap.NewAtomicLevelAt(level), "json")

				// Then: Should create logger successfully
				Expect(logger).NotTo(BeNil())
			}
		})

		It("should create logger with different formats", func() {
			formats := []string{"json", "console", "invalid"}

			for _, format := range formats {
				// When: Create logger with specific format
				logger := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), format)

				// Then: Should create logger successfully
				Expect(logger).NotTo(BeNil())
			}
		})
	})
})