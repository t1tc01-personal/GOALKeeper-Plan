package api

import (
	"errors"
	"fmt"
	appErrors "goalkeeper-plan/internal/errors"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/messages"
	"goalkeeper-plan/internal/response"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// HandlerFunc represents a handler function that processes the request
type HandlerFunc[T any] func(*gin.Context, T) (interface{}, error)

// HandlerConfig configures the handler behavior
type HandlerConfig struct {
	RequireAuth bool
	Logger      logger.Logger
}

// HandleRequest is a generic helper that handles common request processing:
// - Request binding and validation
// - Error handling
// - Response formatting
func HandleRequest[T any](
	ctx *gin.Context,
	config HandlerConfig,
	handler HandlerFunc[T],
) {
	// Bind request
	var req T
	if err := ctx.ShouldBindJSON(&req); err != nil {
		if config.Logger != nil {
			config.Logger.Error("Failed to bind request", zap.Error(err))
		}

		// Extract detailed validation errors
		validationMsg := extractValidationError(err)
		appErr := appErrors.NewValidationError(
			appErrors.CodeInvalidInput,
			validationMsg,
			err,
		)
		response.ErrorResponse(ctx, appErr)
		return
	}

	// Call handler
	result, err := handler(ctx, req)
	if err != nil {
		handleError(ctx, err, config.Logger)
		return
	}

	// Success response
	response.SuccessResponse(ctx, 200, "", result)
}

// HandleRequestWithStatus is similar to HandleRequest but allows custom status code
func HandleRequestWithStatus[T any](
	ctx *gin.Context,
	config HandlerConfig,
	statusCode int,
	successMessage string,
	handler HandlerFunc[T],
) {
	// Bind request
	var req T
	if err := ctx.ShouldBindJSON(&req); err != nil {
		if config.Logger != nil {
			config.Logger.Error("Failed to bind request", zap.Error(err))
		}

		// Extract detailed validation errors
		validationMsg := extractValidationError(err)
		appErr := appErrors.NewValidationError(
			appErrors.CodeInvalidInput,
			validationMsg,
			err,
		)
		response.ErrorResponse(ctx, appErr)
		return
	}

	// Call handler
	result, err := handler(ctx, req)
	if err != nil {
		handleError(ctx, err, config.Logger)
		return
	}

	// Success response
	response.SuccessResponse(ctx, statusCode, successMessage, result)
}

// HandleQueryRequest handles GET requests with query parameters
type QueryHandlerFunc func(*gin.Context) (interface{}, error)

func HandleQueryRequest(
	ctx *gin.Context,
	config HandlerConfig,
	handler QueryHandlerFunc,
) {
	result, err := handler(ctx)
	if err != nil {
		handleError(ctx, err, config.Logger)
		return
	}

	response.SuccessResponse(ctx, 200, "", result)
}

// HandleQueryRequestWithMessage handles GET requests with custom message
func HandleQueryRequestWithMessage(
	ctx *gin.Context,
	config HandlerConfig,
	successMessage string,
	handler QueryHandlerFunc,
) {
	result, err := handler(ctx)
	if err != nil {
		handleError(ctx, err, config.Logger)
		return
	}

	response.SuccessResponse(ctx, 200, successMessage, result)
}

// HandleParamRequest handles requests that extract data from URL parameters
type ParamHandlerFunc func(*gin.Context, string) (interface{}, error)

func HandleParamRequest(
	ctx *gin.Context,
	paramName string,
	config HandlerConfig,
	handler ParamHandlerFunc,
) {
	param := ctx.Param(paramName)
	if param == "" {
		appErr := appErrors.NewValidationError(
			appErrors.CodeMissingField,
			messages.MsgIDRequired,
			nil,
		)
		response.ErrorResponse(ctx, appErr)
		return
	}

	result, err := handler(ctx, param)
	if err != nil {
		handleError(ctx, err, config.Logger)
		return
	}

	response.SuccessResponse(ctx, 200, "", result)
}

// HandleParamRequestWithMessage handles param requests with custom message
func HandleParamRequestWithMessage(
	ctx *gin.Context,
	paramName string,
	config HandlerConfig,
	successMessage string,
	handler ParamHandlerFunc,
) {
	param := ctx.Param(paramName)
	if param == "" {
		appErr := appErrors.NewValidationError(
			appErrors.CodeMissingField,
			messages.MsgIDRequired,
			nil,
		)
		response.ErrorResponse(ctx, appErr)
		return
	}

	result, err := handler(ctx, param)
	if err != nil {
		handleError(ctx, err, config.Logger)
		return
	}

	response.SuccessResponse(ctx, 200, successMessage, result)
}

// handleError processes errors and returns appropriate response
func handleError(ctx *gin.Context, err error, logger logger.Logger) {
	// Check if it's already an AppError
	if appErr, ok := err.(*appErrors.AppError); ok {
		if logger != nil {
			logger.Error("Handler error", zap.Error(appErr.Err), zap.String("code", appErr.Code))
		}
		response.ErrorResponse(ctx, appErr)
		return
	}

	// Try to extract AppError from wrapped error
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		if logger != nil {
			logger.Error("Handler error", zap.Error(appErr.Err), zap.String("code", appErr.Code))
		}
		response.ErrorResponse(ctx, appErr)
		return
	}

	// Default to internal error
	if logger != nil {
		logger.Error("Handler error", zap.Error(err))
	}
	defaultErr := appErrors.NewInternalError(
		appErrors.CodeInternalError,
		messages.MsgInternalError,
		err,
	)
	response.ErrorResponse(ctx, defaultErr)
}

// extractValidationError extracts detailed validation error messages from Gin binding errors
func extractValidationError(err error) string {
	// Check if it's a validator.ValidationErrors
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		var errorMessages []string
		for _, fieldError := range validationErrors {
			fieldName := fieldError.Field()
			tag := fieldError.Tag()

			var msg string
			switch tag {
			case "required":
				msg = fmt.Sprintf("%s is required", fieldName)
			case "min":
				msg = fmt.Sprintf("%s must be at least %s characters", fieldName, fieldError.Param())
			case "max":
				msg = fmt.Sprintf("%s must be at most %s characters", fieldName, fieldError.Param())
			case "email":
				msg = fmt.Sprintf("%s must be a valid email address", fieldName)
			case "len":
				msg = fmt.Sprintf("%s must be exactly %s characters", fieldName, fieldError.Param())
			default:
				msg = fmt.Sprintf("%s is invalid", fieldName)
			}
			errorMessages = append(errorMessages, msg)
		}
		if len(errorMessages) > 0 {
			return strings.Join(errorMessages, "; ")
		}
	}

	// Check if it's a JSON syntax error
	if strings.Contains(err.Error(), "json:") {
		return fmt.Sprintf("Invalid JSON format: %s", err.Error())
	}

	// Default to generic message
	return messages.MsgInvalidRequestBody
}
