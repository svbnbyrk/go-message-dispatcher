package usecases

import (
	"context"
	"time"

	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/errors"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/message"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/repositories"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/services"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/usecases"
)

// messageProcessingService implements MessageProcessingUseCase
type messageProcessingService struct {
	messageRepo    repositories.MessageRepository
	webhookService services.WebhookService
	cacheService   services.CacheService
}

// NewMessageProcessingService creates a new message processing use case
func NewMessageProcessingService(
	messageRepo repositories.MessageRepository,
	webhookService services.WebhookService,
	cacheService services.CacheService,
) usecases.MessageProcessingUseCase {
	return &messageProcessingService{
		messageRepo:    messageRepo,
		webhookService: webhookService,
		cacheService:   cacheService,
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

// processMessage processes a single message using webhook service
func (s *messageProcessingService) processMessage(ctx context.Context, msg *message.Message) error {
	// Prepare webhook request
	webhookReq := services.WebhookRequest{
		PhoneNumber: msg.PhoneNumber.String(),
		Content:     msg.Content.String(),
		MessageID:   msg.ID.String(),
	}

	// Send message via webhook
	webhookResp, err := s.webhookService.SendMessage(ctx, webhookReq)
	if err != nil {
		// Handle webhook failure - increment retry count
		if retryErr := msg.IncrementRetry(); retryErr != nil {
			// Max retries exceeded, mark as failed
			if markErr := msg.MarkAsFailed(); markErr != nil {
				return errors.NewBusinessError("failed to mark message as failed: %v", markErr)
			}
		}

		// Update message in repository
		if updateErr := s.messageRepo.Update(ctx, msg); updateErr != nil {
			return updateErr
		}

		return errors.NewBusinessErrorWithCause(err, "webhook call failed for message %s", msg.ID)
	}

	// Webhook success - mark message as sent
	if err := msg.MarkAsSent(webhookResp.ExternalID); err != nil {
		return errors.NewBusinessError("failed to mark message as sent: %v", err)
	}

	// Update message in repository
	if err := s.messageRepo.Update(ctx, msg); err != nil {
		return err
	}

	// Cache the sent message information
	cacheData := map[string]interface{}{
		"message_id":  msg.ID.String(),
		"external_id": webhookResp.ExternalID,
		"sent_at":     msg.SentAt,
		"status":      string(msg.Status),
	}

	cacheKey := "message:" + msg.ID.String()
	if err := s.cacheService.SetJSON(ctx, cacheKey, cacheData, 30*24*time.Hour); err != nil {
		// Cache failure shouldn't break the flow, just log it
		// In a real application, you might want to use a logger here
		_ = err // Ignore cache errors for now
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
