package dto

import "time"

type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=50"`
	Description string `json:"description,omitempty" validate:"max=255"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=1,max=50"`
	Description string `json:"description,omitempty" validate:"max=255"`
}

type RoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoleWithPermissionsResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	IsSystem    bool                  `json:"is_system"`
	Permissions []PermissionResponse  `json:"permissions"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

type AssignPermissionToRoleRequest struct {
	PermissionName string `json:"permission_name" validate:"required"`
}

type RemovePermissionFromRoleRequest struct {
	PermissionName string `json:"permission_name" validate:"required"`
}
