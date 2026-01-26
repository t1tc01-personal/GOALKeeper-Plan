package messages

const (
	// Success messages
	MsgUserCreated = "User created successfully"
	MsgUserUpdated = "User updated successfully"
	MsgUserDeleted = "User deleted successfully"
	MsgUsersListed = "Users retrieved successfully"

	// Error messages
	MsgUserNotFound       = "User not found"
	MsgUserExists         = "User already exists"
	MsgInvalidUserID      = "Invalid user ID"
	MsgFailedToCreateUser = "Failed to create user"
	MsgFailedToUpdateUser = "Failed to update user"
	MsgFailedToDeleteUser = "Failed to delete user"
	MsgFailedToGetUser    = "Failed to get user"
	MsgFailedToListUsers  = "Failed to list users"

	// Validation messages
	MsgEmailAlreadyInUse = "Email is already in use"
	MsgInvalidUserName   = "Invalid user name"
	MsgUserNameTooLong   = "User name must be less than 100 characters"
)
