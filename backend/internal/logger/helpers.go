package logger

import (
	"goalkeeper-plan/internal/errors"

	"go.uber.org/zap"
)

// LogServiceError logs a service operation error with context
// This is an extension method on the Logger interface
func LogServiceError(l Logger, operation string, err error, fields ...zap.Field) {
	if err == nil {
		return
	}

	if appErr, ok := err.(*errors.AppError); ok {
		LogAppError(l, appErr, zap.String("operation", operation))
	} else {
		l.Error(
			operation+" failed",
			append(fields, zap.Error(err))...,
		)
	}
}

// LogServiceSuccess logs successful service operation
func LogServiceSuccess(l Logger, operation string, fields ...zap.Field) {
	l.Info(operation+" succeeded", fields...)
}

// LogAppError logs structured AppError
func LogAppError(l Logger, appErr *errors.AppError, fields ...zap.Field) {
	if appErr == nil || l == nil {
		return
	}

	baseFields := []zap.Field{
		zap.String("error_type", string(appErr.Type)),
		zap.String("error_code", appErr.Code),
		zap.String("message", appErr.Message),
	}

	if appErr.Details != "" {
		baseFields = append(baseFields, zap.String("details", appErr.Details))
	}

	if appErr.RequestID != "" {
		baseFields = append(baseFields, zap.String("request_id", appErr.RequestID))
	}

	if appErr.Err != nil {
		baseFields = append(baseFields, zap.Error(appErr.Err))
	}

	// Append additional fields
	baseFields = append(baseFields, fields...)

	l.Error(appErr.Message, baseFields...)
}

// LogRepositoryError logs a repository operation error
func LogRepositoryError(l Logger, operation string, entity string, err error, fields ...zap.Field) {
	if err == nil || l == nil {
		return
	}

	allFields := append(
		fields,
		zap.String("entity", entity),
		zap.Error(err),
	)

	l.Error(operation+" repository error", allFields...)
}

// LogRepositorySuccess logs successful repository operation
func LogRepositorySuccess(l Logger, operation string, entity string, fields ...zap.Field) {
	if l == nil {
		return
	}

	allFields := append(
		fields,
		zap.String("entity", entity),
	)

	l.Debug(operation+" repository success", allFields...)
}

// LogOperationStart logs start of an operation (debug level)
func LogOperationStart(l Logger, operation string, fields ...zap.Field) {
	if l == nil {
		return
	}
	l.Debug("starting "+operation, fields...)
}

// LogOperationComplete logs completion of an operation (debug level)
func LogOperationComplete(l Logger, operation string, fields ...zap.Field) {
	if l == nil {
		return
	}
	l.Debug(operation+" completed", fields...)
}
