package model

import (
	"time"

	"github.com/google/uuid"
)

// BlockType represents metadata and configuration for a block kind.
type BlockType struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name            string     `gorm:"type:varchar(50);uniqueIndex;not null"`
	Category        string     `gorm:"type:varchar(50);not null;index"`
	DisplayName     string     `gorm:"type:varchar(100);not null"`
	Description     *string    `gorm:"type:text"`
	Icon            *string    `gorm:"type:varchar(50)"`
	DefaultMetadata JSONBMap   `gorm:"type:jsonb;default:'{}'::jsonb"`
	MetadataSchema  JSONBMap   `gorm:"type:jsonb;default:'{}'::jsonb"`
	IsSystem        bool       `gorm:"column:is_system;default:true"`
	IsFramework     bool       `gorm:"column:is_framework_block;default:false;index"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`

	Blocks []Block `gorm:"foreignKey:TypeID"`
}
