package model

import (
	"time"

	"github.com/google/uuid"
)

// SharePermissionRole represents the level of access a user has to a page.
type SharePermissionRole string

const (
	RoleOwner  SharePermissionRole = "owner"
	RoleEditor SharePermissionRole = "editor"
	RoleViewer SharePermissionRole = "viewer"
)

// SharePermission grants a user access to a specific page with a given role.
type SharePermission struct {
	ID        uuid.UUID           `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	PageID    uuid.UUID           `gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID           `gorm:"type:uuid;not null;index"`
	Role      SharePermissionRole `gorm:"type:varchar(20);not null"`
	CreatedAt time.Time           `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time           `gorm:"column:updated_at;autoUpdateTime"`
}
