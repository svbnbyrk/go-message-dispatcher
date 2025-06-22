package dto

// CreateMessageRequest represents request to create a new message
type CreateMessageRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required" example:"+905551234567" doc:"Phone number in international format (E.164)"`
	Content     string `json:"content" validate:"required" example:"Hello World! This is a test message." doc:"Message content to be sent"`
}

// ListMessagesRequest represents query parameters for listing messages
type ListMessagesRequest struct {
	Status string `json:"status,omitempty" example:"pending" enums:"pending,sent,failed" doc:"Filter messages by status"`
	Limit  int    `json:"limit,omitempty" example:"10" minimum:"1" maximum:"100" doc:"Number of messages to return (1-100)"`
	Offset int    `json:"offset,omitempty" example:"0" minimum:"0" doc:"Number of messages to skip for pagination"`
}
