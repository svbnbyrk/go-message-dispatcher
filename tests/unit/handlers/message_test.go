package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/handlers"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/usecases"
)

// Mock use cases for testing
type mockMessageManagement struct {
	shouldFail bool
}

func (m *mockMessageManagement) CreateMessage(ctx context.Context, cmd usecases.CreateMessageCommand) (*usecases.MessageResponse, error) {
	if m.shouldFail {
		return nil, &testError{message: "create failed"}
	}

	return &usecases.MessageResponse{
		ID:          uuid.New(),
		PhoneNumber: cmd.PhoneNumber,
		Content:     cmd.Content,
		Status:      "pending",
		RetryCount:  0,
	}, nil
}

func (m *mockMessageManagement) GetMessageByID(ctx context.Context, id uuid.UUID) (*usecases.MessageResponse, error) {
	if m.shouldFail {
		return nil, &testError{message: "not found"}
	}

	return &usecases.MessageResponse{
		ID:          id,
		PhoneNumber: "+905551234567",
		Content:     "Test message",
		Status:      "sent",
		RetryCount:  0,
	}, nil
}

func (m *mockMessageManagement) ListMessages(ctx context.Context, query usecases.ListMessagesQuery) (*usecases.ListMessagesResponse, error) {
	if m.shouldFail {
		return nil, &testError{message: "list failed"}
	}

	return &usecases.ListMessagesResponse{
		Messages:   []usecases.MessageResponse{},
		TotalCount: 0,
		HasMore:    false,
	}, nil
}

type mockMessageProcessing struct {
	shouldFail bool
}

func (m *mockMessageProcessing) ProcessPendingMessages(ctx context.Context, batchSize int) (*usecases.ProcessingResult, error) {
	return nil, nil
}

func (m *mockMessageProcessing) GetProcessingStatus(ctx context.Context) (*usecases.ProcessingStatus, error) {
	if m.shouldFail {
		return nil, &testError{message: "status failed"}
	}

	return &usecases.ProcessingStatus{
		IsProcessing:   false,
		PendingCount:   5,
		ProcessedToday: 10,
		FailedToday:    1,
	}, nil
}

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestMessageHandler_CreateMessage(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		shouldFail     bool
		expectedStatus int
	}{
		{
			name:           "successful creation",
			requestBody:    `{"phoneNumber":"+905551234567","content":"Test message"}`,
			shouldFail:     false,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request body",
			requestBody:    `{invalid json}`, // Actually invalid JSON
			shouldFail:     false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "use case error",
			requestBody:    `{"phoneNumber":"+905551234567","content":"Test message"}`,
			shouldFail:     true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgmt := &mockMessageManagement{shouldFail: tt.shouldFail}
			mockProc := &mockMessageProcessing{}

			handler := handlers.NewMessageHandler(mockMgmt, mockProc)

			req := httptest.NewRequest("POST", "/api/v1/messages", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.CreateMessage(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Check content type
			if w.Header().Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestMessageHandler_GetMessage(t *testing.T) {
	tests := []struct {
		name           string
		messageID      string
		shouldFail     bool
		expectedStatus int
	}{
		{
			name:           "successful retrieval",
			messageID:      uuid.New().String(),
			shouldFail:     false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid message ID",
			messageID:      "invalid-uuid",
			shouldFail:     false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "use case error",
			messageID:      uuid.New().String(),
			shouldFail:     true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgmt := &mockMessageManagement{shouldFail: tt.shouldFail}
			mockProc := &mockMessageProcessing{}

			handler := handlers.NewMessageHandler(mockMgmt, mockProc)

			// Create router to test URL params
			r := chi.NewRouter()
			r.Get("/messages/{id}", handler.GetMessage)

			req := httptest.NewRequest("GET", "/messages/"+tt.messageID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestMessageHandler_ListMessages(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		shouldFail     bool
		expectedStatus int
	}{
		{
			name:           "successful listing",
			queryParams:    "",
			shouldFail:     false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "with query parameters",
			queryParams:    "?limit=10&offset=0&status=pending",
			shouldFail:     false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "use case error",
			queryParams:    "",
			shouldFail:     true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgmt := &mockMessageManagement{shouldFail: tt.shouldFail}
			mockProc := &mockMessageProcessing{}

			handler := handlers.NewMessageHandler(mockMgmt, mockProc)

			req := httptest.NewRequest("GET", "/api/v1/messages"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ListMessages(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Check for X-Total-Count header on successful requests
			if tt.expectedStatus == http.StatusOK {
				if w.Header().Get("X-Total-Count") == "" {
					t.Error("Expected X-Total-Count header to be set")
				}
			}
		})
	}
}
