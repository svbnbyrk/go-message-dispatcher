package dto

import "time"

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID          string     `json:"id" example:"123e4567-e89b-12d3-a456-426614174000" doc:"Unique message identifier"`
	PhoneNumber string     `json:"phoneNumber" example:"+905551234567" doc:"Phone number in international format"`
	Content     string     `json:"content" example:"Hello World! This is a test message." doc:"Message content"`
	Status      string     `json:"status" example:"pending" enums:"pending,sent,failed" doc:"Current message status"`
	ExternalID  *string    `json:"externalId,omitempty" example:"whatsapp_msg_12345" doc:"External service message ID (set when sent)"`
	RetryCount  int        `json:"retryCount" example:"0" minimum:"0" maximum:"3" doc:"Number of retry attempts"`
	CreatedAt   time.Time  `json:"createdAt" example:"2024-01-15T10:30:00Z" doc:"Message creation timestamp"`
	UpdatedAt   time.Time  `json:"updatedAt" example:"2024-01-15T10:30:00Z" doc:"Last update timestamp"`
	SentAt      *time.Time `json:"sentAt,omitempty" example:"2024-01-15T10:32:00Z" doc:"Message sent timestamp (if sent)"`
}

// ProcessingStatusResponse represents the current processing status
type ProcessingStatusResponse struct {
	IsProcessing     bool       `json:"isProcessing" example:"false" doc:"Whether system is currently processing messages"`
	LastProcessedAt  *time.Time `json:"lastProcessedAt,omitempty" example:"2024-01-15T10:30:00Z" doc:"Last processing timestamp"`
	PendingCount     int64      `json:"pendingCount" example:"5" doc:"Number of pending messages"`
	ProcessedToday   int64      `json:"processedToday" example:"25" doc:"Messages processed today"`
	FailedToday      int64      `json:"failedToday" example:"2" doc:"Messages failed today"`
	NextProcessingAt *time.Time `json:"nextProcessingAt,omitempty" example:"2024-01-15T10:32:00Z" doc:"Next scheduled processing time"`
}

// ProcessingResultResponse represents the result of manual message processing
type ProcessingResultResponse struct {
	ProcessedCount int      `json:"processedCount" example:"5" doc:"Total number of messages processed"`
	SuccessCount   int      `json:"successCount" example:"4" doc:"Number of successfully sent messages"`
	FailedCount    int      `json:"failedCount" example:"1" doc:"Number of failed messages"`
	Errors         []string `json:"errors,omitempty" example:"[\"webhook timeout for message 123\"]" doc:"List of error messages"`
}

// SchedulerStatusResponse represents the background scheduler status
type SchedulerStatusResponse struct {
	IsRunning             bool      `json:"isRunning" example:"true" doc:"Whether scheduler is running"`
	IsCurrentlyProcessing bool      `json:"isCurrentlyProcessing" example:"false" doc:"Whether currently processing a batch"`
	TotalProcessed        int64     `json:"totalProcessed" example:"150" doc:"Total messages processed since start"`
	TotalSuccessful       int64     `json:"totalSuccessful" example:"140" doc:"Total successful messages"`
	TotalFailed           int64     `json:"totalFailed" example:"10" doc:"Total failed messages"`
	LastProcessingTime    time.Time `json:"lastProcessingTime" example:"2024-01-15T10:30:00Z" doc:"Last processing timestamp"`
	NextProcessingIn      string    `json:"nextProcessingIn" example:"1m30s" doc:"Time until next processing"`
	Interval              string    `json:"interval" example:"2m" doc:"Processing interval"`
	BatchSize             int       `json:"batchSize" example:"2" doc:"Number of messages processed per batch"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Validation failed" doc:"Error type or category"`
	Message string `json:"message,omitempty" example:"phoneNumber is required" doc:"Detailed error message"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string `json:"status" example:"ok" enums:"ok,error" doc:"Health status"`
	Uptime  string `json:"uptime" example:"2h30m45s" doc:"Server uptime"`
	Version string `json:"version" example:"1.0.0" doc:"Application version"`
}
