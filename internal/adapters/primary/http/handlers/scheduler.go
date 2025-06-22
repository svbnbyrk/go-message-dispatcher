package handlers

import (
	"net/http"
	"time"

	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/dto"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/services"
)

// SchedulerHandler handles HTTP requests for scheduler operations
type SchedulerHandler struct {
	scheduler *services.ProcessingScheduler
	interval  time.Duration
	batchSize int
}

// NewSchedulerHandler creates a new scheduler handler
func NewSchedulerHandler(scheduler *services.ProcessingScheduler, interval time.Duration, batchSize int) *SchedulerHandler {
	return &SchedulerHandler{
		scheduler: scheduler,
		interval:  interval,
		batchSize: batchSize,
	}
}

// GetSchedulerStatus handles GET /api/v1/scheduler/status
// @Summary      Get scheduler status
// @Description  Retrieve the current status and statistics of the background message scheduler
// @Tags         scheduler
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.SchedulerStatusResponse
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      429  {object}  dto.ErrorResponse "Rate limit exceeded"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /scheduler/status [get]
func (h *SchedulerHandler) GetSchedulerStatus(w http.ResponseWriter, r *http.Request) {
	stats := h.scheduler.GetStats()

	// Calculate next processing time
	nextProcessingTime := stats.LastProcessingTime.Add(h.interval)
	nextProcessingIn := time.Until(nextProcessingTime)
	if nextProcessingIn < 0 {
		nextProcessingIn = 0
	}

	response := dto.SchedulerStatusResponse{
		IsRunning:             h.scheduler.IsRunning(),
		IsCurrentlyProcessing: stats.IsCurrentlyProcessing,
		TotalProcessed:        stats.TotalProcessed,
		TotalSuccessful:       stats.TotalSuccessful,
		TotalFailed:           stats.TotalFailed,
		LastProcessingTime:    stats.LastProcessingTime,
		NextProcessingIn:      nextProcessingIn.Round(time.Second).String(),
		Interval:              h.interval.String(),
		BatchSize:             h.batchSize,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// StartScheduler handles POST /api/v1/scheduler/start
// @Summary      Start scheduler
// @Description  Start the background message processing scheduler
// @Tags         scheduler
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string "Scheduler started successfully"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      409  {object}  dto.ErrorResponse "Scheduler already running"
// @Failure      429  {object}  dto.ErrorResponse "Rate limit exceeded"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /scheduler/start [post]
func (h *SchedulerHandler) StartScheduler(w http.ResponseWriter, r *http.Request) {
	if h.scheduler.IsRunning() {
		writeJSONError(w, http.StatusConflict, "scheduler is already running")
		return
	}

	if err := h.scheduler.Start(r.Context()); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to start scheduler: "+err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "scheduler started successfully",
		"status":  "running",
	})
}

// StopScheduler handles POST /api/v1/scheduler/stop
// @Summary      Stop scheduler
// @Description  Stop the background message processing scheduler
// @Tags         scheduler
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string "Scheduler stopped successfully"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      409  {object}  dto.ErrorResponse "Scheduler not running"
// @Failure      429  {object}  dto.ErrorResponse "Rate limit exceeded"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /scheduler/stop [post]
func (h *SchedulerHandler) StopScheduler(w http.ResponseWriter, r *http.Request) {
	if !h.scheduler.IsRunning() {
		writeJSONError(w, http.StatusConflict, "scheduler is not running")
		return
	}

	if err := h.scheduler.Stop(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to stop scheduler: "+err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "scheduler stopped successfully",
		"status":  "stopped",
	})
}
