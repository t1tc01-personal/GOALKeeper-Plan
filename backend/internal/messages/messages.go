package messages

// Common request/response messages
const (
	// Request validation
	MsgInvalidRequestBody = "Invalid request body"
	MsgMissingField       = "Required field is missing"
	MsgInvalidFormat      = "Invalid data format"
	MsgInvalidID          = "Invalid ID format"

	// Generic success/error
	MsgSuccess            = "Operation completed successfully"
	MsgInternalError      = "An internal error occurred"
	MsgDatabaseError      = "Database operation failed"
	MsgServiceUnavailable = "Service temporarily unavailable"

	// Common validation
	MsgInvalidEmail  = "Invalid email format"
	MsgEmailRequired = "Email is required"
	MsgNameRequired  = "Name is required"
	MsgIDRequired    = "ID is required"
)
