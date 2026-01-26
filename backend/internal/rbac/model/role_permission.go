package model

import (
	"time"

	"github.com/google/uuid"
)

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoleID       uuid.UUID `gorm:"type:uuid;not null;index" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null;index" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`

	// Unique constraint to prevent duplicate assignments
	// This is handled at the database level
}

// TableName specifies the table name
func (RolePermission) TableName() string {
	return "role_permissions"
}
