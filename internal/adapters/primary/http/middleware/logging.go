package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/svbnbyrk/go-message-dispatcher/internal/shared/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// RequestLogging creates a middleware for structured request logging
func RequestLogging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate correlation ID if not present
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			// Add correlation ID to response header
			w.Header().Set("X-Correlation-ID", correlationID)

			// Get request ID from Chi middleware
			requestID := middleware.GetReqID(r.Context())

			// Create context with correlation ID and request ID
			ctx := context.WithValue(r.Context(), logger.CorrelationIDKey, correlationID)
			ctx = context.WithValue(ctx, logger.RequestIDKey, requestID)

			// Create a response writer wrapper to capture status and size
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Log request start
			logger.InfoCtx(ctx, "HTTP request started",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.String("user_agent", r.UserAgent()),
				zap.String("client_ip", getClientIP(r)),
				zap.String("remote_addr", r.RemoteAddr),
			)

			// Process request
			next.ServeHTTP(ww, r.WithContext(ctx))

			// Calculate duration
			duration := time.Since(start)

			// Log request completion
			logger.LogHTTPRequest(
				ctx,
				r.Method,
				r.URL.Path,
				r.UserAgent(),
				getClientIP(r),
				ww.Status(),
				duration.Milliseconds(),
				int64(ww.BytesWritten()),
			)

			// Log errors for 4xx and 5xx responses
			if ww.Status() >= 400 {
				level := "warn"
				if ww.Status() >= 500 {
					level = "error"
				}

				logger.WithContext(ctx).Log(
					getLogLevel(level),
					"HTTP request completed with error",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status_code", ww.Status()),
					zap.Int64("duration_ms", duration.Milliseconds()),
				)
			}
		})
	}
}

// PanicRecovery creates a middleware for panic recovery with structured logging
func PanicRecovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					ctx := r.Context()

					logger.ErrorCtx(ctx, "Panic recovered",
						nil,
						zap.Any("panic", rvr),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String("user_agent", r.UserAgent()),
						zap.String("client_ip", getClientIP(r)),
					)

					// Return 500 Internal Server Error
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP from various headers
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (most common)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Check X-Forwarded header
	if xf := r.Header.Get("X-Forwarded"); xf != "" {
		return xf
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// getLogLevel converts string level to zap level
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}
