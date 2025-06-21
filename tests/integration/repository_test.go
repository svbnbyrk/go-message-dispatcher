package integration

import (
	"context"
	"testing"
	"time"

	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/secondary/repositories/postgres"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/message"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/repositories"
)

// Note: These tests require a running PostgreSQL database with proper setup
//
// SETUP: Use the automated script
// ./deployments/scripts/setup-test-database.sh
//
// This script will:
// - Start PostgreSQL container
// - Create database and user
// - Apply migrations
// - Configure everything for testing
//
// Cleanup: Stop and remove container
// docker stop msg-dispatcher-test-db && docker rm msg-dispatcher-test-db

func setupTestDB(t *testing.T) repositories.MessageRepository {
	t.Helper()

	config := postgres.DatabaseConfig{
		Host:              "localhost",
		Port:              5432,
		Username:          "msg_dispatcher_user",
		Password:          "msg_test_pass123",
		Database:          "message_dispatcher_test",
		SSLMode:           "disable",
		MaxConnections:    5,
		MinConnections:    1,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   time.Minute * 30,
		HealthCheckPeriod: time.Minute,
	}

	ctx := context.Background()
	pool, err := postgres.NewConnectionPool(ctx, config)
	if err != nil {
		t.Skipf("Skipping integration tests: failed to connect to database: %v", err)
	}

	repo := postgres.NewMessageRepository(pool)

	// Cleanup function
	t.Cleanup(func() {
		postgres.CloseConnectionPool(pool)
	})

	return repo
}

func createTestMessage(t *testing.T) *message.Message {
	t.Helper()

	phoneNumber, err := message.NewPhoneNumber("+905551234567")
	if err != nil {
		t.Fatalf("Failed to create phone number: %v", err)
	}

	content, err := message.NewContent("Test message for integration testing")
	if err != nil {
		t.Fatalf("Failed to create content: %v", err)
	}

	msg, err := message.NewMessage(phoneNumber, content)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	return msg
}

func TestMessageRepository_CreateAndGetByID(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Create test message
	msg := createTestMessage(t)

	// Test Create
	err := repo.Create(ctx, msg)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Test GetByID
	retrievedMsg, err := repo.GetByID(ctx, msg.ID)
	if err != nil {
		t.Fatalf("Failed to get message by ID: %v", err)
	}

	// Verify the message
	if retrievedMsg.ID != msg.ID {
		t.Errorf("Expected ID %s, got %s", msg.ID, retrievedMsg.ID)
	}

	if retrievedMsg.PhoneNumber.String() != msg.PhoneNumber.String() {
		t.Errorf("Expected phone number %s, got %s", msg.PhoneNumber.String(), retrievedMsg.PhoneNumber.String())
	}

	if retrievedMsg.Content.String() != msg.Content.String() {
		t.Errorf("Expected content %s, got %s", msg.Content.String(), retrievedMsg.Content.String())
	}

	if retrievedMsg.Status != message.StatusPending {
		t.Errorf("Expected status %s, got %s", message.StatusPending, retrievedMsg.Status)
	}

	// Cleanup
	repo.DeleteByID(ctx, msg.ID)
}

func TestMessageRepository_Update(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Create test message
	msg := createTestMessage(t)
	err := repo.Create(ctx, msg)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Update the message
	externalID := "webhook-response-123"
	err = msg.MarkAsSent(externalID)
	if err != nil {
		t.Fatalf("Failed to mark message as sent: %v", err)
	}

	// Update in repository
	err = repo.Update(ctx, msg)
	if err != nil {
		t.Fatalf("Failed to update message: %v", err)
	}

	// Retrieve and verify
	updatedMsg, err := repo.GetByID(ctx, msg.ID)
	if err != nil {
		t.Fatalf("Failed to get updated message: %v", err)
	}

	if updatedMsg.Status != message.StatusSent {
		t.Errorf("Expected status %s, got %s", message.StatusSent, updatedMsg.Status)
	}

	if updatedMsg.ExternalID == nil || *updatedMsg.ExternalID != externalID {
		t.Errorf("Expected external ID %s, got %v", externalID, updatedMsg.ExternalID)
	}

	if updatedMsg.SentAt == nil {
		t.Error("Expected sent_at to be set")
	}

	// Cleanup
	repo.DeleteByID(ctx, msg.ID)
}

func TestMessageRepository_GetPendingMessages(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Create multiple test messages
	var testMessages []*message.Message
	for i := 0; i < 3; i++ {
		msg := createTestMessage(t)
		err := repo.Create(ctx, msg)
		if err != nil {
			t.Fatalf("Failed to create test message %d: %v", i, err)
		}
		testMessages = append(testMessages, msg)
	}

	// Get pending messages
	pendingMessages, err := repo.GetPendingMessages(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to get pending messages: %v", err)
	}

	// Verify we have at least our test messages
	if len(pendingMessages) < 3 {
		t.Errorf("Expected at least 3 pending messages, got %d", len(pendingMessages))
	}

	// Verify all are pending
	for _, msg := range pendingMessages {
		if msg.Status != message.StatusPending {
			t.Errorf("Expected all messages to be pending, found status: %s", msg.Status)
		}
	}

	// Cleanup
	for _, msg := range testMessages {
		repo.DeleteByID(ctx, msg.ID)
	}
}

func TestMessageRepository_GetByStatus(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Create and mark a message as sent
	msg := createTestMessage(t)
	err := repo.Create(ctx, msg)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	err = msg.MarkAsSent("external-123")
	if err != nil {
		t.Fatalf("Failed to mark message as sent: %v", err)
	}

	err = repo.Update(ctx, msg)
	if err != nil {
		t.Fatalf("Failed to update message: %v", err)
	}

	// Get sent messages
	pagination := repositories.DefaultPagination()
	sentMessages, err := repo.GetByStatus(ctx, message.StatusSent, pagination)
	if err != nil {
		t.Fatalf("Failed to get sent messages: %v", err)
	}

	// Verify we have at least one sent message
	found := false
	for _, sentMsg := range sentMessages {
		if sentMsg.ID == msg.ID {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find our test message in sent messages")
	}

	// Cleanup
	repo.DeleteByID(ctx, msg.ID)
}

func TestMessageRepository_CountByStatus(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Get initial count
	initialCount, err := repo.CountByStatus(ctx, message.StatusPending)
	if err != nil {
		t.Fatalf("Failed to count pending messages: %v", err)
	}

	// Create test message
	msg := createTestMessage(t)
	err = repo.Create(ctx, msg)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Get new count
	newCount, err := repo.CountByStatus(ctx, message.StatusPending)
	if err != nil {
		t.Fatalf("Failed to count pending messages after create: %v", err)
	}

	// Verify count increased
	if newCount != initialCount+1 {
		t.Errorf("Expected count to increase by 1, initial: %d, new: %d", initialCount, newCount)
	}

	// Cleanup
	repo.DeleteByID(ctx, msg.ID)
}
