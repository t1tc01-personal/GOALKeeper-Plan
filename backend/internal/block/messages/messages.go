package messages

const (
	// Success messages
	MsgBlockCreated    = "Block created successfully"
	MsgBlockUpdated    = "Block updated successfully"
	MsgBlockDeleted    = "Block deleted successfully"
	MsgBlocksFetched   = "Blocks retrieved successfully"
	MsgBlocksListed    = "Blocks retrieved successfully"
	MsgBlocksReordered = "Blocks reordered successfully"
	MsgBlocksSynced    = "Blocks synced successfully"
	MsgFailedToSyncBlocks = "Failed to sync blocks"

	// Error messages
	MsgBlockNotFound         = "Block not found"
	MsgInvalidBlockID        = "Invalid block ID"
	MsgInvalidBlockType      = "Invalid block type"
	MsgInvalidPageID         = "Invalid page ID"
	MsgFailedToCreateBlock   = "Failed to create block"
	MsgFailedToUpdateBlock   = "Failed to update block"
	MsgFailedToDeleteBlock   = "Failed to delete block"
	MsgFailedToFetchBlocks   = "Failed to fetch blocks"
	MsgFailedToListBlocks    = "Failed to list blocks"
	MsgFailedToReorderBlocks = "Failed to reorder blocks"
	MsgPageIDRequired        = "Page ID is required"

	// Validation messages
	MsgBlockTypeTooLong = "Block type must be at least 1 character"
	MsgInvalidBlockIDs  = "One or more block IDs are invalid"
)
