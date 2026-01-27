package service

import (
	"context"
	"fmt"
	"goalkeeper-plan/internal/block/dto"
	"goalkeeper-plan/internal/block/model"
	"goalkeeper-plan/internal/block/repository"
	"goalkeeper-plan/internal/cache"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/validation"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BlockService interface {
	CreateBlock(ctx context.Context, pageID uuid.UUID, blockType *model.BlockType, content *string, position int64) (*model.Block, error)
	GetBlock(ctx context.Context, id uuid.UUID) (*model.Block, error)
	ListBlocksByPage(ctx context.Context, pageID uuid.UUID) ([]*model.Block, error)
	ListBlocksByPageWithPagination(ctx context.Context, pageID uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Block, *dto.PaginationMeta, error)
	ListBlocksByParent(ctx context.Context, parentBlockID *uuid.UUID) ([]*model.Block, error)
	ListBlocksByParentWithPagination(ctx context.Context, parentBlockID *uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Block, *dto.PaginationMeta, error)
	UpdateBlock(ctx context.Context, id uuid.UUID, content *string) (*model.Block, error)
	DeleteBlock(ctx context.Context, id uuid.UUID) error
	ReorderBlocks(ctx context.Context, pageID uuid.UUID, blockIDs []uuid.UUID) error
}

type blockService struct {
	repo      repository.BlockRepository
	cacheServ cache.CacheService
	logger    logger.Logger
}

type BlockServiceOption func(*blockService)

func WithBlockRepository(r repository.BlockRepository) BlockServiceOption {
	return func(s *blockService) {
		s.repo = r
	}
}

func WithBlockCacheService(c cache.CacheService) BlockServiceOption {
	return func(s *blockService) {
		s.cacheServ = c
	}
}

func WithBlockLogger(log logger.Logger) BlockServiceOption {
	return func(s *blockService) {
		s.logger = log
	}
}

func NewBlockService(opts ...BlockServiceOption) (BlockService, error) {
	s := &blockService{}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

func (s *blockService) CreateBlock(ctx context.Context, pageID uuid.UUID, blockType *model.BlockType, content *string, position int64) (*model.Block, error) {
	// Validate inputs
	if err := validation.ValidateUUID(pageID, "page_id"); err != nil {
		return nil, err
	}

	if blockType == nil {
		return nil, validation.ValidateRequired(blockType, "block_type")
	}

	block := &model.Block{
		PageID:   pageID,
		TypeID:   blockType.ID,
		Content:  content,
		Rank:     position,
		Metadata: make(map[string]any),
	}

	if err := s.repo.Create(ctx, block); err != nil {
		logger.LogServiceError(s.logger, "create_block", err, zap.String("pageID", pageID.String()), zap.String("typeID", blockType.ID.String()))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "create_block", zap.String("id", block.ID.String()))
	return block, nil
}

func (s *blockService) GetBlock(ctx context.Context, id uuid.UUID) (*model.Block, error) {
	if err := validation.ValidateUUID(id, "block_id"); err != nil {
		return nil, err
	}

	block, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.LogServiceError(s.logger, "get_block", err, zap.String("id", id.String()))
		return nil, err
	}

	return block, nil
}

func (s *blockService) ListBlocksByPage(ctx context.Context, pageID uuid.UUID) ([]*model.Block, error) {
	if err := validation.ValidateUUID(pageID, "page_id"); err != nil {
		return nil, err
	}

	blocks, err := s.repo.ListByPageID(ctx, pageID)
	if err != nil {
		logger.LogServiceError(s.logger, "list_blocks_by_page", err, zap.String("pageID", pageID.String()))
		return nil, err
	}

	return blocks, nil
}

func (s *blockService) ListBlocksByPageWithPagination(ctx context.Context, pageID uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Block, *dto.PaginationMeta, error) {
	if err := validation.ValidateUUID(pageID, "page_id"); err != nil {
		return nil, nil, err
	}

	blocks, meta, err := s.repo.ListByPageIDWithPagination(ctx, pageID, pagReq)
	if err != nil {
		logger.LogServiceError(s.logger, "list_blocks_by_page_paginated", err, zap.String("pageID", pageID.String()))
		return nil, nil, err
	}

	return blocks, meta, nil
}

func (s *blockService) ListBlocksByParent(ctx context.Context, parentBlockID *uuid.UUID) ([]*model.Block, error) {
	blocks, err := s.repo.ListByParentBlockID(ctx, parentBlockID)
	if err != nil {
		if parentBlockID != nil {
			logger.LogServiceError(s.logger, "list_blocks_by_parent", err, zap.String("parentBlockID", parentBlockID.String()))
		} else {
			logger.LogServiceError(s.logger, "list_root_blocks", err)
		}
		return nil, err
	}

	return blocks, nil
}

func (s *blockService) ListBlocksByParentWithPagination(ctx context.Context, parentBlockID *uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Block, *dto.PaginationMeta, error) {
	blocks, meta, err := s.repo.ListByParentBlockIDWithPagination(ctx, parentBlockID, pagReq)
	if err != nil {
		if parentBlockID != nil {
			logger.LogServiceError(s.logger, "list_blocks_by_parent_paginated", err, zap.String("parentBlockID", parentBlockID.String()))
		} else {
			logger.LogServiceError(s.logger, "list_root_blocks_paginated", err)
		}
		return nil, nil, err
	}

	return blocks, meta, nil
}

func (s *blockService) UpdateBlock(ctx context.Context, id uuid.UUID, content *string) (*model.Block, error) {
	if err := validation.ValidateUUID(id, "block_id"); err != nil {
		return nil, err
	}

	block, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.LogServiceError(s.logger, "update_block_fetch", err, zap.String("id", id.String()))
		return nil, err
	}

	// Last-write-wins: Simply update the content with provided value
	if content != nil {
		block.Content = content
	}

	if err := s.repo.Update(ctx, block); err != nil {
		logger.LogServiceError(s.logger, "update_block_save", err, zap.String("id", id.String()))
		return nil, err
	}

	// Invalidate cache entries for this block and its page
	if s.cacheServ != nil {
		blockCacheKey := fmt.Sprintf(cache.BlockByIDPattern, id.String())
		pageCacheKey := fmt.Sprintf(cache.BlocksByPagePattern, block.PageID.String())
		_ = s.cacheServ.Delete(ctx, blockCacheKey, pageCacheKey)
	}

	logger.LogServiceSuccess(s.logger, "update_block", zap.String("id", id.String()))
	return block, nil
}

func (s *blockService) DeleteBlock(ctx context.Context, id uuid.UUID) error {
	if err := validation.ValidateUUID(id, "block_id"); err != nil {
		return err
	}

	// Get block before deletion to clear related caches
	block, err := s.repo.GetByID(ctx, id)
	if err == nil && block != nil && s.cacheServ != nil {
		blockCacheKey := fmt.Sprintf(cache.BlockByIDPattern, id.String())
		pageCacheKey := fmt.Sprintf(cache.BlocksByPagePattern, block.PageID.String())
		_ = s.cacheServ.Delete(ctx, blockCacheKey, pageCacheKey)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.LogServiceError(s.logger, "delete_block", err, zap.String("id", id.String()))
		return err
	}

	logger.LogServiceSuccess(s.logger, "delete_block", zap.String("id", id.String()))
	return nil
}

func (s *blockService) ReorderBlocks(ctx context.Context, pageID uuid.UUID, blockIDs []uuid.UUID) error {
	if err := validation.ValidateUUID(pageID, "page_id"); err != nil {
		return err
	}

	if err := validation.ValidateSliceNotEmpty(blockIDs, "block_ids"); err != nil {
		return err
	}

	if err := s.repo.Reorder(ctx, blockIDs); err != nil {
		logger.LogServiceError(s.logger, "reorder_blocks", err, zap.String("pageID", pageID.String()), zap.Int("count", len(blockIDs)))
		return err
	}

	// Invalidate cache for this page
	if s.cacheServ != nil {
		pageCacheKey := fmt.Sprintf(cache.BlocksByPagePattern, pageID.String())
		_ = s.cacheServ.Delete(ctx, pageCacheKey)
	}

	logger.LogServiceSuccess(s.logger, "reorder_blocks", zap.String("pageID", pageID.String()), zap.Int("count", len(blockIDs)))
	return nil
}
