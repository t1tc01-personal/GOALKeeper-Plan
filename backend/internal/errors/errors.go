package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeAuthorization  ErrorType = "authorization"
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeConflict       ErrorType = "conflict"
	ErrorTypeInternal       ErrorType = "internal"
)

// AppError represents a structured application error
type AppError struct {
	Type      ErrorType `json:"type"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
	Err       error     `json:"-"` // Original error, not exposed to client
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s:%s] %s - %v", e.Type, e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// GetHTTPStatusCode returns appropriate HTTP status code
func (e *AppError) GetHTTPStatusCode() int {
	switch e.Type {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeAuthentication:
		return http.StatusUnauthorized
	case ErrorTypeAuthorization:
		return http.StatusForbidden
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithRequestID adds a request ID to the error
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// Error constructors
func NewValidationError(code, message string, err error) *AppError {
	return &AppError{
		Type:      ErrorTypeValidation,
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Err:       err,
	}
}

func NewAuthenticationError(code, message string, err error) *AppError {
	return &AppError{
		Type:      ErrorTypeAuthentication,
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Err:       err,
	}
}

func NewAuthorizationError(code, message string, err error) *AppError {
	return &AppError{
		Type:      ErrorTypeAuthorization,
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Err:       err,
	}
}

func NewNotFoundError(code, message string, err error) *AppError {
	return &AppError{
		Type:      ErrorTypeNotFound,
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Err:       err,
	}
}

func NewConflictError(code, message string, err error) *AppError {
	return &AppError{
		Type:      ErrorTypeConflict,
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Err:       err,
	}
}

func NewInternalError(code, message string, err error) *AppError {
	return &AppError{
		Type:      ErrorTypeInternal,
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Err:       err,
	}
}

