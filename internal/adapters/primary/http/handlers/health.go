package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/dto"
	"github.com/svbnbyrk/go-message-dispatcher/internal/shared/logger"
	"go.uber.org/zap"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	startTime time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// Health handles the basic health check endpoint
// @Summary      Health check
// @Description  Get the health status of the API server
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.HealthResponse
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Router       /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger.DebugCtx(ctx, "Health check requested")

	response := dto.HealthResponse{
		Status:  "ok",
		Uptime:  time.Since(h.startTime).String(),
		Version: "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	logger.DebugCtx(ctx, "Health check completed", zap.String("status", "ok"))
}
