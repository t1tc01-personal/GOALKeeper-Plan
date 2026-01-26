package dto

import "time"

type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Resource    string `json:"resource" validate:"required,min=1,max=50"`
	Action      string `json:"action" validate:"required,min=1,max=50"`
	Description string `json:"description,omitempty" validate:"max=255"`
}

type UpdatePermissionRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Resource    string `json:"resource,omitempty" validate:"omitempty,min=1,max=50"`
	Action      string `json:"action,omitempty" validate:"omitempty,min=1,max=50"`
	Description string `json:"description,omitempty" validate:"max=255"`
}

type PermissionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
