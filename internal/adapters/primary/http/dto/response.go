package dto

import "time"

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID          string     `json:"id"`
	PhoneNumber string     `json:"phoneNumber"`
	Content     string     `json:"content"`
	Status      string     `json:"status"`
	ExternalID  *string    `json:"externalId,omitempty"`
	RetryCount  int        `json:"retryCount"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	SentAt      *time.Time `json:"sentAt,omitempty"`
}

// ProcessingStatusResponse represents the current processing status
type ProcessingStatusResponse struct {
	IsProcessing     bool       `json:"isProcessing"`
	LastProcessedAt  *time.Time `json:"lastProcessedAt,omitempty"`
	PendingCount     int64      `json:"pendingCount"`
	ProcessedToday   int64      `json:"processedToday"`
	FailedToday      int64      `json:"failedToday"`
	NextProcessingAt *time.Time `json:"nextProcessingAt,omitempty"`
}

// ProcessingResultResponse represents the result of manual message processing
type ProcessingResultResponse struct {
	ProcessedCount int      `json:"processedCount"`
	SuccessCount   int      `json:"successCount"`
	FailedCount    int      `json:"failedCount"`
	Errors         []string `json:"errors,omitempty"`
}

// SchedulerStatusResponse represents the background scheduler status
type SchedulerStatusResponse struct {
	IsRunning             bool      `json:"isRunning"`
	IsCurrentlyProcessing bool      `json:"isCurrentlyProcessing"`
	TotalProcessed        int64     `json:"totalProcessed"`
	TotalSuccessful       int64     `json:"totalSuccessful"`
	TotalFailed           int64     `json:"totalFailed"`
	LastProcessingTime    time.Time `json:"lastProcessingTime"`
	NextProcessingIn      string    `json:"nextProcessingIn"`
	Interval              string    `json:"interval"`
	BatchSize             int       `json:"batchSize"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
