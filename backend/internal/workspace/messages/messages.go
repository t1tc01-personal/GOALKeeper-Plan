package messages

const (
	// Success messages
	MsgWorkspaceCreated = "Workspace created successfully"
	MsgWorkspaceUpdated = "Workspace updated successfully"
	MsgWorkspaceDeleted = "Workspace deleted successfully"
	MsgWorkspacesListed = "Workspaces retrieved successfully"

	// Error messages
	MsgWorkspaceNotFound       = "Workspace not found"
	MsgInvalidWorkspaceID      = "Invalid workspace ID"
	MsgFailedToCreateWorkspace = "Failed to create workspace"
	MsgFailedToUpdateWorkspace = "Failed to update workspace"
	MsgFailedToDeleteWorkspace = "Failed to delete workspace"
	MsgFailedToListWorkspaces  = "Failed to list workspaces"
	MsgFailedToFetchWorkspaces = "Failed to fetch workspaces"
	MsgWorkspaceNameRequired   = "Workspace name is required"

	// Validation messages
	MsgWorkspaceNameTooLong = "Workspace name must be less than 200 characters"
	MsgWorkspaceDescTooLong = "Workspace description must be less than 1000 characters"
)
