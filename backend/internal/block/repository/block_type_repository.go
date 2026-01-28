package repository

import (
	"context"
	"goalkeeper-plan/internal/block/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BlockTypeRepository interface {
	GetByName(ctx context.Context, name string) (*model.BlockType, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.BlockType, error)
}

type blockTypeRepository struct {
	db *gorm.DB
}

func NewBlockTypeRepository(db *gorm.DB) BlockTypeRepository {
	return &blockTypeRepository{db: db}
}

func (r *blockTypeRepository) GetByName(ctx context.Context, name string) (*model.BlockType, error) {
	var blockType model.BlockType
	err := r.db.WithContext(ctx).
		Where("name = ? AND deleted_at IS NULL", name).
		First(&blockType).Error
	if err != nil {
		return nil, err
	}
	return &blockType, nil
}

func (r *blockTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.BlockType, error) {
	var blockType model.BlockType
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&blockType).Error
	if err != nil {
		return nil, err
	}
	return &blockType, nil
}

