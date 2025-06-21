package domain

import (
	"errors"
	"testing"

	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/message"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		content     string
		expectError bool
	}{
		{
			name:        "valid message",
			phoneNumber: "+905551234567",
			content:     "Hello World",
			expectError: false,
		},
		{
			name:        "invalid phone number",
			phoneNumber: "123",
			content:     "Hello World",
			expectError: true,
		},
		{
			name:        "empty content",
			phoneNumber: "+905551234567",
			content:     "",
			expectError: true,
		},
		{
			name:        "content too long",
			phoneNumber: "+905551234567",
			content:     "This is a very long message that definitely exceeds the maximum length limit of one hundred sixty characters which should cause a validation error to be returned and this message continues to get longer and longer to definitely exceed the limit set by the SMS standard",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phoneNumber, phoneErr := message.NewPhoneNumber(tt.phoneNumber)
			content, contentErr := message.NewContent(tt.content)

			// If we expect an error, at least one of the validations should fail
			if tt.expectError {
				if phoneErr == nil && contentErr == nil {
					t.Error("expected error in validation, got none")
				}
				return
			}

			// If we don't expect an error, both validations should succeed
			if !tt.expectError {
				if phoneErr != nil {
					t.Errorf("unexpected error for phone number: %v", phoneErr)
					return
				}
				if contentErr != nil {
					t.Errorf("unexpected error for content: %v", contentErr)
					return
				}

				msg, err := message.NewMessage(phoneNumber, content)
				if err != nil {
					t.Errorf("unexpected error creating message: %v", err)
					return
				}

				if msg.Status != message.StatusPending {
					t.Errorf("expected status to be PENDING, got %s", msg.Status)
				}

				if msg.RetryCount != 0 {
					t.Errorf("expected retry count to be 0, got %d", msg.RetryCount)
				}

				if msg.ID.IsEmpty() {
					t.Error("expected message ID to be generated")
				}
			}
		})
	}
}

func TestMessage_MarkAsSent(t *testing.T) {
	phoneNumber, _ := message.NewPhoneNumber("+905551234567")
	content, _ := message.NewContent("Test message")
	msg, _ := message.NewMessage(phoneNumber, content)

	externalID := "webhook-123"
	err := msg.MarkAsSent(externalID)
	if err != nil {
		t.Errorf("unexpected error marking message as sent: %v", err)
	}

	if msg.Status != message.StatusSent {
		t.Errorf("expected status to be SENT, got %s", msg.Status)
	}

	if msg.ExternalID == nil || *msg.ExternalID != externalID {
		t.Errorf("expected external ID to be %s, got %v", externalID, msg.ExternalID)
	}

	if msg.SentAt == nil {
		t.Error("expected sent_at to be set")
	}
}

func TestMessage_MarkAsFailed(t *testing.T) {
	phoneNumber, _ := message.NewPhoneNumber("+905551234567")
	content, _ := message.NewContent("Test message")
	msg, _ := message.NewMessage(phoneNumber, content)

	err := msg.MarkAsFailed()
	if err != nil {
		t.Errorf("unexpected error marking message as failed: %v", err)
	}

	if msg.Status != message.StatusFailed {
		t.Errorf("expected status to be FAILED, got %s", msg.Status)
	}
}

func TestMessage_IncrementRetry(t *testing.T) {
	phoneNumber, _ := message.NewPhoneNumber("+905551234567")
	content, _ := message.NewContent("Test message")
	msg, _ := message.NewMessage(phoneNumber, content)

	// Test successful retry increment
	err := msg.IncrementRetry()
	if err != nil {
		t.Errorf("unexpected error incrementing retry: %v", err)
	}

	if msg.RetryCount != 1 {
		t.Errorf("expected retry count to be 1, got %d", msg.RetryCount)
	}

	// Test maximum retries exceeded
	for i := 0; i < message.MaxRetryAttempts; i++ {
		msg.IncrementRetry()
	}

	err = msg.IncrementRetry()
	if err == nil {
		t.Error("expected error when exceeding max retries")
	}

	if !errors.As(err, &message.MaxRetriesExceededError{}) {
		t.Errorf("expected MaxRetriesExceededError, got %T", err)
	}
}

func TestMessage_CanRetry(t *testing.T) {
	phoneNumber, _ := message.NewPhoneNumber("+905551234567")
	content, _ := message.NewContent("Test message")
	msg, _ := message.NewMessage(phoneNumber, content)

	// Pending message can retry
	if !msg.CanRetry() {
		t.Error("expected pending message to be retryable")
	}

	// Failed message can retry
	msg.MarkAsFailed()
	if !msg.CanRetry() {
		t.Error("expected failed message to be retryable")
	}

	// Sent message cannot retry
	msg2, _ := message.NewMessage(phoneNumber, content)
	msg2.MarkAsSent("external-123")
	if msg2.CanRetry() {
		t.Error("expected sent message to not be retryable")
	}

	// Message with max retries cannot retry
	msg3, _ := message.NewMessage(phoneNumber, content)
	for i := 0; i < message.MaxRetryAttempts; i++ {
		msg3.IncrementRetry()
	}
	if msg3.CanRetry() {
		t.Error("expected message with max retries to not be retryable")
	}
}

func TestPhoneNumber_Validate(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		expectError bool
	}{
		{"valid international", "+905551234567", false},
		{"valid without plus", "905551234567", false},
		{"valid US number", "+15551234567", false},
		{"too short", "123", true},
		{"too long", "+12345678901234567", true},
		{"empty", "", true},
		{"invalid format", "abc123", true},
		{"starts with zero", "+05551234567", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := message.NewPhoneNumber(tt.phoneNumber)
			hasError := err != nil

			if hasError != tt.expectError {
				t.Errorf("expected error: %v, got error: %v", tt.expectError, hasError)
			}

			if hasError && !errors.As(err, &message.ValidationError{}) {
				t.Errorf("expected ValidationError, got %T", err)
			}
		})
	}
}

func TestContent_Validate(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{"valid short message", "Hello", false},
		{"valid max length", "This message is exactly one hundred and sixty characters long which is the maximum allowed length for SMS messages according to standards", false},
		{"empty content", "", true},
		{"too long", "This message is definitely longer than one hundred and sixty characters which exceeds the maximum allowed length for SMS messages and should definitely cause a validation error to be thrown because it is way too long for any normal SMS message", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := message.NewContent(tt.content)
			hasError := err != nil

			if hasError != tt.expectError {
				t.Errorf("expected error: %v, got error: %v", tt.expectError, hasError)
			}

			if hasError && !errors.As(err, &message.ValidationError{}) {
				t.Errorf("expected ValidationError, got %T", err)
			}
		})
	}
}
