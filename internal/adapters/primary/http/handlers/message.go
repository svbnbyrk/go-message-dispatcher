package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/dto"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/message"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/usecases"
)

// MessageHandler handles HTTP requests for message operations
type MessageHandler struct {
	messageManagement usecases.MessageManagementUseCase
	messageProcessing usecases.MessageProcessingUseCase
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	messageManagement usecases.MessageManagementUseCase,
	messageProcessing usecases.MessageProcessingUseCase,
) *MessageHandler {
	return &MessageHandler{
		messageManagement: messageManagement,
		messageProcessing: messageProcessing,
	}
}

// CreateMessage handles POST /api/v1/messages
func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := usecases.CreateMessageCommand{
		PhoneNumber: req.PhoneNumber,
		Content:     req.Content,
	}

	result, err := h.messageManagement.CreateMessage(r.Context(), cmd)
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	response := dto.MessageResponse{
		ID:          result.ID.String(),
		PhoneNumber: result.PhoneNumber,
		Content:     result.Content,
		Status:      result.Status,
		ExternalID:  result.ExternalID,
		RetryCount:  result.RetryCount,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
		SentAt:      result.SentAt,
	}

	writeJSONResponse(w, http.StatusCreated, response)
}

// GetMessage handles GET /api/v1/messages/{id}
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid message ID")
		return
	}

	result, err := h.messageManagement.GetMessageByID(r.Context(), id)
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	response := dto.MessageResponse{
		ID:          result.ID.String(),
		PhoneNumber: result.PhoneNumber,
		Content:     result.Content,
		Status:      result.Status,
		ExternalID:  result.ExternalID,
		RetryCount:  result.RetryCount,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
		SentAt:      result.SentAt,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// ListMessages handles GET /api/v1/messages
func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	query := usecases.ListMessagesQuery{
		Limit:  20, // Default limit
		Offset: 0,  // Default offset
	}

	// Parse query parameters
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			query.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := message.Status(statusStr)
		if status.IsValid() {
			query.Status = &status
		}
	}

	result, err := h.messageManagement.ListMessages(r.Context(), query)
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	// Convert to response DTOs
	responses := make([]dto.MessageResponse, len(result.Messages))
	for i, msg := range result.Messages {
		responses[i] = dto.MessageResponse{
			ID:          msg.ID.String(),
			PhoneNumber: msg.PhoneNumber,
			Content:     msg.Content,
			Status:      msg.Status,
			ExternalID:  msg.ExternalID,
			RetryCount:  msg.RetryCount,
			CreatedAt:   msg.CreatedAt,
			UpdatedAt:   msg.UpdatedAt,
			SentAt:      msg.SentAt,
		}
	}

	// Set total count in header as requested
	w.Header().Set("X-Total-Count", strconv.FormatInt(result.TotalCount, 10))
	writeJSONResponse(w, http.StatusOK, responses)
}

// GetProcessingStatus handles GET /api/v1/messaging/status
func (h *MessageHandler) GetProcessingStatus(w http.ResponseWriter, r *http.Request) {
	result, err := h.messageProcessing.GetProcessingStatus(r.Context())
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	response := dto.ProcessingStatusResponse{
		IsProcessing:     result.IsProcessing,
		LastProcessedAt:  result.LastProcessedAt,
		PendingCount:     result.PendingCount,
		ProcessedToday:   result.ProcessedToday,
		FailedToday:      result.FailedToday,
		NextProcessingAt: result.NextProcessingAt,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// ProcessMessages handles POST /api/v1/messaging/process - Manual processing trigger for testing
func (h *MessageHandler) ProcessMessages(w http.ResponseWriter, r *http.Request) {
	// Get batch size from query param, default to 2
	batchSize := 2
	if batchStr := r.URL.Query().Get("batch_size"); batchStr != "" {
		if size, err := strconv.Atoi(batchStr); err == nil && size > 0 && size <= 10 {
			batchSize = size
		}
	}

	result, err := h.messageProcessing.ProcessPendingMessages(r.Context(), batchSize)
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	response := dto.ProcessingResultResponse{
		ProcessedCount: result.ProcessedCount,
		SuccessCount:   result.SuccessCount,
		FailedCount:    result.FailedCount,
		Errors:         make([]string, len(result.Errors)),
	}

	// Convert errors to strings
	for i, err := range result.Errors {
		response.Errors[i] = err.Error()
	}

	writeJSONResponse(w, http.StatusOK, response)
}
