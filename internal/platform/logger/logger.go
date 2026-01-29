package logger

import (
	"context"
	"fmt"
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
func With(fields ...zap.Field) *zap.Logger {
	return Get().With(fields...)
}

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

type contextKey string

const (
	reqIDKey contextKey = "req_id"
)

const (
	SeverityP0Critical = "P0-Critical"
	SeverityP1High     = "P1-High"
	SeverityP2Medium   = "P2-Medium"
)

func ContextWithReqID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, reqIDKey, reqID)
}

func ReqID(ctx context.Context) string {
	if v := ctx.Value(reqIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

func RequestFields(ctx context.Context) []zap.Field {
	return []zap.Field{
		zap.String("traceId", ReqID(ctx)),
	}
}

func WithRequest(ctx context.Context, fields ...zap.Field) *zap.Logger {
	return Get().With(append(RequestFields(ctx), fields...)...)
}

func Msg(domain, component, layer, msg string) string {
	return fmt.Sprintf("[%s][%s][%s]: %s", domain, component, layer, msg)
}

func Alert(severity, domain, component, layer, msg string) string {
	return fmt.Sprintf("Alert Severity:%s, [%s][%s][%s]: %s", severity, domain, component, layer, msg)
}
