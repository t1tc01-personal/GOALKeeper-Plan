package model

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null;index" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`

	// Unique constraint to prevent duplicate assignments
	// This is handled at the database level
}

// TableName specifies the table name
func (UserRole) TableName() string {
	return "user_roles"
}
