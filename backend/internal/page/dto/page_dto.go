package dto

import "time"

type CreatePageRequest struct {
	WorkspaceID  string  `json:"workspaceId" validate:"required"`
	Title        string  `json:"title" validate:"required,min=1,max=200"`
	ParentPageID *string `json:"parentPageId,omitempty"`
}

type UpdatePageRequest struct {
	Title      string         `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	ViewConfig map[string]any `json:"viewConfig,omitempty"`
}

type PageResponse struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
