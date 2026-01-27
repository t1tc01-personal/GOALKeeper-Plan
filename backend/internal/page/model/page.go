package model

import (
	"time"

	"github.com/google/uuid"
)

// Page represents a page inside a workspace and supports hierarchy.
type Page struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	WorkspaceID  uuid.UUID  `gorm:"type:uuid;not null;index"`
	ParentPageID *uuid.UUID `gorm:"type:uuid;index"`
	Title        string     `gorm:"type:text;not null"`

	// Planner-related fields
	IsPlanner   bool       `gorm:"column:is_planner;default:false;index"`
	FrameworkID *uuid.UUID `gorm:"type:uuid;index"`

	ViewConfig map[string]any `gorm:"type:jsonb;default:'{}'::jsonb"`

	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`

	Parent *Page `gorm:"foreignKey:ParentPageID"`
	// Blocks are managed by block package, referenced by page_id foreign key
	// Workspace and blocks omitted to avoid circular imports
}
