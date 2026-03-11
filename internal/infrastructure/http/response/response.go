package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/freelance-engineer/go-service-template-pro/internal/domain"
)

type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId,omitempty"`
	Details   any    `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, status int, apiErr APIError) {
	JSON(w, status, map[string]APIError{"error": apiErr})
}

func FromDomain(err error, requestID string) (int, APIError) {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "Authentication is required", RequestID: requestID}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, APIError{Code: "forbidden", Message: "Not enough permissions", RequestID: requestID}
	case errors.Is(err, domain.ErrRateLimited):
		return http.StatusTooManyRequests, APIError{Code: "rate_limited", Message: "Rate limit exceeded", RequestID: requestID}
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest, APIError{Code: "validation_error", Message: "Request validation failed", RequestID: requestID}
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, APIError{Code: "not_found", Message: "Resource not found", RequestID: requestID}
	default:
		return http.StatusInternalServerError, APIError{Code: "internal_error", Message: "Unexpected server error", RequestID: requestID}
	}
}
