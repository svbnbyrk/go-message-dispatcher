package dto

// CreateMessageRequest represents request to create a new message
type CreateMessageRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required"`
	Content     string `json:"content" validate:"required"`
}

// ListMessagesRequest represents query parameters for listing messages
type ListMessagesRequest struct {
	Status string `json:"status,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}
