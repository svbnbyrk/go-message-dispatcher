package postgres

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/errors"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/message"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/repositories"
)

// MessageRepository implements the MessageRepository interface using PostgreSQL
type MessageRepository struct {
	pool *pgxpool.Pool
	qb   squirrel.StatementBuilderType
}

// NewMessageRepository creates a new PostgreSQL message repository
func NewMessageRepository(pool *pgxpool.Pool) repositories.MessageRepository {
	return &MessageRepository{
		pool: pool,
		qb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Create creates a new message in the database
func (r *MessageRepository) Create(ctx context.Context, msg *message.Message) error {
	query, args, err := r.qb.
		Insert("messages").
		Columns(
			"id", "phone_number", "content", "status",
			"external_id", "retry_count", "created_at", "updated_at", "sent_at",
		).
		Values(
			msg.ID.String(),
			msg.PhoneNumber.String(),
			msg.Content.String(),
			msg.Status.String(),
			msg.ExternalID,
			msg.RetryCount,
			msg.CreatedAt,
			msg.UpdatedAt,
			msg.SentAt,
		).
		ToSql()

	if err != nil {
		return errors.NewRepositoryError("failed to build insert query: %v", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a message by its ID
func (r *MessageRepository) GetByID(ctx context.Context, id message.MessageID) (*message.Message, error) {
	query, args, err := r.qb.
		Select(
			"id", "phone_number", "content", "status",
			"external_id", "retry_count", "created_at", "updated_at", "sent_at",
		).
		From("messages").
		Where(squirrel.Eq{"id": id.String()}).
		ToSql()

	if err != nil {
		return nil, errors.NewRepositoryError("failed to build select query: %v", err)
	}

	row := r.pool.QueryRow(ctx, query, args...)
	msg, err := r.scanMessage(row)
	if err != nil {
		return nil, errors.MapNotFoundError(err, "message")
	}

	return msg, nil
}

// GetPendingMessages retrieves pending messages with limit
func (r *MessageRepository) GetPendingMessages(ctx context.Context, limit int) ([]*message.Message, error) {
	if limit <= 0 {
		return nil, errors.NewValidationError("limit must be greater than zero")
	}

	query, args, err := r.qb.
		Select(
			"id", "phone_number", "content", "status",
			"external_id", "retry_count", "created_at", "updated_at", "sent_at",
		).
		From("messages").
		Where(squirrel.Eq{"status": message.StatusPending.String()}).
		OrderBy("created_at ASC").
		Limit(uint64(limit)).
		ToSql()

	if err != nil {
		return nil, errors.NewRepositoryError("failed to build pending messages query: %v", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// Update updates an existing message
func (r *MessageRepository) Update(ctx context.Context, msg *message.Message) error {
	updateQuery := r.qb.
		Update("messages").
		Set("phone_number", msg.PhoneNumber.String()).
		Set("content", msg.Content.String()).
		Set("status", msg.Status.String()).
		Set("retry_count", msg.RetryCount).
		Set("updated_at", msg.UpdatedAt).
		Set("sent_at", msg.SentAt).
		Where(squirrel.Eq{"id": msg.ID.String()})

	// Set external_id if message is sent
	if msg.ExternalID != nil {
		updateQuery = updateQuery.Set("external_id", *msg.ExternalID)
	}

	query, args, err := updateQuery.ToSql()
	if err != nil {
		return errors.NewRepositoryError("failed to build update query: %v", err)
	}

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFoundError("message not found for update")
	}

	return nil
}

// GetByStatus retrieves messages by status with pagination
func (r *MessageRepository) GetByStatus(ctx context.Context, status message.Status, pagination repositories.Pagination) ([]*message.Message, error) {
	if pagination.Limit <= 0 {
		return nil, errors.NewValidationError("limit must be greater than zero")
	}
	if pagination.Offset < 0 {
		return nil, errors.NewValidationError("offset cannot be negative")
	}

	query, args, err := r.qb.
		Select(
			"id", "phone_number", "content", "status",
			"external_id", "retry_count", "created_at", "updated_at", "sent_at",
		).
		From("messages").
		Where(squirrel.Eq{"status": status.String()}).
		OrderBy("created_at DESC").
		Limit(uint64(pagination.Limit)).
		Offset(uint64(pagination.Offset)).
		ToSql()

	if err != nil {
		return nil, errors.NewRepositoryError("failed to build status query: %v", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// CountByStatus counts messages by status
func (r *MessageRepository) CountByStatus(ctx context.Context, status message.Status) (int64, error) {
	query, args, err := r.qb.
		Select("COUNT(*)").
		From("messages").
		Where(squirrel.Eq{"status": status.String()}).
		ToSql()

	if err != nil {
		return 0, errors.NewRepositoryError("failed to build count query: %v", err)
	}

	var count int64
	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// DeleteByID deletes a message by its ID
func (r *MessageRepository) DeleteByID(ctx context.Context, id message.MessageID) error {
	query, args, err := r.qb.
		Delete("messages").
		Where(squirrel.Eq{"id": id.String()}).
		ToSql()

	if err != nil {
		return errors.NewRepositoryError("failed to build delete query: %v", err)
	}

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFoundError("message not found for deletion")
	}

	return nil
}

// GetSentMessages retrieves sent messages with pagination
func (r *MessageRepository) GetSentMessages(ctx context.Context, pagination repositories.Pagination) ([]*message.Message, error) {
	if pagination.Limit <= 0 {
		return nil, errors.NewValidationError("limit must be greater than zero")
	}
	if pagination.Offset < 0 {
		return nil, errors.NewValidationError("offset cannot be negative")
	}

	query, args, err := r.qb.
		Select(
			"id", "phone_number", "content", "status",
			"external_id", "retry_count", "created_at", "updated_at", "sent_at",
		).
		From("messages").
		Where(squirrel.Eq{"status": message.StatusSent.String()}).
		OrderBy("sent_at DESC").
		Limit(uint64(pagination.Limit)).
		Offset(uint64(pagination.Offset)).
		ToSql()

	if err != nil {
		return nil, errors.NewRepositoryError("failed to build sent messages query: %v", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// GetByPhoneNumber retrieves messages by phone number with pagination
func (r *MessageRepository) GetByPhoneNumber(ctx context.Context, phoneNumber message.PhoneNumber, pagination repositories.Pagination) ([]*message.Message, error) {
	if pagination.Limit <= 0 {
		return nil, errors.NewValidationError("limit must be greater than zero")
	}
	if pagination.Offset < 0 {
		return nil, errors.NewValidationError("offset cannot be negative")
	}

	query, args, err := r.qb.
		Select(
			"id", "phone_number", "content", "status",
			"external_id", "retry_count", "created_at", "updated_at", "sent_at",
		).
		From("messages").
		Where(squirrel.Eq{"phone_number": phoneNumber.String()}).
		OrderBy("created_at DESC").
		Limit(uint64(pagination.Limit)).
		Offset(uint64(pagination.Offset)).
		ToSql()

	if err != nil {
		return nil, errors.NewRepositoryError("failed to build phone number query: %v", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// scanMessage scans a single row into a message
func (r *MessageRepository) scanMessage(row pgx.Row) (*message.Message, error) {
	var (
		idStr      string
		phoneStr   string
		contentStr string
		statusStr  string
		externalID *string
		retryCount int
		createdAt  time.Time
		updatedAt  time.Time
		sentAt     *time.Time
	)

	err := row.Scan(
		&idStr, &phoneStr, &contentStr, &statusStr,
		&externalID, &retryCount, &createdAt, &updatedAt, &sentAt,
	)
	if err != nil {
		return nil, err
	}

	return r.buildMessage(idStr, phoneStr, contentStr, statusStr, externalID, retryCount, createdAt, updatedAt, sentAt)
}

// scanMessages scans multiple rows into messages
func (r *MessageRepository) scanMessages(rows pgx.Rows) ([]*message.Message, error) {
	var messages []*message.Message

	for rows.Next() {
		var (
			idStr      string
			phoneStr   string
			contentStr string
			statusStr  string
			externalID *string
			retryCount int
			createdAt  time.Time
			updatedAt  time.Time
			sentAt     *time.Time
		)

		err := rows.Scan(
			&idStr, &phoneStr, &contentStr, &statusStr,
			&externalID, &retryCount, &createdAt, &updatedAt, &sentAt,
		)
		if err != nil {
			return nil, err
		}

		msg, err := r.buildMessage(idStr, phoneStr, contentStr, statusStr, externalID, retryCount, createdAt, updatedAt, sentAt)
		if err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// buildMessage builds a message from scanned values
func (r *MessageRepository) buildMessage(
	idStr, phoneStr, contentStr, statusStr string,
	externalID *string,
	retryCount int,
	createdAt, updatedAt time.Time,
	sentAt *time.Time,
) (*message.Message, error) {
	// Create value objects
	phoneNumber, err := message.NewPhoneNumber(phoneStr)
	if err != nil {
		return nil, errors.NewRepositoryError("invalid phone number from database: %v", err)
	}

	content, err := message.NewContent(contentStr)
	if err != nil {
		return nil, errors.NewRepositoryError("invalid content from database: %v", err)
	}

	status := message.Status(statusStr)
	if !status.IsValid() {
		return nil, errors.NewRepositoryError("invalid status from database: %s", statusStr)
	}

	// Create and populate message
	msg := &message.Message{
		ID:          message.MessageID(idStr),
		PhoneNumber: phoneNumber,
		Content:     content,
		Status:      status,
		RetryCount:  retryCount,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		SentAt:      sentAt,
	}

	if externalID != nil {
		msg.ExternalID = externalID
	}

	return msg, nil
}
