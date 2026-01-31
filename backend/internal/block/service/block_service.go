package service

import (
	"context"
	"errors"
	"fmt"
	"goalkeeper-plan/internal/block/dto"
	"goalkeeper-plan/internal/block/messages"
	"goalkeeper-plan/internal/block/model"
	"goalkeeper-plan/internal/block/repository"
	"goalkeeper-plan/internal/cache"
	appErrors "goalkeeper-plan/internal/errors"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/validation"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BlockService interface {
	CreateBlock(ctx context.Context, pageID uuid.UUID, blockType *model.BlockType, content *string, position int64, parentBlockID *uuid.UUID) (*model.Block, error)
	CreateBlockWithMetadata(ctx context.Context, pageID uuid.UUID, blockType *model.BlockType, content *string, metadata model.JSONBMap, position int64, parentBlockID *uuid.UUID) (*model.Block, error)
	GetBlockTypeByName(ctx context.Context, name string) (*model.BlockType, error)
	GetBlock(ctx context.Context, id uuid.UUID) (*model.Block, error)
	ListBlocksByPage(ctx context.Context, pageID uuid.UUID) ([]*model.Block, error)
	ListBlocksByPageWithPagination(ctx context.Context, pageID uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Block, *dto.PaginationMeta, error)
	ListBlocksByParent(ctx context.Context, parentBlockID *uuid.UUID) ([]*model.Block, error)
	ListBlocksByParentWithPagination(ctx context.Context, parentBlockID *uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Block, *dto.PaginationMeta, error)
	UpdateBlock(ctx context.Context, id uuid.UUID, content *string) (*model.Block, error)
	UpdateBlockWithMetadata(ctx context.Context, id uuid.UUID, content *string, metadata model.JSONBMap) (*model.Block, error)
	DeleteBlock(ctx context.Context, id uuid.UUID) error
	ReorderBlocks(ctx context.Context, pageID uuid.UUID, blockIDs []uuid.UUID) error
	BatchSync(ctx context.Context, req *dto.BatchSyncRequest) (*dto.BatchSyncResponse, error)
}

type blockService struct {
	repo          repository.BlockRepository
	blockTypeRepo repository.BlockTypeRepository
	cacheServ     cache.CacheService
	logger        logger.Logger
}

type BlockServiceOption func(*blockService)

func WithBlockRepository(r repository.BlockRepository) BlockServiceOption {
	return func(s *blockService) {
		s.repo = r
	}
}

func WithBlockTypeRepository(r repository.BlockTypeRepository) BlockServiceOption {
	return func(s *blockService) {
		s.blockTypeRepo = r
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

func (s *blockService) CreateBlock(ctx context.Context, pageID uuid.UUID, blockType *model.BlockType, content *string, position int64, parentBlockID *uuid.UUID) (*model.Block, error) {
	// Validate inputs
	if err := validation.ValidateUUID(pageID, "page_id"); err != nil {
		return nil, err
	}

	if blockType == nil {
		return nil, validation.ValidateRequired(blockType, "block_type")
	}

	block := &model.Block{
		PageID:        pageID,
		TypeID:        blockType.ID,
		Content:       content,
		Rank:          position,
		ParentBlockID: parentBlockID,
		Metadata:      make(model.JSONBMap),
	}

	if err := s.repo.Create(ctx, block); err != nil {
		logger.LogServiceError(s.logger, "create_block", err, zap.String("pageID", pageID.String()), zap.String("typeID", blockType.ID.String()))
		return nil, err
	}

	// Manually attach BlockType so it's available in the return value
	block.BlockType = blockType

	logger.LogServiceSuccess(s.logger, "create_block", zap.String("id", block.ID.String()))
	return block, nil
}

func (s *blockService) CreateBlockWithMetadata(ctx context.Context, pageID uuid.UUID, blockType *model.BlockType, content *string, metadata model.JSONBMap, position int64, parentBlockID *uuid.UUID) (*model.Block, error) {
	if err := validation.ValidateUUID(pageID, "page_id"); err != nil {
		return nil, err
	}

	if blockType == nil {
		return nil, validation.ValidateRequired(blockType, "block_type")
	}

	if metadata == nil {
		metadata = make(model.JSONBMap)
	}

	block := &model.Block{
		PageID:        pageID,
		TypeID:        blockType.ID,
		Content:       content,
		Rank:          position,
		ParentBlockID: parentBlockID,
		Metadata:      metadata,
	}

	if err := s.repo.Create(ctx, block); err != nil {
		logger.LogServiceError(s.logger, "create_block_with_metadata", err, zap.String("pageID", pageID.String()), zap.String("typeID", blockType.ID.String()))
		return nil, err
	}

	block.BlockType = blockType

	logger.LogServiceSuccess(s.logger, "create_block_with_metadata", zap.String("id", block.ID.String()))
	return block, nil
}

func (s *blockService) GetBlockTypeByName(ctx context.Context, name string) (*model.BlockType, error) {
	if s.blockTypeRepo == nil {
		return nil, fmt.Errorf("block type repository not initialized")
	}
	return s.blockTypeRepo.GetByName(ctx, name)
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
		// Map DB not-found to domain not-found error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := appErrors.NewNotFoundError(
				appErrors.CodeBlockNotFound,
				messages.MsgBlockNotFound,
				err,
			)
			logger.LogServiceError(s.logger, "update_block_not_found", appErr, zap.String("id", id.String()))
			return nil, appErr
		}

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

func (s *blockService) UpdateBlockWithMetadata(ctx context.Context, id uuid.UUID, content *string, metadata model.JSONBMap) (*model.Block, error) {
	if err := validation.ValidateUUID(id, "block_id"); err != nil {
		return nil, err
	}

	block, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := appErrors.NewNotFoundError(
				appErrors.CodeBlockNotFound,
				messages.MsgBlockNotFound,
				err,
			)
			logger.LogServiceError(s.logger, "update_block_with_metadata_not_found", appErr, zap.String("id", id.String()))
			return nil, appErr
		}

		logger.LogServiceError(s.logger, "update_block_with_metadata_fetch", err, zap.String("id", id.String()))
		return nil, err
	}

	if content != nil {
		block.Content = content
	}

	if metadata != nil {
		block.Metadata = metadata
	}

	if err := s.repo.Update(ctx, block); err != nil {
		logger.LogServiceError(s.logger, "update_block_with_metadata_save", err, zap.String("id", id.String()))
		return nil, err
	}

	if s.cacheServ != nil {
		blockCacheKey := fmt.Sprintf(cache.BlockByIDPattern, id.String())
		pageCacheKey := fmt.Sprintf(cache.BlocksByPagePattern, block.PageID.String())
		_ = s.cacheServ.Delete(ctx, blockCacheKey, pageCacheKey)
	}

	logger.LogServiceSuccess(s.logger, "update_block_with_metadata", zap.String("id", id.String()))
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

func (s *blockService) BatchSync(ctx context.Context, req *dto.BatchSyncRequest) (*dto.BatchSyncResponse, error) {
	response := &dto.BatchSyncResponse{
		Creates: []dto.BatchCreateResponse{},
		Updates: []dto.BatchUpdateResponse{},
		Deletes: []string{},
		Errors:  []dto.BatchError{},
	}

	// Process creates
	for _, createItem := range req.Creates {
		pageID, err := uuid.Parse(createItem.PageID)
		if err != nil {
			response.Errors = append(response.Errors, dto.BatchError{
				OperationID: createItem.TempID,
				Type:        "create",
				Error:       fmt.Sprintf("invalid page ID: %v", err),
			})
			continue
		}

		blockType, err := s.GetBlockTypeByName(ctx, createItem.Type)
		if err != nil {
			response.Errors = append(response.Errors, dto.BatchError{
				OperationID: createItem.TempID,
				Type:        "create",
				Error:       fmt.Sprintf("invalid block type: %v", err),
			})
			continue
		}

		var content *string
		if createItem.Content != "" {
			content = &createItem.Content
		}

		var parentBlockID *uuid.UUID
		if createItem.ParentBlockID != "" {
			parsedID, err := uuid.Parse(createItem.ParentBlockID)
			if err != nil {
				response.Errors = append(response.Errors, dto.BatchError{
					OperationID: createItem.TempID,
					Type:        "create",
					Error:       fmt.Sprintf("invalid parent block ID: %v", err),
				})
				continue
			}
			parentBlockID = &parsedID
		}

		var block *model.Block
		if createItem.BlockConfig != nil && len(createItem.BlockConfig) > 0 {
			metadata := model.JSONBMap(createItem.BlockConfig)
			block, err = s.CreateBlockWithMetadata(ctx, pageID, blockType, content, metadata, int64(createItem.Position), parentBlockID)
		} else {
			block, err = s.CreateBlock(ctx, pageID, blockType, content, int64(createItem.Position), parentBlockID)
		}

		if err != nil {
			response.Errors = append(response.Errors, dto.BatchError{
				OperationID: createItem.TempID,
				Type:        "create",
				Error:       err.Error(),
			})
			continue
		}

		blockContent := ""
		if block.Content != nil {
			blockContent = *block.Content
		}

		typeName := ""
		if block.BlockType != nil {
			typeName = block.BlockType.Name
		}

		var parentID string
		if block.ParentBlockID != nil {
			parentID = block.ParentBlockID.String()
		}

		response.Creates = append(response.Creates, dto.BatchCreateResponse{
			TempID: createItem.TempID,
			Block: dto.BlockResponse{
				ID:            block.ID.String(),
				PageID:        block.PageID.String(),
				Type:          typeName,
				Content:       blockContent,
				Position:      int(block.Rank),
				ParentBlockID: parentID,
				BlockConfig:   block.Metadata,
				CreatedAt:     block.CreatedAt,
				UpdatedAt:     block.UpdatedAt,
			},
		})
	}

	// Process updates
	for _, updateItem := range req.Updates {
		blockID, err := uuid.Parse(updateItem.ID)
		if err != nil {
			response.Errors = append(response.Errors, dto.BatchError{
				OperationID: updateItem.ID,
				Type:        "update",
				Error:       fmt.Sprintf("invalid block ID: %v", err),
			})
			continue
		}

		// Get existing block
		block, err := s.GetBlock(ctx, blockID)
		if err != nil {
			response.Errors = append(response.Errors, dto.BatchError{
				OperationID: updateItem.ID,
				Type:        "update",
				Error:       fmt.Sprintf("block not found: %v", err),
			})
			continue
		}

		// Update content if provided
		var content *string
		if updateItem.Content != "" {
			content = &updateItem.Content
		} else if block.Content != nil {
			content = block.Content
		}

		var updatedBlock *model.Block
		if updateItem.BlockConfig != nil && len(updateItem.BlockConfig) > 0 {
			metadata := model.JSONBMap(updateItem.BlockConfig)
			updatedBlock, err = s.UpdateBlockWithMetadata(ctx, blockID, content, metadata)
		} else {
			updatedBlock, err = s.UpdateBlock(ctx, blockID, content)
		}

		if err != nil {
			response.Errors = append(response.Errors, dto.BatchError{
				OperationID: updateItem.ID,
				Type:        "update",
				Error:       err.Error(),
			})
			continue
		}

		// Update type if provided
		if updateItem.Type != "" && (block.BlockType == nil || block.BlockType.Name != updateItem.Type) {
			blockType, err := s.GetBlockTypeByName(ctx, updateItem.Type)
			if err != nil {
				response.Errors = append(response.Errors, dto.BatchError{
					OperationID: updateItem.ID,
					Type:        "update",
					Error:       fmt.Sprintf("invalid block type: %v", err),
				})
				continue
			}

			// Update block type in model
			updatedBlock.TypeID = blockType.ID
			updatedBlock.BlockType = blockType

			// Persist type change
			if err := s.repo.Update(ctx, updatedBlock); err != nil {
				response.Errors = append(response.Errors, dto.BatchError{
					OperationID: updateItem.ID,
					Type:        "update",
					Error:       fmt.Sprintf("failed to update block type: %v", err),
				})
				continue
			}
		}

		// Update position if provided
		if updateItem.Position != nil && *updateItem.Position != int(updatedBlock.Rank) {
			// Update block position (rank/order)
			updatedBlock.Rank = int64(*updateItem.Position)

			// Persist position change
			if err := s.repo.Update(ctx, updatedBlock); err != nil {
				response.Errors = append(response.Errors, dto.BatchError{
					OperationID: updateItem.ID,
					Type:        "update",
					Error:       fmt.Sprintf("failed to update block position: %v", err),
				})
				continue
			}
		}

		blockContent := ""
		if updatedBlock.Content != nil {
			blockContent = *updatedBlock.Content
		}

		typeName := ""
		if updatedBlock.BlockType != nil {
			typeName = updatedBlock.BlockType.Name
		}



		var parentID string
		if updatedBlock.ParentBlockID != nil {
			parentID = updatedBlock.ParentBlockID.String()
		}

		response.Updates = append(response.Updates, dto.BatchUpdateResponse{
			ID: updateItem.ID,
			Block: dto.BlockResponse{
				ID:            updatedBlock.ID.String(),
				PageID:        updatedBlock.PageID.String(),
				Type:          typeName,
				Content:       blockContent,
				Position:      int(updatedBlock.Rank),
				ParentBlockID: parentID,
				BlockConfig:   updatedBlock.Metadata,
				CreatedAt:     updatedBlock.CreatedAt,
				UpdatedAt:     updatedBlock.UpdatedAt,
			},
		})
	}

	// Process deletes
	for _, blockIDStr := range req.Deletes {
		blockID, err := uuid.Parse(blockIDStr)
		if err != nil {
			response.Errors = append(response.Errors, dto.BatchError{
				OperationID: blockIDStr,
				Type:        "delete",
				Error:       fmt.Sprintf("invalid block ID: %v", err),
			})
			continue
		}

		err = s.DeleteBlock(ctx, blockID)
		if err != nil {
			response.Errors = append(response.Errors, dto.BatchError{
				OperationID: blockIDStr,
				Type:        "delete",
				Error:       err.Error(),
			})
			continue
		}

		response.Deletes = append(response.Deletes, blockIDStr)
	}

	logger.LogServiceSuccess(s.logger, "batch_sync",
		zap.Int("creates", len(response.Creates)),
		zap.Int("updates", len(response.Updates)),
		zap.Int("deletes", len(response.Deletes)),
		zap.Int("errors", len(response.Errors)))

	return response, nil
}
