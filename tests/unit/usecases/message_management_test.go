package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/message"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/repositories"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/usecases"
	usecaseImpl "github.com/svbnbyrk/go-message-dispatcher/internal/core/usecases"
)

// Mock repository for testing
type mockMessageRepository struct {
	messages     map[message.MessageID]*message.Message
	shouldFailOp string
	callCount    map[string]int
}

func newMockMessageRepository() *mockMessageRepository {
	return &mockMessageRepository{
		messages:  make(map[message.MessageID]*message.Message),
		callCount: make(map[string]int),
	}
}

func (m *mockMessageRepository) Create(ctx context.Context, msg *message.Message) error {
	m.callCount["Create"]++
	if m.shouldFailOp == "Create" {
		return errors.New("mock create error")
	}
	m.messages[msg.ID] = msg
	return nil
}

func (m *mockMessageRepository) GetByID(ctx context.Context, id message.MessageID) (*message.Message, error) {
	m.callCount["GetByID"]++
	if m.shouldFailOp == "GetByID" {
		return nil, errors.New("mock get error")
	}
	msg, exists := m.messages[id]
	if !exists {
		return nil, errors.New("message not found")
	}
	return msg, nil
}

func (m *mockMessageRepository) GetPendingMessages(ctx context.Context, limit int) ([]*message.Message, error) {
	m.callCount["GetPendingMessages"]++
	if m.shouldFailOp == "GetPendingMessages" {
		return nil, errors.New("mock get pending error")
	}

	var pending []*message.Message
	count := 0
	for _, msg := range m.messages {
		if msg.Status == message.StatusPending && count < limit {
			pending = append(pending, msg)
			count++
		}
	}
	return pending, nil
}

func (m *mockMessageRepository) Update(ctx context.Context, msg *message.Message) error {
	m.callCount["Update"]++
	if m.shouldFailOp == "Update" {
		return errors.New("mock update error")
	}
	m.messages[msg.ID] = msg
	return nil
}

func (m *mockMessageRepository) GetSentMessages(ctx context.Context, pagination repositories.Pagination) ([]*message.Message, error) {
	return nil, nil // Not needed for these tests
}

func (m *mockMessageRepository) GetByStatus(ctx context.Context, status message.Status, pagination repositories.Pagination) ([]*message.Message, error) {
	m.callCount["GetByStatus"]++
	if m.shouldFailOp == "GetByStatus" {
		return nil, errors.New("mock get by status error")
	}

	var result []*message.Message
	count := 0
	for _, msg := range m.messages {
		if msg.Status == status && count < pagination.Limit {
			if count >= pagination.Offset {
				result = append(result, msg)
			}
			count++
		}
	}
	return result, nil
}

func (m *mockMessageRepository) GetByPhoneNumber(ctx context.Context, phoneNumber message.PhoneNumber, pagination repositories.Pagination) ([]*message.Message, error) {
	return nil, nil // Not needed for these tests
}

func (m *mockMessageRepository) CountByStatus(ctx context.Context, status message.Status) (int64, error) {
	m.callCount["CountByStatus"]++
	if m.shouldFailOp == "CountByStatus" {
		return 0, errors.New("mock count error")
	}

	count := int64(0)
	for _, msg := range m.messages {
		if msg.Status == status {
			count++
		}
	}
	return count, nil
}

func (m *mockMessageRepository) DeleteByID(ctx context.Context, id message.MessageID) error {
	delete(m.messages, id)
	return nil
}

// Test helper to create a test message
func createTestMessage(t *testing.T) *message.Message {
	t.Helper()

	phoneNumber, err := message.NewPhoneNumber("+905551234567")
	if err != nil {
		t.Fatalf("Failed to create phone number: %v", err)
	}

	content, err := message.NewContent("Test message")
	if err != nil {
		t.Fatalf("Failed to create content: %v", err)
	}

	msg, err := message.NewMessage(phoneNumber, content)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	return msg
}

func TestMessageManagementService_CreateMessage(t *testing.T) {
	tests := []struct {
		name           string
		command        usecases.CreateMessageCommand
		shouldFailRepo bool
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful creation",
			command: usecases.CreateMessageCommand{
				PhoneNumber: "+905551234567",
				Content:     "Test message",
			},
			shouldFailRepo: false,
			expectError:    false,
		},
		{
			name: "empty phone number",
			command: usecases.CreateMessageCommand{
				PhoneNumber: "",
				Content:     "Test message",
			},
			shouldFailRepo: false,
			expectError:    true,
			errorContains:  "phone number is required",
		},
		{
			name: "empty content",
			command: usecases.CreateMessageCommand{
				PhoneNumber: "+905551234567",
				Content:     "",
			},
			shouldFailRepo: false,
			expectError:    true,
			errorContains:  "content is required",
		},
		{
			name: "invalid phone number",
			command: usecases.CreateMessageCommand{
				PhoneNumber: "invalid",
				Content:     "Test message",
			},
			shouldFailRepo: false,
			expectError:    true,
			errorContains:  "invalid phone number",
		},
		{
			name: "repository error",
			command: usecases.CreateMessageCommand{
				PhoneNumber: "+905551234567",
				Content:     "Test message",
			},
			shouldFailRepo: true,
			expectError:    true,
			errorContains:  "mock create error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockMessageRepository()
			if tt.shouldFailRepo {
				mockRepo.shouldFailOp = "Create"
			}

			service := usecaseImpl.NewMessageManagementService(mockRepo)
			ctx := context.Background()

			result, err := service.CreateMessage(ctx, tt.command)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				if result != nil {
					t.Errorf("Expected nil result on error, got: %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("Expected result but got nil")
				} else {
					if result.Status != string(message.StatusPending) {
						t.Errorf("Expected status %s, got %s", message.StatusPending, result.Status)
					}
					if result.PhoneNumber != tt.command.PhoneNumber {
						t.Errorf("Expected phone number %s, got %s", tt.command.PhoneNumber, result.PhoneNumber)
					}
					if result.Content != tt.command.Content {
						t.Errorf("Expected content %s, got %s", tt.command.Content, result.Content)
					}
				}
			}
		})
	}
}

func TestMessageManagementService_GetMessageByID(t *testing.T) {
	mockRepo := newMockMessageRepository()
	service := usecaseImpl.NewMessageManagementService(mockRepo)
	ctx := context.Background()

	// Create a test message
	testMsg := createTestMessage(t)
	mockRepo.messages[testMsg.ID] = testMsg

	t.Run("successful retrieval", func(t *testing.T) {
		id, _ := uuid.Parse(testMsg.ID.String())
		result, err := service.GetMessageByID(ctx, id)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Errorf("Expected result but got nil")
		} else {
			if result.ID != id {
				t.Errorf("Expected ID %s, got %s", id, result.ID)
			}
		}
	})

	t.Run("message not found", func(t *testing.T) {
		id := uuid.New()
		result, err := service.GetMessageByID(ctx, id)

		if err == nil {
			t.Errorf("Expected error but got none")
		}
		if result != nil {
			t.Errorf("Expected nil result but got: %v", result)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.shouldFailOp = "GetByID"
		id, _ := uuid.Parse(testMsg.ID.String())
		result, err := service.GetMessageByID(ctx, id)

		if err == nil {
			t.Errorf("Expected error but got none")
		}
		if result != nil {
			t.Errorf("Expected nil result but got: %v", result)
		}
	})
}

func TestMessageManagementService_ListMessages(t *testing.T) {
	mockRepo := newMockMessageRepository()
	service := usecaseImpl.NewMessageManagementService(mockRepo)
	ctx := context.Background()

	// Create test messages
	for i := 0; i < 5; i++ {
		testMsg := createTestMessage(t)
		mockRepo.messages[testMsg.ID] = testMsg
	}

	t.Run("successful listing with status filter", func(t *testing.T) {
		status := message.StatusPending
		query := usecases.ListMessagesQuery{
			Status: &status,
			Limit:  10,
			Offset: 0,
		}

		result, err := service.ListMessages(ctx, query)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Errorf("Expected result but got nil")
		} else {
			if len(result.Messages) != 5 {
				t.Errorf("Expected 5 messages, got %d", len(result.Messages))
			}
			if result.TotalCount != 5 {
				t.Errorf("Expected total count 5, got %d", result.TotalCount)
			}
		}
	})

	t.Run("pagination defaults", func(t *testing.T) {
		query := usecases.ListMessagesQuery{
			Limit:  0, // Should default to 20
			Offset: 0, // Valid offset
		}

		result, err := service.ListMessages(ctx, query)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Errorf("Expected result but got nil")
		}
	})

	t.Run("negative offset validation", func(t *testing.T) {
		query := usecases.ListMessagesQuery{
			Limit:  10,
			Offset: -1, // Invalid offset
		}

		result, err := service.ListMessages(ctx, query)

		if err == nil {
			t.Errorf("Expected error for negative offset but got none")
		}
		if result != nil {
			t.Errorf("Expected nil result on error, got: %v", result)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
