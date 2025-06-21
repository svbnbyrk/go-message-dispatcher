package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

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
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
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
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	row := r.pool.QueryRow(ctx, query, args...)
	msg, err := r.scanMessage(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("message not found: %s", id.String())
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return msg, nil
}

// GetPendingMessages retrieves pending messages with limit
func (r *MessageRepository) GetPendingMessages(ctx context.Context, limit int) ([]*message.Message, error) {
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
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending messages: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// Update updates an existing message
func (r *MessageRepository) Update(ctx context.Context, msg *message.Message) error {
	query, args, err := r.qb.
		Update("messages").
		Set("phone_number", msg.PhoneNumber.String()).
		Set("content", msg.Content.String()).
		Set("status", msg.Status.String()).
		Set("external_id", msg.ExternalID).
		Set("retry_count", msg.RetryCount).
		Set("updated_at", msg.UpdatedAt).
		Set("sent_at", msg.SentAt).
		Where(squirrel.Eq{"id": msg.ID.String()}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("message not found: %s", msg.ID.String())
	}

	return nil
}

// GetSentMessages retrieves sent messages with pagination
func (r *MessageRepository) GetSentMessages(ctx context.Context, pagination repositories.Pagination) ([]*message.Message, error) {
	return r.GetByStatus(ctx, message.StatusSent, pagination)
}

// GetByStatus retrieves messages by status with pagination
func (r *MessageRepository) GetByStatus(ctx context.Context, status message.Status, pagination repositories.Pagination) ([]*message.Message, error) {
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
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages by status: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// GetByPhoneNumber retrieves messages by phone number with pagination
func (r *MessageRepository) GetByPhoneNumber(ctx context.Context, phoneNumber message.PhoneNumber, pagination repositories.Pagination) ([]*message.Message, error) {
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
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages by phone number: %w", err)
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
		return 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var count int64
	err = r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// DeleteByID deletes a message by ID (for testing purposes)
func (r *MessageRepository) DeleteByID(ctx context.Context, id message.MessageID) error {
	query, args, err := r.qb.
		Delete("messages").
		Where(squirrel.Eq{"id": id.String()}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("message not found: %s", id.String())
	}

	return nil
}

// scanMessage scans a single row into a Message struct
func (r *MessageRepository) scanMessage(row pgx.Row) (*message.Message, error) {
	var (
		id          string
		phoneNumber string
		content     string
		status      string
		externalID  sql.NullString
		retryCount  int
		createdAt   time.Time
		updatedAt   time.Time
		sentAt      sql.NullTime
	)

	err := row.Scan(
		&id, &phoneNumber, &content, &status,
		&externalID, &retryCount, &createdAt, &updatedAt, &sentAt,
	)
	if err != nil {
		return nil, err
	}

	// Convert database values to domain types
	phoneNumberVO, err := message.NewPhoneNumber(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number in database: %w", err)
	}

	contentVO, err := message.NewContent(content)
	if err != nil {
		return nil, fmt.Errorf("invalid content in database: %w", err)
	}

	msg := &message.Message{
		ID:          message.MessageID(id),
		PhoneNumber: phoneNumberVO,
		Content:     contentVO,
		Status:      message.Status(status),
		RetryCount:  retryCount,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	if externalID.Valid {
		msg.ExternalID = &externalID.String
	}

	if sentAt.Valid {
		msg.SentAt = &sentAt.Time
	}

	return msg, nil
}

// scanMessages scans multiple rows into Message structs
func (r *MessageRepository) scanMessages(rows pgx.Rows) ([]*message.Message, error) {
	var messages []*message.Message

	for rows.Next() {
		msg, err := r.scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return messages, nil
}
