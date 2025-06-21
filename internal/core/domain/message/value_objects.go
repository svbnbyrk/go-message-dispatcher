package message

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// PhoneNumber represents a phone number value object
type PhoneNumber string

// Content represents message content value object
type Content string

// Phone validation constants
const (
	minPhoneLength   = 10
	maxPhoneLength   = 15
	maxContentLength = 160 // SMS character limit
)

// Phone number regex pattern (international format)
var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

// NewPhoneNumber creates a new PhoneNumber after validation
func NewPhoneNumber(phone string) (PhoneNumber, error) {
	p := PhoneNumber(phone)
	if err := p.Validate(); err != nil {
		return "", err
	}
	return p, nil
}

// String returns the string representation of PhoneNumber
func (p PhoneNumber) String() string {
	return string(p)
}

// Validate validates the phone number format
func (p PhoneNumber) Validate() error {
	phone := strings.TrimSpace(string(p))

	if phone == "" {
		return NewValidationError("phone number cannot be empty")
	}

	// Remove spaces and common separators for validation
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")

	if len(cleanPhone) < minPhoneLength || len(cleanPhone) > maxPhoneLength {
		return NewValidationError("phone number must be between 10 and 15 digits")
	}

	if !phoneRegex.MatchString(cleanPhone) {
		return NewValidationError("invalid phone number format")
	}

	return nil
}

// IsEmpty checks if the phone number is empty
func (p PhoneNumber) IsEmpty() bool {
	return strings.TrimSpace(string(p)) == ""
}

// NewContent creates a new Content after validation
func NewContent(content string) (Content, error) {
	c := Content(content)
	if err := c.Validate(); err != nil {
		return "", err
	}
	return c, nil
}

// String returns the string representation of Content
func (c Content) String() string {
	return string(c)
}

// Validate validates the message content
func (c Content) Validate() error {
	content := strings.TrimSpace(string(c))

	if content == "" {
		return NewValidationError("message content cannot be empty")
	}

	if utf8.RuneCountInString(content) > maxContentLength {
		return NewValidationError("message content exceeds maximum length of 160 characters")
	}

	return nil
}

// IsEmpty checks if the content is empty
func (c Content) IsEmpty() bool {
	return strings.TrimSpace(string(c)) == ""
}

// Length returns the character count of the content
func (c Content) Length() int {
	return utf8.RuneCountInString(string(c))
}
