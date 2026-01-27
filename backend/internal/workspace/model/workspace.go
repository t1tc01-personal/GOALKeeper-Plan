package model

import (
	"time"

	"github.com/google/uuid"
)

// Workspace represents a logical workspace that can contain many pages.
type Workspace struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string     `gorm:"type:varchar(255);not null"`
	Description *string    `gorm:"type:text"`
	OwnerID     uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`

	// Pages are managed by page package, referenced by workspace_id foreign key
}
