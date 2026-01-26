package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a user role in the system
type Role struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"uniqueIndex;not null;size:50" json:"name"`
	Description string     `gorm:"size:255" json:"description"`
	IsSystem    bool       `gorm:"default:false" json:"is_system"` // System roles cannot be deleted
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`

	// Relationships
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	// Users relationship is defined in the User model (many2many:user_roles)
}

// BeforeCreate hook to generate UUID
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Role) TableName() string {
	return "roles"
}
