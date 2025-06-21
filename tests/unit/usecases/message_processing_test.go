package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/message"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/services"
	usecaseImpl "github.com/svbnbyrk/go-message-dispatcher/internal/core/usecases"
)

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// Mock webhook service for testing
type mockWebhookService struct {
	shouldFail bool
}

func (m *mockWebhookService) SendMessage(ctx context.Context, request services.WebhookRequest) (*services.WebhookResponse, error) {
	if m.shouldFail {
		return nil, &testError{message: "webhook send failed"}
	}
	return &services.WebhookResponse{
		MessageID: "webhook-123-message-id",
	}, nil
}

func (m *mockWebhookService) IsHealthy(ctx context.Context) error {
	return nil
}

// Mock cache service for testing
type mockCacheService struct {
	data       map[string]interface{}
	shouldFail bool
}

func newMockCacheService() *mockCacheService {
	return &mockCacheService{
		data: make(map[string]interface{}),
	}
}

func (m *mockCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if m.shouldFail {
		return &testError{message: "cache set failed"}
	}
	m.data[key] = value
	return nil
}

func (m *mockCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	if m.shouldFail {
		return nil, &testError{message: "cache get failed"}
	}
	if value, exists := m.data[key]; exists {
		return value, nil
	}
	return nil, &testError{message: "key not found"}
}

func (m *mockCacheService) Delete(ctx context.Context, key string) error {
	if m.shouldFail {
		return &testError{message: "cache delete failed"}
	}
	delete(m.data, key)
	return nil
}

func (m *mockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	if m.shouldFail {
		return false, &testError{message: "cache exists failed"}
	}
	_, exists := m.data[key]
	return exists, nil
}

func (m *mockCacheService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if m.shouldFail {
		return &testError{message: "cache setJSON failed"}
	}
	m.data[key] = value
	return nil
}

func (m *mockCacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	if m.shouldFail {
		return &testError{message: "cache getJSON failed"}
	}
	// Simplified implementation for testing
	return nil
}

func (m *mockCacheService) IsHealthy(ctx context.Context) error {
	return nil
}

func TestMessageProcessingService_ProcessPendingMessages(t *testing.T) {
	tests := []struct {
		name              string
		batchSize         int
		pendingMessages   int
		shouldFailRepo    string
		expectedProcessed int
		expectError       bool
	}{
		{
			name:              "successful processing with default batch size",
			batchSize:         0,
			pendingMessages:   5,
			expectedProcessed: 2, // Default batch size is 2
			expectError:       false,
		},
		{
			name:              "successful processing with custom batch size",
			batchSize:         3,
			pendingMessages:   5,
			expectedProcessed: 3,
			expectError:       false,
		},
		{
			name:              "batch size too large - should cap at 10",
			batchSize:         15,
			pendingMessages:   5,
			expectedProcessed: 5,
			expectError:       false,
		},
		{
			name:              "no pending messages",
			batchSize:         2,
			pendingMessages:   0,
			expectedProcessed: 0,
			expectError:       false,
		},
		{
			name:            "repository error",
			batchSize:       2,
			pendingMessages: 3,
			shouldFailRepo:  "GetPendingMessages",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockMessageRepository()

			// Create pending messages
			for i := 0; i < tt.pendingMessages; i++ {
				testMsg := createTestMessage(t)
				mockRepo.messages[testMsg.ID] = testMsg
			}

			if tt.shouldFailRepo != "" {
				mockRepo.shouldFailOp = tt.shouldFailRepo
			}

			service := usecaseImpl.NewMessageProcessingService(mockRepo, &mockWebhookService{}, newMockCacheService())
			ctx := context.Background()

			result, err := service.ProcessPendingMessages(ctx, tt.batchSize)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Expected result but got nil")
				return
			}

			if result.ProcessedCount != tt.expectedProcessed {
				t.Errorf("Expected processed count %d, got %d", tt.expectedProcessed, result.ProcessedCount)
			}

			// Check that the sum of success and failed equals processed
			if result.SuccessCount+result.FailedCount != result.ProcessedCount {
				t.Errorf("Success count (%d) + Failed count (%d) should equal Processed count (%d)",
					result.SuccessCount, result.FailedCount, result.ProcessedCount)
			}

			// Verify repository calls
			if tt.shouldFailRepo == "" {
				if mockRepo.callCount["GetPendingMessages"] != 1 {
					t.Errorf("Expected GetPendingMessages to be called once, called %d times",
						mockRepo.callCount["GetPendingMessages"])
				}
			}
		})
	}
}

func TestMessageProcessingService_GetProcessingStatus(t *testing.T) {
	tests := []struct {
		name           string
		pendingCount   int
		sentCount      int
		failedCount    int
		shouldFailRepo string
		expectError    bool
	}{
		{
			name:         "successful status retrieval",
			pendingCount: 5,
			sentCount:    10,
			failedCount:  2,
			expectError:  false,
		},
		{
			name:           "repository error on pending count",
			shouldFailRepo: "CountByStatus",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockMessageRepository()

			// Create messages with different statuses
			for i := 0; i < tt.pendingCount; i++ {
				testMsg := createTestMessage(t)
				mockRepo.messages[testMsg.ID] = testMsg
			}

			for i := 0; i < tt.sentCount; i++ {
				testMsg := createTestMessage(t)
				testMsg.Status = message.StatusSent
				mockRepo.messages[testMsg.ID] = testMsg
			}

			for i := 0; i < tt.failedCount; i++ {
				testMsg := createTestMessage(t)
				testMsg.Status = message.StatusFailed
				mockRepo.messages[testMsg.ID] = testMsg
			}

			if tt.shouldFailRepo != "" {
				mockRepo.shouldFailOp = tt.shouldFailRepo
			}

			service := usecaseImpl.NewMessageProcessingService(mockRepo, &mockWebhookService{}, newMockCacheService())
			ctx := context.Background()

			result, err := service.GetProcessingStatus(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Expected result but got nil")
				return
			}

			if result.PendingCount != int64(tt.pendingCount) {
				t.Errorf("Expected pending count %d, got %d", tt.pendingCount, result.PendingCount)
			}

			if result.ProcessedToday != int64(tt.sentCount) {
				t.Errorf("Expected processed today %d, got %d", tt.sentCount, result.ProcessedToday)
			}

			if result.FailedToday != int64(tt.failedCount) {
				t.Errorf("Expected failed today %d, got %d", tt.failedCount, result.FailedToday)
			}

			if result.NextProcessingAt == nil {
				t.Errorf("Expected NextProcessingAt to be set")
			}

			// Verify repository calls (3 calls to CountByStatus for each status)
			if tt.shouldFailRepo == "" {
				expectedCalls := 3 // pending, sent, failed
				if mockRepo.callCount["CountByStatus"] != expectedCalls {
					t.Errorf("Expected CountByStatus to be called %d times, called %d times",
						expectedCalls, mockRepo.callCount["CountByStatus"])
				}
			}
		})
	}
}

func TestMessageProcessingService_ProcessMessage_Integration(t *testing.T) {
	// This is more of an integration test that verifies the message processing
	// behavior with different message states

	t.Run("processing updates message status correctly", func(t *testing.T) {
		mockRepo := newMockMessageRepository()

		// Create multiple pending messages
		var testMessages []*message.Message
		for i := 0; i < 5; i++ {
			testMsg := createTestMessage(t)
			mockRepo.messages[testMsg.ID] = testMsg
			testMessages = append(testMessages, testMsg)
		}

		service := usecaseImpl.NewMessageProcessingService(mockRepo, &mockWebhookService{}, newMockCacheService())
		ctx := context.Background()

		// Process the messages
		result, err := service.ProcessPendingMessages(ctx, 5)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		if result.ProcessedCount != 5 {
			t.Errorf("Expected 5 messages to be processed, got %d", result.ProcessedCount)
		}

		// Verify that messages were updated (either sent or retry count increased)
		updatedCount := 0
		for _, originalMsg := range testMessages {
			updatedMsg := mockRepo.messages[originalMsg.ID]
			if updatedMsg.Status != message.StatusPending || updatedMsg.RetryCount > 0 {
				updatedCount++
			}
		}

		// All messages should have been processed in some way
		if updatedCount < result.ProcessedCount {
			t.Errorf("Expected at least %d messages to be updated, got %d", result.ProcessedCount, updatedCount)
		}

		// Verify repository calls
		if mockRepo.callCount["Update"] < result.ProcessedCount {
			t.Errorf("Expected at least %d Update calls, got %d", result.ProcessedCount, mockRepo.callCount["Update"])
		}
	})
}
