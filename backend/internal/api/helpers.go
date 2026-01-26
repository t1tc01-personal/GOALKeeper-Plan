package api

import (
	"errors"
	appErrors "goalkeeper-plan/internal/errors"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/messages"
	"goalkeeper-plan/internal/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BindRequest binds and validates a request, returns error response if failed
func BindRequest[T any](ctx *gin.Context, logger logger.Logger) (T, bool) {
	var req T
	if err := ctx.ShouldBindJSON(&req); err != nil {
		if logger != nil {
			logger.Error("Failed to bind request", zap.Error(err))
		}

		// Extract detailed validation errors
		validationMsg := extractValidationError(err)
		appErr := appErrors.NewValidationError(
			appErrors.CodeInvalidInput,
			validationMsg,
			err,
		)
		response.ErrorResponse(ctx, appErr)
		return req, false
	}
	return req, true
}

// GetParam gets a URL parameter, returns error response if missing
func GetParam(ctx *gin.Context, paramName string, errorMessage string) (string, bool) {
	param := ctx.Param(paramName)
	if param == "" {
		appErr := appErrors.NewValidationError(
			appErrors.CodeMissingField,
			errorMessage,
			nil,
		)
		response.ErrorResponse(ctx, appErr)
		return "", false
	}
	return param, true
}

// GetQuery gets a query parameter with default value
func GetQuery(ctx *gin.Context, key string, defaultValue string) string {
	value := ctx.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// HandleError processes an error and sends appropriate response
func HandleError(ctx *gin.Context, err error, logger logger.Logger) {
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

// SendSuccess sends a success response
func SendSuccess(ctx *gin.Context, statusCode int, message string, data interface{}) {
	response.SuccessResponse(ctx, statusCode, message, data)
}

// SendError sends an error response
func SendError(ctx *gin.Context, appErr *appErrors.AppError) {
	response.ErrorResponse(ctx, appErr)
}
