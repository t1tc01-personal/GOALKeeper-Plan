package model

import (
	"time"

	"github.com/google/uuid"
)

// BlockHistory stores snapshots/versions of a block.
type BlockHistory struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BlockID   uuid.UUID      `gorm:"type:uuid;not null;index"`
	Content   *string        `gorm:"type:text"`
	Metadata  map[string]any `gorm:"type:jsonb;default:'{}'::jsonb"`
	Version   int            `gorm:"column:version_number;not null;index"`
	Reason    *string        `gorm:"column:snapshot_reason;type:varchar(50)"`
	AIContext map[string]any `gorm:"type:jsonb;default:'{}'::jsonb"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	CreatedBy *uuid.UUID     `gorm:"type:uuid;index"`
}

// BlockDelta stores detailed changes for a block.
type BlockDelta struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BlockID   uuid.UUID      `gorm:"type:uuid;not null;index"`
	Delta     map[string]any `gorm:"type:jsonb;not null;default:'{}'::jsonb"`
	ActorType string         `gorm:"type:varchar(20);not null;index"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
}
