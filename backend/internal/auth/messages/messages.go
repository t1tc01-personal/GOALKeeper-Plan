package messages

const (
	// Success messages
	MsgSignUpSuccess     = "Account created successfully"
	MsgLoginSuccess      = "Login successful"
	MsgLogoutSuccess     = "Logged out successfully"
	MsgTokenRefreshed    = "Token refreshed successfully"
	MsgOAuthURLGenerated = "OAuth URL generated successfully"
	MsgOAuthSuccess      = "OAuth authentication successful"

	// Error messages
	MsgInvalidCredentials    = "Invalid email or password"
	MsgTokenExpired          = "Token has expired"
	MsgTokenInvalid          = "Invalid token"
	MsgAuthRequired          = "Authentication required"
	MsgOAuthFailed           = "OAuth authentication failed"
	MsgInvalidOAuthProvider  = "Invalid OAuth provider"
	MsgOAuthProviderRequired = "OAuth provider is required"
	MsgInvalidRefreshToken   = "Invalid refresh token"
	MsgExpiredRefreshToken   = "Refresh token has expired"
	MsgFailedToSignUp        = "Failed to create account"
	MsgFailedToLogin         = "Failed to login"
	MsgFailedToLogout        = "Failed to logout"
	MsgFailedToRefreshToken  = "Failed to refresh token"
	MsgFailedToGetOAuthURL   = "Failed to get OAuth URL"
	MsgFailedOAuthCallback   = "Failed to process OAuth callback"
	MsgEmailAlreadyExists    = "Email is already registered"
	MsgOAuthEmailExists      = "This email is already associated with an account"

	// Validation messages
	MsgPasswordTooShort      = "Password must be at least 8 characters"
	MsgPasswordRequired      = "Password is required"
	MsgOAuthCodeRequired     = "OAuth code is required"
	MsgInvalidPasswordFormat = "Password format is invalid"
)
