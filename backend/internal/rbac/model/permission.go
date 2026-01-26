package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission represents a permission in the system
type Permission struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"uniqueIndex;not null;size:100" json:"name"` // e.g., "users.create", "users.read"
	Resource    string     `gorm:"not null;size:50" json:"resource"`           // e.g., "users", "goals", "tasks"
	Action      string     `gorm:"not null;size:50" json:"action"`             // e.g., "create", "read", "update", "delete"
	Description string     `gorm:"size:255" json:"description"`
	IsSystem    bool       `gorm:"default:false" json:"is_system"` // System permissions cannot be deleted
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`

	// Relationships
	Roles []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

// BeforeCreate hook to generate UUID
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Permission) TableName() string {
	return "permissions"
}
