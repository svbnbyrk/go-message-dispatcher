package usecases

import (
	"context"
	"time"

	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/errors"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/message"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/repositories"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/usecases"
)

// messageProcessingService implements MessageProcessingUseCase
type messageProcessingService struct {
	messageRepo repositories.MessageRepository
}

// NewMessageProcessingService creates a new message processing use case
func NewMessageProcessingService(messageRepo repositories.MessageRepository) usecases.MessageProcessingUseCase {
	return &messageProcessingService{
		messageRepo: messageRepo,
	}
}

// ProcessPendingMessages processes a batch of pending messages
func (s *messageProcessingService) ProcessPendingMessages(ctx context.Context, batchSize int) (*usecases.ProcessingResult, error) {
	// Validate and set default batch size
	if batchSize <= 0 {
		batchSize = 2 // Default batch size
	}
	if batchSize > 10 {
		batchSize = 10 // Maximum batch size for safety
	}

	// Get pending messages
	pendingMessages, err := s.messageRepo.GetPendingMessages(ctx, batchSize)
	if err != nil {
		return nil, err
	}

	result := &usecases.ProcessingResult{
		ProcessedCount: len(pendingMessages),
		SuccessCount:   0,
		FailedCount:    0,
		Errors:         []error{},
	}

	// Process each message
	for _, msg := range pendingMessages {
		if err := s.processMessage(ctx, msg); err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, errors.NewBusinessErrorWithCause(err, "failed to process message %s", msg.ID))
		} else {
			result.SuccessCount++
		}
	}

	return result, nil
}

// processMessage processes a single message (simulated webhook call)
func (s *messageProcessingService) processMessage(ctx context.Context, msg *message.Message) error {
	// TODO: In Phase 4, this will be replaced with actual webhook service call
	// For now, we'll simulate the webhook call

	// Simulate processing time and potential failure
	time.Sleep(100 * time.Millisecond) // Simulate network call

	// Simulate 80% success rate for testing
	if time.Now().UnixNano()%5 == 0 {
		// Simulate failure - handle retry logic
		if err := msg.IncrementRetry(); err != nil {
			// Max retries exceeded, mark as failed
			if markErr := msg.MarkAsFailed(); markErr != nil {
				return errors.NewBusinessError("failed to mark message as failed: %v", markErr)
			}
		}

		// Update message in repository
		if err := s.messageRepo.Update(ctx, msg); err != nil {
			return err
		}

		return errors.NewBusinessError("simulated webhook failure")
	}

	// Simulate successful webhook call
	externalID := "webhook-" + time.Now().Format("20060102150405")
	if err := msg.MarkAsSent(externalID); err != nil {
		return errors.NewBusinessError("failed to mark message as sent: %v", err)
	}

	// Update message in repository
	if err := s.messageRepo.Update(ctx, msg); err != nil {
		return err
	}

	return nil
}

// GetProcessingStatus returns the current processing status
func (s *messageProcessingService) GetProcessingStatus(ctx context.Context) (*usecases.ProcessingStatus, error) {
	// Get pending count
	pendingCount, err := s.messageRepo.CountByStatus(ctx, message.StatusPending)
	if err != nil {
		return nil, err
	}

	// Get sent count
	sentCount, err := s.messageRepo.CountByStatus(ctx, message.StatusSent)
	if err != nil {
		return nil, err
	}

	// Get failed count
	failedCount, err := s.messageRepo.CountByStatus(ctx, message.StatusFailed)
	if err != nil {
		return nil, err
	}

	// Calculate next processing time (every 2 minutes)
	nextProcessing := time.Now().Add(2 * time.Minute)

	return &usecases.ProcessingStatus{
		IsProcessing:     false, // TODO: Track actual processing state
		LastProcessedAt:  nil,   // TODO: Track last processing time
		PendingCount:     pendingCount,
		ProcessedToday:   sentCount,   // Simplified for now
		FailedToday:      failedCount, // Simplified for now
		NextProcessingAt: &nextProcessing,
	}, nil
}
