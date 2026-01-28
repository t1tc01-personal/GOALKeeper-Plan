package repository

import (
	"context"
	"goalkeeper-plan/internal/block/dto"
	"goalkeeper-plan/internal/block/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BlockRepository interface {
	Create(ctx context.Context, block *model.Block) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Block, error)
	ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*model.Block, error)
	ListByPageIDWithPagination(ctx context.Context, pageID uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Block, *dto.PaginationMeta, error)
	ListByParentBlockID(ctx context.Context, parentBlockID *uuid.UUID) ([]*model.Block, error)
	ListByParentBlockIDWithPagination(ctx context.Context, parentBlockID *uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Block, *dto.PaginationMeta, error)
	Update(ctx context.Context, block *model.Block) error
	Delete(ctx context.Context, id uuid.UUID) error
	Reorder(ctx context.Context, blockIDs []uuid.UUID) error
}

type blockRepository struct {
	db *gorm.DB
}

func NewBlockRepository(db *gorm.DB) BlockRepository {
	return &blockRepository{db: db}
}

func (r *blockRepository) Create(ctx context.Context, block *model.Block) error {
	return r.db.WithContext(ctx).Create(block).Error
}

func (r *blockRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Block, error) {
	var block *model.Block
	err := r.db.WithContext(ctx).
		Preload("BlockType").
		Where("id = ?", id).
		First(&block).Error
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (r *blockRepository) ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*model.Block, error) {
	var blocks []*model.Block
	err := r.db.WithContext(ctx).
		Preload("BlockType").
		Where("page_id = ? AND deleted_at IS NULL", pageID).
		Order("rank ASC, created_at ASC").
		Find(&blocks).Error
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func (r *blockRepository) ListByPageIDWithPagination(ctx context.Context, pageID uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Block, *dto.PaginationMeta, error) {
	var blocks []*model.Block
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).
		Model(&model.Block{}).
		Where("page_id = ? AND deleted_at IS NULL", pageID).
		Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Preload("BlockType").
		Where("page_id = ? AND deleted_at IS NULL", pageID).
		Order("rank ASC, created_at ASC").
		Limit(pagReq.GetLimit()).
		Offset(pagReq.GetOffset()).
		Find(&blocks).Error
	if err != nil {
		return nil, nil, err
	}

	meta := &dto.PaginationMeta{
		Total:   int(total),
		Limit:   pagReq.GetLimit(),
		Offset:  pagReq.GetOffset(),
		HasMore: int64(pagReq.GetOffset()+pagReq.GetLimit()) < total,
	}

	return blocks, meta, nil
}

func (r *blockRepository) ListByParentBlockID(ctx context.Context, parentBlockID *uuid.UUID) ([]*model.Block, error) {
	var blocks []*model.Block
	if parentBlockID == nil {
		// Get root blocks (blocks without a parent)
		err := r.db.WithContext(ctx).
			Where("parent_block_id IS NULL AND deleted_at IS NULL").
			Order("rank ASC, created_at ASC").
			Find(&blocks).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := r.db.WithContext(ctx).
			Where("parent_block_id = ? AND deleted_at IS NULL", parentBlockID).
			Order("rank ASC, created_at ASC").
			Find(&blocks).Error
		if err != nil {
			return nil, err
		}
	}
	return blocks, nil
}

func (r *blockRepository) ListByParentBlockIDWithPagination(ctx context.Context, parentBlockID *uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Block, *dto.PaginationMeta, error) {
	var blocks []*model.Block
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Block{})

	// Build WHERE clause based on parentBlockID
	if parentBlockID == nil {
		query = query.Where("parent_block_id IS NULL AND deleted_at IS NULL")
	} else {
		query = query.Where("parent_block_id = ? AND deleted_at IS NULL", parentBlockID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// Reset query for data retrieval
	query = r.db.WithContext(ctx)
	if parentBlockID == nil {
		query = query.Where("parent_block_id IS NULL AND deleted_at IS NULL")
	} else {
		query = query.Where("parent_block_id = ? AND deleted_at IS NULL", parentBlockID)
	}

	// Get paginated results
	err := query.
		Preload("BlockType").
		Order("rank ASC, created_at ASC").
		Limit(pagReq.GetLimit()).
		Offset(pagReq.GetOffset()).
		Find(&blocks).Error
	if err != nil {
		return nil, nil, err
	}

	meta := &dto.PaginationMeta{
		Total:   int(total),
		Limit:   pagReq.GetLimit(),
		Offset:  pagReq.GetOffset(),
		HasMore: int64(pagReq.GetOffset()+pagReq.GetLimit()) < total,
	}

	return blocks, meta, nil
}

func (r *blockRepository) Update(ctx context.Context, block *model.Block) error {
	return r.db.WithContext(ctx).Save(block).Error
}

func (r *blockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Block{}).Error
}

// Reorder updates the rank of blocks to reflect new ordering (batch operation for performance)
func (r *blockRepository) Reorder(ctx context.Context, blockIDs []uuid.UUID) error {
	// Use a transaction to ensure atomicity
	tx := r.db.WithContext(ctx).Begin()

	for i, blockID := range blockIDs {
		if err := tx.
			Model(&model.Block{}).
			Where("id = ?", blockID).
			Update("rank", int64(i)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
