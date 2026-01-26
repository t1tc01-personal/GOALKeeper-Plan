package response

import (
	"goalkeeper-plan/internal/errors"
	"time"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorInfo represents error information in response
type ErrorInfo struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse sends a successful response
func SuccessResponse(ctx *gin.Context, statusCode int, message string, data interface{}) {
	ctx.JSON(statusCode, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(ctx),
	})
}

// ErrorResponse sends an error response
func ErrorResponse(ctx *gin.Context, appErr *errors.AppError) {
	statusCode := appErr.GetHTTPStatusCode()

	ctx.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Type:    string(appErr.Type),
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
		Timestamp: appErr.Timestamp.Unix(),
		RequestID: appErr.RequestID,
	})
}

// getRequestID extracts request ID from context (if using request ID middleware)
func getRequestID(ctx *gin.Context) string {
	if requestID, exists := ctx.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
