package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/svbnbyrk/go-message-dispatcher/internal/shared/logger"
	"go.uber.org/zap"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	startTime time.Time
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// Health handles the basic health check endpoint
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger.DebugCtx(ctx, "Health check requested")

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Uptime:    time.Since(h.startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	logger.DebugCtx(ctx, "Health check completed", zap.String("status", "ok"))
}
