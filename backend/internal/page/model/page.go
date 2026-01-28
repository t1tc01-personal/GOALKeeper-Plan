package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JSONBMap is a custom type for handling JSONB fields in PostgreSQL
type JSONBMap map[string]any

// Value implements the driver.Valuer interface for JSONBMap
func (j JSONBMap) Value() (driver.Value, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONBMap
func (j *JSONBMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONBMap)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		bytes = []byte("{}")
	}

	if len(bytes) == 0 {
		*j = make(JSONBMap)
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// Page represents a page inside a workspace and supports hierarchy.
type Page struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	WorkspaceID  uuid.UUID  `gorm:"type:uuid;not null;index"`
	ParentPageID *uuid.UUID `gorm:"type:uuid;index"`
	Title        string     `gorm:"type:text;not null"`

	// Planner-related fields
	IsPlanner   bool       `gorm:"column:is_planner;default:false;index"`
	FrameworkID *uuid.UUID `gorm:"type:uuid;index"`

	ViewConfig JSONBMap `gorm:"type:jsonb;default:'{}'::jsonb"`

	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`

	Parent *Page `gorm:"foreignKey:ParentPageID"`
	// Blocks are managed by block package, referenced by page_id foreign key
	// Workspace and blocks omitted to avoid circular imports
}
