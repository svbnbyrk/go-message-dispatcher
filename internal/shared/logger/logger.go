package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const (
	CorrelationIDKey contextKey = "correlation_id"
	UserIDKey        contextKey = "user_id"
	RequestIDKey     contextKey = "request_id"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Environment string `mapstructure:"environment"`
	EnableJSON  bool   `mapstructure:"enable_json"`
}

// Initialize initializes the global logger
func Initialize(config LoggerConfig) error {
	var zapConfig zap.Config

	// Configure based on environment
	if config.Environment == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.MessageKey = "message"
		zapConfig.EncoderConfig.LevelKey = "level"
		zapConfig.EncoderConfig.CallerKey = "caller"
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		if !config.EnableJSON {
			zapConfig.Encoding = "console"
		}
	}

	// Set log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Add caller information
	zapConfig.DisableCaller = false
	zapConfig.DisableStacktrace = config.Environment == "production"

	// Build logger
	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1), // Skip one level for wrapper functions
	)
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	Logger = logger
	Sugar = logger.Sugar()

	return nil
}

// WithContext creates a logger with context fields
func WithContext(ctx context.Context) *zap.Logger {
	logger := Logger

	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		logger = logger.With(zap.String("correlation_id", correlationID.(string)))
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		logger = logger.With(zap.String("user_id", userID.(string)))
	}

	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		logger = logger.With(zap.String("request_id", requestID.(string)))
	}

	return logger
}

// WithContextSugar creates a sugared logger with context fields
func WithContextSugar(ctx context.Context) *zap.SugaredLogger {
	return WithContext(ctx).Sugar()
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// InfoCtx logs an info message with context
func InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Info(msg, fields...)
}

// Error logs an error message
func Error(msg string, err error, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	Logger.Error(msg, allFields...)
}

// ErrorCtx logs an error message with context
func ErrorCtx(ctx context.Context, msg string, err error, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	WithContext(ctx).Error(msg, allFields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// WarnCtx logs a warning message with context
func WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Warn(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// DebugCtx logs a debug message with context
func DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Debug(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Close gracefully closes the logger
func Close() {
	if Logger != nil {
		Logger.Sync()
	}
}

// LogSystemInfo logs system information at startup
func LogSystemInfo(appName, version, environment string) {
	hostname, _ := os.Hostname()
	pid := os.Getpid()

	Info("Application starting",
		zap.String("app_name", appName),
		zap.String("version", version),
		zap.String("environment", environment),
		zap.String("hostname", hostname),
		zap.Int("pid", pid),
	)
}

// LogHTTPRequest logs HTTP request information
func LogHTTPRequest(ctx context.Context, method, path, userAgent, clientIP string, statusCode int, duration int64, responseSize int64) {
	InfoCtx(ctx, "HTTP request completed",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("user_agent", userAgent),
		zap.String("client_ip", clientIP),
		zap.Int("status_code", statusCode),
		zap.Int64("duration_ms", duration),
		zap.Int64("response_size", responseSize),
	)
}

// LogDatabaseOperation logs database operations
func LogDatabaseOperation(ctx context.Context, operation, table string, duration int64, affected int64) {
	InfoCtx(ctx, "Database operation completed",
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Int64("duration_ms", duration),
		zap.Int64("affected_rows", affected),
	)
}

// LogWebhookRequest logs webhook requests
func LogWebhookRequest(ctx context.Context, url string, statusCode int, duration int64, success bool) {
	if success {
		InfoCtx(ctx, "Webhook request successful",
			zap.String("url", url),
			zap.Int("status_code", statusCode),
			zap.Int64("duration_ms", duration),
		)
	} else {
		WarnCtx(ctx, "Webhook request failed",
			zap.String("url", url),
			zap.Int("status_code", statusCode),
			zap.Int64("duration_ms", duration),
		)
	}
}

// LogMessageProcessing logs message processing events
func LogMessageProcessing(ctx context.Context, messageID string, status string, retryCount int, phoneNumber string) {
	InfoCtx(ctx, "Message processing event",
		zap.String("message_id", messageID),
		zap.String("status", status),
		zap.Int("retry_count", retryCount),
		zap.String("phone_number", phoneNumber),
	)
}

// LogSchedulerEvent logs background scheduler events
func LogSchedulerEvent(ctx context.Context, event string, batchSize int, processed int, successful int, failed int) {
	InfoCtx(ctx, "Scheduler event",
		zap.String("event", event),
		zap.Int("batch_size", batchSize),
		zap.Int("processed", processed),
		zap.Int("successful", successful),
		zap.Int("failed", failed),
	)
}
