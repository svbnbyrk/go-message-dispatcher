package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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
// @Summary      Create a new message
// @Description  Create a new message to be sent via webhook with automatic retry mechanism
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        message  body      dto.CreateMessageRequest  true  "Message data"
// @Success      201      {object}  dto.MessageResponse
// @Failure      400      {object}  dto.ErrorResponse "Bad request - invalid input"
// @Failure      401      {object}  dto.ErrorResponse "Unauthorized - missing or invalid API key"
// @Failure      429      {object}  dto.ErrorResponse "Rate limit exceeded"
// @Failure      500      {object}  dto.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /messages [post]
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
// @Summary      Get a message by ID
// @Description  Retrieve a specific message by its unique identifier
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Message ID (UUID)"
// @Success      200  {object}  dto.MessageResponse
// @Failure      400  {object}  dto.ErrorResponse "Bad request - invalid message ID"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      404  {object}  dto.ErrorResponse "Message not found"
// @Failure      429  {object}  dto.ErrorResponse "Rate limit exceeded"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /messages/{id} [get]
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
// @Summary      List messages with pagination
// @Description  Retrieve a list of messages with optional filtering and pagination
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        status  query     string  false  "Filter by status"  Enums(pending, sent, failed)
// @Param        limit   query     int     false  "Number of messages to return (1-100)"  minimum(1)  maximum(100)  default(20)
// @Param        offset  query     int     false  "Number of messages to skip"  minimum(0)  default(0)
// @Success      200     {array}   dto.MessageResponse
// @Header       200     {string}  X-Total-Count  "Total number of messages"
// @Failure      400     {object}  dto.ErrorResponse "Bad request - invalid parameters"
// @Failure      401     {object}  dto.ErrorResponse "Unauthorized"
// @Failure      429     {object}  dto.ErrorResponse "Rate limit exceeded"
// @Failure      500     {object}  dto.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /messages [get]
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
		status := message.Status(strings.ToUpper(statusStr))
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
// @Summary      Get processing status
// @Description  Retrieve the current message processing status and statistics
// @Tags         messaging
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.ProcessingStatusResponse
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      429  {object}  dto.ErrorResponse "Rate limit exceeded"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /messaging/status [get]
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
// @Summary      Process messages manually
// @Description  Manually trigger message processing for testing purposes
// @Tags         messaging
// @Accept       json
// @Produce      json
// @Param        batch_size  query     int  false  "Number of messages to process (1-10)"  minimum(1)  maximum(10)  default(2)
// @Success      200         {object}  dto.ProcessingResultResponse
// @Failure      401         {object}  dto.ErrorResponse "Unauthorized"
// @Failure      429         {object}  dto.ErrorResponse "Rate limit exceeded"
// @Failure      500         {object}  dto.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /messaging/process [post]
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
