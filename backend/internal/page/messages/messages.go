package messages

const (
	// Success messages
	MsgPageCreated      = "Page created successfully"
	MsgPageUpdated      = "Page updated successfully"
	MsgPageDeleted      = "Page deleted successfully"
	MsgPagesFetched     = "Pages retrieved successfully"
	MsgHierarchyFetched = "Page hierarchy retrieved successfully"

	// Error messages
	MsgPageNotFound           = "Page not found"
	MsgInvalidPageID          = "Invalid page ID"
	MsgInvalidWorkspaceID     = "Invalid workspace ID"
	MsgFailedToCreatePage     = "Failed to create page"
	MsgFailedToUpdatePage     = "Failed to update page"
	MsgFailedToDeletePage     = "Failed to delete page"
	MsgFailedToFetchPages     = "Failed to fetch pages"
	MsgFailedToFetchHierarchy = "Failed to fetch page hierarchy"
	MsgWorkspaceIDRequired    = "Workspace ID is required"

	// Validation messages
	MsgPageTitleTooLong = "Page title must be less than 200 characters"
)
