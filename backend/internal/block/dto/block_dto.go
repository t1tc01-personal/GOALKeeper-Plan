package dto

import "time"

type CreateBlockRequest struct {
	PageID      string         `json:"pageId" validate:"required"`
	Type        string         `json:"type" validate:"required,min=1"`
	Content     string         `json:"content,omitempty"`
	Position    int            `json:"position,omitempty"`
	BlockConfig map[string]any `json:"blockConfig,omitempty"`
}

type UpdateBlockRequest struct {
	Type        string         `json:"type,omitempty" validate:"omitempty,min=1"`
	Content     string         `json:"content,omitempty"`
	Position    int            `json:"position,omitempty"`
	BlockConfig map[string]any `json:"blockConfig,omitempty"`
}

type ReorderBlocksRequest struct {
	BlockIDs []string `json:"blockIds" validate:"required,min=1"`
}

type BlockResponse struct {
	ID          string         `json:"id"`
	PageID      string         `json:"pageId"`
	Type        string         `json:"type"`
	Content     string         `json:"content"`
	Position    int            `json:"position"`
	BlockConfig map[string]any `json:"blockConfig"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Batch sync request DTOs
type BatchCreateItem struct {
	PageID      string         `json:"pageId" validate:"required"`
	Type        string         `json:"type" validate:"required,min=1"`
	Content     string         `json:"content,omitempty"`
	Position    int            `json:"position"`
	BlockConfig map[string]any `json:"blockConfig,omitempty"`
	TempID      string         `json:"tempId" validate:"required"` // Frontend temp ID for mapping
}

type BatchUpdateItem struct {
	ID          string         `json:"id" validate:"required"`
	Content     string         `json:"content,omitempty"`
	Type        string         `json:"type,omitempty" validate:"omitempty,min=1"`
	Position    int            `json:"position,omitempty"`
	BlockConfig map[string]any `json:"blockConfig,omitempty"`
}

type BatchSyncRequest struct {
	Creates []BatchCreateItem `json:"creates,omitempty"`
	Updates []BatchUpdateItem `json:"updates,omitempty"`
	Deletes []string           `json:"deletes,omitempty" validate:"dive,uuid"`
}

type BatchSyncResponse struct {
	Creates []BatchCreateResponse `json:"creates"`
	Updates []BatchUpdateResponse `json:"updates"`
	Deletes []string               `json:"deletes"`
	Errors  []BatchError           `json:"errors,omitempty"`
}

type BatchCreateResponse struct {
	TempID string         `json:"tempId"`
	Block  BlockResponse  `json:"block"`
}

type BatchUpdateResponse struct {
	ID    string        `json:"id"`
	Block BlockResponse `json:"block"`
}

type BatchError struct {
	OperationID string `json:"operationId"`
	Type        string `json:"type"` // "create", "update", "delete"
	Error       string `json:"error"`
}
