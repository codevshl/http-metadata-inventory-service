package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *zap.Logger
	once     sync.Once
)

// Init initializes the singleton logger instance
func Init(logLevel string, isProduction bool) {
	once.Do(func() {
		var config zap.Config

		if isProduction {
			config = zap.NewProductionConfig()
		} else {
			config = zap.NewDevelopmentConfig()
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}

		// Parse log level
		level, err := zapcore.ParseLevel(logLevel)
		if err != nil {
			level = zapcore.InfoLevel
		}
		config.Level = zap.NewAtomicLevelAt(level)

		logger, err := config.Build()
		if err != nil {
			// Fallback to basic logger if zap fails
			panic(err)
		}

		instance = logger
	})
}

// Get returns the singleton logger instance
func Get() *zap.Logger {
	if instance == nil {
		// Fallback for tests or if Init wasn't called (safe default)
		return zap.NewExample()
	}
	return instance
}

// Sync flushes any buffered log entries
func Sync() {
	if instance != nil {
		_ = instance.Sync()
	}
}

// Helper functions for easy access
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}
