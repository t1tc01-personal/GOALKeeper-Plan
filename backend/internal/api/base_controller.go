package api

import (
	"goalkeeper-plan/internal/errors"
	"goalkeeper-plan/internal/logger"

	"github.com/gin-gonic/gin"
)

// BaseController provides common functionality for all controllers
type BaseController struct {
	Logger logger.Logger
}

// RespondError responds with structured error
// Automatically converts to AppError if needed and selects appropriate HTTP status code
func (bc *BaseController) RespondError(ctx *gin.Context, err error) {
	if err == nil {
		err = &errors.AppError{
			Type:    errors.ErrorTypeInternal,
			Code:    "UNKNOWN_ERROR",
			Message: "An unknown error occurred",
		}
	}

	var appErr *errors.AppError

	// Convert to AppError if needed
	if e, ok := err.(*errors.AppError); ok {
		appErr = e
	} else {
		appErr = &errors.AppError{
			Type:    errors.ErrorTypeInternal,
			Code:    "INTERNAL_ERROR",
			Message: "An unexpected error occurred",
			Err:     err,
		}
	}

	// Log the error
	if bc.Logger != nil {
		logger.LogAppError(bc.Logger, appErr)
	}

	// Respond with error and appropriate HTTP status code
	ctx.JSON(appErr.GetHTTPStatusCode(), appErr)
}

// RespondSuccess responds with data and status code
func (bc *BaseController) RespondSuccess(ctx *gin.Context, statusCode int, data interface{}) {
	ctx.JSON(statusCode, data)
}

// RespondCreated is shortcut for 201 Created
func (bc *BaseController) RespondCreated(ctx *gin.Context, data interface{}) {
	bc.RespondSuccess(ctx, 201, data)
}

// RespondOK is shortcut for 200 OK
func (bc *BaseController) RespondOK(ctx *gin.Context, data interface{}) {
	bc.RespondSuccess(ctx, 200, data)
}

// RespondNoContent is shortcut for 204 No Content
func (bc *BaseController) RespondNoContent(ctx *gin.Context) {
	ctx.Status(204)
}

// RespondBadRequest is shortcut for 400 Bad Request
func (bc *BaseController) RespondBadRequest(ctx *gin.Context, message string) {
	bc.RespondError(ctx, &errors.AppError{
		Type:    errors.ErrorTypeValidation,
		Code:    "BAD_REQUEST",
		Message: message,
	})
}

// RespondNotFound is shortcut for 404 Not Found
func (bc *BaseController) RespondNotFound(ctx *gin.Context, message string) {
	bc.RespondError(ctx, &errors.AppError{
		Type:    errors.ErrorTypeNotFound,
		Code:    "NOT_FOUND",
		Message: message,
	})
}

// RespondForbidden is shortcut for 403 Forbidden
func (bc *BaseController) RespondForbidden(ctx *gin.Context, message string) {
	bc.RespondError(ctx, &errors.AppError{
		Type:    errors.ErrorTypeAuthorization,
		Code:    "FORBIDDEN",
		Message: message,
	})
}

// RespondUnauthorized is shortcut for 401 Unauthorized
func (bc *BaseController) RespondUnauthorized(ctx *gin.Context, message string) {
	bc.RespondError(ctx, &errors.AppError{
		Type:    errors.ErrorTypeAuthentication,
		Code:    "UNAUTHORIZED",
		Message: message,
	})
}

// RespondConflict is shortcut for 409 Conflict
func (bc *BaseController) RespondConflict(ctx *gin.Context, message string) {
	bc.RespondError(ctx, &errors.AppError{
		Type:    errors.ErrorTypeConflict,
		Code:    "CONFLICT",
		Message: message,
	})
}

// RespondInternalError is shortcut for 500 Internal Server Error
func (bc *BaseController) RespondInternalError(ctx *gin.Context, err error) {
	appErr := &errors.AppError{
		Type:    errors.ErrorTypeInternal,
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "An internal server error occurred",
		Err:     err,
	}
	bc.RespondError(ctx, appErr)
}

// Package-level convenience functions for controllers that don't embed BaseController

var defaultController = &BaseController{}

// RespondError responds with a structured error
func RespondError(ctx *gin.Context, err error) {
	defaultController.RespondError(ctx, err)
}

// RespondSuccess responds with data and status code
func RespondSuccess(ctx *gin.Context, statusCode int, data interface{}) {
	defaultController.RespondSuccess(ctx, statusCode, data)
}

// RespondCreated is a shortcut for 201 Created
func RespondCreated(ctx *gin.Context, data interface{}) {
	defaultController.RespondCreated(ctx, data)
}

// RespondOK is a shortcut for 200 OK
func RespondOK(ctx *gin.Context, data interface{}) {
	defaultController.RespondOK(ctx, data)
}

// RespondNoContent is a shortcut for 204 No Content
func RespondNoContent(ctx *gin.Context) {
	defaultController.RespondNoContent(ctx)
}

// RespondBadRequest is a shortcut for 400 Bad Request
func RespondBadRequest(ctx *gin.Context, message string) {
	defaultController.RespondBadRequest(ctx, message)
}

// RespondNotFound is a shortcut for 404 Not Found
func RespondNotFound(ctx *gin.Context, message string) {
	defaultController.RespondNotFound(ctx, message)
}

// RespondForbidden is a shortcut for 403 Forbidden
func RespondForbidden(ctx *gin.Context, message string) {
	defaultController.RespondForbidden(ctx, message)
}

// RespondUnauthorized is a shortcut for 401 Unauthorized
func RespondUnauthorized(ctx *gin.Context, message string) {
	defaultController.RespondUnauthorized(ctx, message)
}

// RespondConflict is a shortcut for 409 Conflict
func RespondConflict(ctx *gin.Context, message string) {
	defaultController.RespondConflict(ctx, message)
}

// RespondInternalError is a shortcut for 500 Internal Server Error
func RespondInternalError(ctx *gin.Context, err error) {
	defaultController.RespondInternalError(ctx, err)
}
