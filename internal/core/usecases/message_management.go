package usecases

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/errors"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/message"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/repositories"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/services"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/usecases"
)

// messageManagementService implements MessageManagementUseCase
type messageManagementService struct {
	messageRepo  repositories.MessageRepository
	cacheService services.CacheService
}

// NewMessageManagementService creates a new message management use case
func NewMessageManagementService(
	messageRepo repositories.MessageRepository,
	cacheService services.CacheService,
) usecases.MessageManagementUseCase {
	return &messageManagementService{
		messageRepo:  messageRepo,
		cacheService: cacheService,
	}
}

// CreateMessage creates a new message for sending
func (s *messageManagementService) CreateMessage(ctx context.Context, cmd usecases.CreateMessageCommand) (*usecases.MessageResponse, error) {
	// Validate input
	if cmd.PhoneNumber == "" {
		return nil, errors.NewValidationError("phone number is required")
	}
	if cmd.Content == "" {
		return nil, errors.NewValidationError("message content is required")
	}

	// Create domain value objects
	phoneNumber, err := message.NewPhoneNumber(cmd.PhoneNumber)
	if err != nil {
		return nil, errors.NewValidationError("invalid phone number format: %s", cmd.PhoneNumber)
	}

	content, err := message.NewContent(cmd.Content)
	if err != nil {
		return nil, errors.NewValidationError("invalid message content: %v", err)
	}

	// Create message entity
	msg, err := message.NewMessage(phoneNumber, content)
	if err != nil {
		return nil, errors.NewBusinessError("failed to create message: %v", err)
	}

	// Save to repository
	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	return s.messageToResponse(msg), nil
}

// GetMessageByID retrieves a message by its ID
func (s *messageManagementService) GetMessageByID(ctx context.Context, id uuid.UUID) (*usecases.MessageResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("message ID cannot be empty")
	}

	msg, err := s.messageRepo.GetByID(ctx, message.MessageID(id.String()))
	if err != nil {
		return nil, err
	}

	return s.messageToResponse(msg), nil
}

// ListMessages retrieves messages with optional filtering and pagination
func (s *messageManagementService) ListMessages(ctx context.Context, query usecases.ListMessagesQuery) (*usecases.ListMessagesResponse, error) {
	// Validate pagination
	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}

	offset := query.Offset
	if offset < 0 {
		return nil, errors.NewValidationError("offset cannot be negative")
	}

	// For sent messages with small limit (≤10), try cache first for better performance
	if query.Status != nil && *query.Status == message.StatusSent && limit <= 10 && offset == 0 {
		if cachedMessages, err := s.getRecentSentMessagesFromCache(ctx, limit); err == nil && len(cachedMessages) > 0 {
			// Get total count from database
			totalCount, err := s.messageRepo.CountByStatus(ctx, message.StatusSent)
			if err != nil {
				return nil, err
			}

			return &usecases.ListMessagesResponse{
				Messages:   cachedMessages,
				TotalCount: totalCount,
				HasMore:    int64(len(cachedMessages)) < totalCount,
			}, nil
		}
	}

	var messages []*message.Message
	var err error

	// Get messages by status or pending messages
	if query.Status != nil {
		if !query.Status.IsValid() {
			return nil, errors.NewValidationError("invalid message status: %s", *query.Status)
		}

		pagination := repositories.NewPagination(limit, offset)
		messages, err = s.messageRepo.GetByStatus(ctx, *query.Status, pagination)
	} else {
		messages, err = s.messageRepo.GetPendingMessages(ctx, limit)
	}

	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	responses := make([]usecases.MessageResponse, len(messages))
	for i, msg := range messages {
		responses[i] = *s.messageToResponse(msg)
	}

	// Get total count
	var totalCount int64
	if query.Status != nil {
		totalCount, err = s.messageRepo.CountByStatus(ctx, *query.Status)
	} else {
		totalCount, err = s.messageRepo.CountByStatus(ctx, message.StatusPending)
	}

	if err != nil {
		return nil, err
	}

	hasMore := int64(offset+len(messages)) < totalCount

	return &usecases.ListMessagesResponse{
		Messages:   responses,
		TotalCount: totalCount,
		HasMore:    hasMore,
	}, nil
}

// getRecentSentMessagesFromCache retrieves recent sent messages from Redis sorted set
func (s *messageManagementService) getRecentSentMessagesFromCache(ctx context.Context, limit int) ([]usecases.MessageResponse, error) {
	// Get recent message IDs from sorted set (latest first)
	members, err := s.cacheService.ZRevRange(ctx, "sent_messages", 0, int64(limit-1))
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []usecases.MessageResponse{}, nil
	}

	responses := make([]usecases.MessageResponse, 0, len(members))

	// Parse each member and get detailed data
	for _, member := range members {
		// Member format: "messageId:externalId:phoneNumber"
		parts := strings.Split(member, ":")
		if len(parts) != 3 {
			continue // Skip malformed entries
		}

		messageID := parts[0]
		externalID := parts[1]
		phoneNumber := parts[2]

		// Get detailed message data from cache
		var cacheData services.SentMessageCacheData
		cacheKey := "message_detail:" + messageID
		if err := s.cacheService.GetJSON(ctx, cacheKey, &cacheData); err != nil {
			// If detail not found in cache, skip this message
			continue
		}

		// Convert to response
		id, err := uuid.Parse(messageID)
		if err != nil {
			continue // Skip invalid UUIDs
		}

		response := usecases.MessageResponse{
			ID:          id,
			PhoneNumber: phoneNumber,
			Content:     cacheData.Content,
			Status:      string(message.StatusSent),
			ExternalID:  &externalID,
			RetryCount:  0,                                  // Sent messages have 0 retry count
			CreatedAt:   cacheData.SentAt.Add(-time.Minute), // Approximate created time
			UpdatedAt:   cacheData.SentAt,
			SentAt:      &cacheData.SentAt,
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// messageToResponse converts a domain message to response DTO
func (s *messageManagementService) messageToResponse(msg *message.Message) *usecases.MessageResponse {
	id, _ := uuid.Parse(msg.ID.String())
	return &usecases.MessageResponse{
		ID:          id,
		PhoneNumber: msg.PhoneNumber.String(),
		Content:     msg.Content.String(),
		Status:      string(msg.Status),
		ExternalID:  msg.ExternalID,
		RetryCount:  msg.RetryCount,
		CreatedAt:   msg.CreatedAt,
		UpdatedAt:   msg.UpdatedAt,
		SentAt:      msg.SentAt,
	}
}
