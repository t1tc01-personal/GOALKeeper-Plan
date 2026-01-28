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
