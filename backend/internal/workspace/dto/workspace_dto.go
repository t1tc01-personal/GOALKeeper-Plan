package dto

import "time"

type CreateWorkspaceRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1000"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1000"`
}

type WorkspaceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
