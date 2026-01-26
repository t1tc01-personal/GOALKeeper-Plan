package errors

// Common error codes
const (
	// Validation
	CodeInvalidInput  = "INVALID_INPUT"
	CodeMissingField  = "MISSING_FIELD"
	CodeInvalidFormat = "INVALID_FORMAT"
	CodeInvalidID     = "INVALID_ID"

	// Internal
	CodeInternalError = "INTERNAL_ERROR"
	CodeDatabaseError = "DATABASE_ERROR"
)

// User error codes
const (
	CodeUserNotFound       = "USER_NOT_FOUND"
	CodeUserExists         = "USER_EXISTS"
	CodeFailedToCreateUser = "FAILED_TO_CREATE_USER"
	CodeFailedToUpdateUser = "FAILED_TO_UPDATE_USER"
	CodeFailedToDeleteUser = "FAILED_TO_DELETE_USER"
	CodeFailedToGetUser    = "FAILED_TO_GET_USER"
	CodeFailedToListUsers  = "FAILED_TO_LIST_USERS"
)

// Auth error codes
const (
	CodeInvalidCredentials    = "INVALID_CREDENTIALS"
	CodeTokenExpired          = "TOKEN_EXPIRED"
	CodeTokenInvalid          = "TOKEN_INVALID"
	CodeAuthRequired          = "AUTH_REQUIRED"
	CodeOAuthFailed           = "OAUTH_FAILED"
	CodeInvalidOAuthProvider  = "INVALID_OAUTH_PROVIDER"
	CodeOAuthProviderRequired = "OAUTH_PROVIDER_REQUIRED"
	CodeInvalidRefreshToken   = "INVALID_REFRESH_TOKEN"
	CodeExpiredRefreshToken   = "EXPIRED_REFRESH_TOKEN"
	CodeFailedToSignUp        = "FAILED_TO_SIGN_UP"
	CodeFailedToLogin         = "FAILED_TO_LOGIN"
	CodeFailedToLogout        = "FAILED_TO_LOGOUT"
	CodeFailedToRefreshToken  = "FAILED_TO_REFRESH_TOKEN"
	CodeFailedToGetOAuthURL   = "FAILED_TO_GET_OAUTH_URL"
	CodeFailedOAuthCallback   = "FAILED_OAUTH_CALLBACK"
	CodeEmailAlreadyExists    = "EMAIL_ALREADY_EXISTS"
	CodeOAuthEmailExists      = "OAUTH_EMAIL_EXISTS"
)
