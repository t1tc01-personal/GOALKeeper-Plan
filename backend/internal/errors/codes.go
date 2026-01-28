package errors

// Common error codes
const (
	// Validation
	CodeInvalidInput    = "INVALID_INPUT"
	CodeMissingField    = "MISSING_FIELD"
	CodeMissingRequired = "MISSING_REQUIRED"
	CodeInvalidFormat   = "INVALID_FORMAT"
	CodeInvalidID       = "INVALID_ID"

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

// Workspace error codes
const (
	CodeWorkspaceNotFound       = "WORKSPACE_NOT_FOUND"
	CodeFailedToCreateWorkspace = "FAILED_TO_CREATE_WORKSPACE"
	CodeFailedToUpdateWorkspace = "FAILED_TO_UPDATE_WORKSPACE"
	CodeFailedToDeleteWorkspace = "FAILED_TO_DELETE_WORKSPACE"
	CodeFailedToFetchWorkspaces = "FAILED_TO_FETCH_WORKSPACES"
	CodeFailedToListWorkspaces  = "FAILED_TO_LIST_WORKSPACES"
)

// Block error codes
const (
	CodeBlockNotFound         = "BLOCK_NOT_FOUND"
	CodeFailedToCreateBlock   = "FAILED_TO_CREATE_BLOCK"
	CodeFailedToUpdateBlock   = "FAILED_TO_UPDATE_BLOCK"
	CodeFailedToDeleteBlock   = "FAILED_TO_DELETE_BLOCK"
	CodeFailedToFetchBlocks   = "FAILED_TO_FETCH_BLOCKS"
	CodeFailedToListBlocks    = "FAILED_TO_LIST_BLOCKS"
	CodeFailedToReorderBlocks = "FAILED_TO_REORDER_BLOCKS"
)

// Page error codes
const (
	CodePageNotFound           = "PAGE_NOT_FOUND"
	CodeFailedToCreatePage     = "FAILED_TO_CREATE_PAGE"
	CodeFailedToUpdatePage     = "FAILED_TO_UPDATE_PAGE"
	CodeFailedToDeletePage     = "FAILED_TO_DELETE_PAGE"
	CodeFailedToFetchPages     = "FAILED_TO_FETCH_PAGES"
	CodeFailedToFetchHierarchy = "FAILED_TO_FETCH_HIERARCHY"
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

// RBAC error codes
const (
	CodeRoleNotFound             = "ROLE_NOT_FOUND"
	CodeRoleExists               = "ROLE_EXISTS"
	CodeFailedToCreateRole       = "FAILED_TO_CREATE_ROLE"
	CodeFailedToUpdateRole       = "FAILED_TO_UPDATE_ROLE"
	CodeFailedToDeleteRole       = "FAILED_TO_DELETE_ROLE"
	CodeFailedToGetRole          = "FAILED_TO_GET_ROLE"
	CodeFailedToListRoles        = "FAILED_TO_LIST_ROLES"
	CodeSystemRoleCannotDelete   = "SYSTEM_ROLE_CANNOT_DELETE"
	CodeFailedToAssignPermission = "FAILED_TO_ASSIGN_PERMISSION"
	CodeFailedToRemovePermission = "FAILED_TO_REMOVE_PERMISSION"

	CodePermissionNotFound           = "PERMISSION_NOT_FOUND"
	CodePermissionExists             = "PERMISSION_EXISTS"
	CodeFailedToCreatePermission     = "FAILED_TO_CREATE_PERMISSION"
	CodeFailedToUpdatePermission     = "FAILED_TO_UPDATE_PERMISSION"
	CodeFailedToDeletePermission     = "FAILED_TO_DELETE_PERMISSION"
	CodeFailedToGetPermission        = "FAILED_TO_GET_PERMISSION"
	CodeFailedToListPermissions      = "FAILED_TO_LIST_PERMISSIONS"
	CodeSystemPermissionCannotDelete = "SYSTEM_PERMISSION_CANNOT_DELETE"
)
