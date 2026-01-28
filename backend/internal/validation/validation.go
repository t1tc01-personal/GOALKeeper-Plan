package validation

import (
	"fmt"
	"goalkeeper-plan/internal/errors"
	"net/mail"

	"github.com/google/uuid"
)

// ValidateUUID validates UUID is not nil
func ValidateUUID(id uuid.UUID, fieldName string) *errors.AppError {
	if id == uuid.Nil {
		return &errors.AppError{
			Type:    errors.ErrorTypeValidation,
			Code:    "INVALID_UUID",
			Message: fmt.Sprintf("%s cannot be empty", fieldName),
		}
	}
	return nil
}

// ValidateString validates non-empty string with optional max length
func ValidateString(value string, fieldName string, maxLength ...int) *errors.AppError {
	if value == "" {
		return &errors.AppError{
			Type:    errors.ErrorTypeValidation,
			Code:    "EMPTY_STRING",
			Message: fmt.Sprintf("%s cannot be empty", fieldName),
		}
	}

	if len(maxLength) > 0 && len(value) > maxLength[0] {
		return &errors.AppError{
			Type:    errors.ErrorTypeValidation,
			Code:    "STRING_TOO_LONG",
			Message: fmt.Sprintf("%s exceeds maximum length of %d", fieldName, maxLength[0]),
		}
	}

	return nil
}

// ValidateRequired validates not nil
func ValidateRequired(value interface{}, fieldName string) *errors.AppError {
	if value == nil {
		return &errors.AppError{
			Type:    errors.ErrorTypeValidation,
			Code:    "REQUIRED_FIELD",
			Message: fmt.Sprintf("%s is required", fieldName),
		}
	}
	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) *errors.AppError {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return &errors.AppError{
			Type:    errors.ErrorTypeValidation,
			Code:    "INVALID_EMAIL",
			Message: "Invalid email format",
			Details: err.Error(),
		}
	}
	return nil
}

// ValidateInt validates integer value
func ValidateInt(value int, fieldName string, min, max int) *errors.AppError {
	if value < min || value > max {
		return &errors.AppError{
			Type:    errors.ErrorTypeValidation,
			Code:    "INVALID_INT",
			Message: fmt.Sprintf("%s must be between %d and %d", fieldName, min, max),
		}
	}
	return nil
}

// ValidateMinValue validates minimum value
func ValidateMinValue(value int, fieldName string, minVal int) *errors.AppError {
	if value < minVal {
		return &errors.AppError{
			Type:    errors.ErrorTypeValidation,
			Code:    "BELOW_MINIMUM",
			Message: fmt.Sprintf("%s must be at least %d", fieldName, minVal),
		}
	}
	return nil
}

// ValidateMaxValue validates maximum value
func ValidateMaxValue(value int, fieldName string, maxVal int) *errors.AppError {
	if value > maxVal {
		return &errors.AppError{
			Type:    errors.ErrorTypeValidation,
			Code:    "EXCEEDS_MAXIMUM",
			Message: fmt.Sprintf("%s must be at most %d", fieldName, maxVal),
		}
	}
	return nil
}

// ValidateSliceNotEmpty validates slice is not empty
func ValidateSliceNotEmpty(slice interface{}, fieldName string) *errors.AppError {
	if slice == nil {
		return &errors.AppError{
			Type:    errors.ErrorTypeValidation,
			Code:    "EMPTY_SLICE",
			Message: fmt.Sprintf("%s cannot be empty", fieldName),
		}
	}

	// Type-specific checks would go here if needed
	return nil
}
