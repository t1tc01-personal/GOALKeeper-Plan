package model

import (
	"time"

	"github.com/google/uuid"
)

// Block represents an atomic content unit within a page.
type Block struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	PageID        uuid.UUID `gorm:"type:uuid;not null;index"`
	ParentBlockID *uuid.UUID `gorm:"type:uuid;index"`
	TypeID        uuid.UUID `gorm:"type:uuid;not null;index"`
	Content       *string   `gorm:"type:text"`
	Metadata      JSONBMap  `gorm:"type:jsonb;default:'{}'::jsonb"`
	Rank          int64     `gorm:"column:rank;index"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"`

	Parent    *Block     `gorm:"foreignKey:ParentBlockID"`
	BlockType *BlockType `gorm:"foreignKey:TypeID"`
	History   []BlockHistory
	Deltas    []BlockDelta
	// Page reference omitted to avoid circular imports - referenced by page_id foreign key
}
