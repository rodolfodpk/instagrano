package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger for easier usage
type Logger struct {
	*zap.Logger
}

// New creates a new logger instance
func New(level zap.AtomicLevel, format string) *Logger {
	var config zap.Config
	
	if format == "console" {
		config = zap.NewDevelopmentConfig()
		config.Level = level
	} else {
		config = zap.NewProductionConfig()
		config.Level = level
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := config.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	return &Logger{Logger: logger}
}

// WithRequestID adds request ID to logger context
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("request_id", requestID))}
}

// WithUserID adds user ID to logger context
func (l *Logger) WithUserID(userID uint) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Uint("user_id", userID))}
}

// WithField adds a field to logger context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Any(key, value))}
}

// WithFields adds multiple fields to logger context
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &Logger{Logger: l.Logger.With(zapFields...)}
}
