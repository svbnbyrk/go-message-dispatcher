package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/dto"
	domainErrors "github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/errors"
)

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Fallback error response
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal Server Error","message":"failed to encode response"}`))
	}
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := dto.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	json.NewEncoder(w).Encode(errorResp)
}

// handleUseCaseError handles errors from use cases and maps them to HTTP responses
func handleUseCaseError(w http.ResponseWriter, err error) {
	var validationErr domainErrors.ValidationError
	var notFoundErr domainErrors.NotFoundError
	var businessErr domainErrors.BusinessError

	switch {
	case errors.As(err, &validationErr):
		writeJSONError(w, http.StatusBadRequest, err.Error())
	case errors.As(err, &notFoundErr):
		writeJSONError(w, http.StatusNotFound, err.Error())
	case errors.As(err, &businessErr):
		writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
	}
}
