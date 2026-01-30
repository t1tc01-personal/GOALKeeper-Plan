package controller

import (
	"goalkeeper-plan/internal/api"
	"goalkeeper-plan/internal/block/dto"
	blockMessages "goalkeeper-plan/internal/block/messages"
	"goalkeeper-plan/internal/block/service"
	"goalkeeper-plan/internal/errors"
	"goalkeeper-plan/internal/logger"
	pageService "goalkeeper-plan/internal/page/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BlockController interface {
	CreateBlock(*gin.Context)
	GetBlock(*gin.Context)
	ListBlocks(*gin.Context)
	UpdateBlock(*gin.Context)
	DeleteBlock(*gin.Context)
	ReorderBlocks(*gin.Context)
	BatchSync(*gin.Context)
}

type blockController struct {
	blockService service.BlockService
	pageService  pageService.PageService
	logger       logger.Logger
}

func NewBlockController(blockService service.BlockService, pageService pageService.PageService, l logger.Logger) BlockController {
	return &blockController{
		blockService: blockService,
		pageService:  pageService,
		logger:       l,
	}
}

func (c *blockController) CreateBlock(ctx *gin.Context) {
	api.HandleRequestWithStatus(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, 201, blockMessages.MsgBlockCreated, func(ctx *gin.Context, req dto.CreateBlockRequest) (interface{}, error) {
		pageID, err := uuid.Parse(req.PageID)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				blockMessages.MsgInvalidPageID,
				err,
			)
		}

		// Look up block type by name from database
		blockType, err := c.blockService.GetBlockTypeByName(ctx, req.Type)
		if err != nil {
			c.logger.Error("Failed to get block type", zap.Error(err), zap.String("type", req.Type))
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				blockMessages.MsgInvalidBlockType,
				err,
			)
		}

		var content *string
		if req.Content != "" {
			content = &req.Content
		}

		var parentBlockID *uuid.UUID
		if req.ParentBlockID != "" {
			parsedID, err := uuid.Parse(req.ParentBlockID)
			if err != nil {
				return nil, errors.NewValidationError(
					errors.CodeInvalidID,
					blockMessages.MsgInvalidBlockID, // Reuse or add MsgInvalidParentBlockID
					err,
				)
			}
			parentBlockID = &parsedID
		}

		block, err := c.blockService.CreateBlock(ctx, pageID, blockType, content, int64(req.Position), parentBlockID)
		if err != nil {
			c.logger.Error("Failed to create block", zap.Error(err), zap.String("page_id", req.PageID))
			return nil, errors.NewInternalError(
				errors.CodeFailedToCreateBlock,
				blockMessages.MsgFailedToCreateBlock,
				err,
			)
		}

		blockContent := ""
		if block.Content != nil {
			blockContent = *block.Content
		}

		typeName := ""
		if block.BlockType != nil {
			typeName = block.BlockType.Name
		}

		return dto.BlockResponse{
			ID:          block.ID.String(),
			PageID:      block.PageID.String(),
			Type:        typeName,
			Content:     blockContent,
			Position:    int(block.Rank),
			BlockConfig: block.Metadata,
			CreatedAt:   block.CreatedAt,
			UpdatedAt:   block.UpdatedAt,
		}, nil
	})
}

func (c *blockController) GetBlock(ctx *gin.Context) {
	api.HandleParamRequest(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context, id string) (interface{}, error) {
		blockID, err := uuid.Parse(id)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				blockMessages.MsgInvalidBlockID,
				err,
			)
		}

		block, err := c.blockService.GetBlock(ctx, blockID)
		if err != nil {
			c.logger.Error("Failed to get block", zap.Error(err), zap.String("block_id", id))
			return nil, errors.NewNotFoundError(
				errors.CodeBlockNotFound,
				blockMessages.MsgBlockNotFound,
				err,
			)
		}

		blockContent := ""
		if block.Content != nil {
			blockContent = *block.Content
		}

		typeName := ""
		if block.BlockType != nil {
			typeName = block.BlockType.Name
		}

		return dto.BlockResponse{
			ID:          block.ID.String(),
			PageID:      block.PageID.String(),
			Type:        typeName,
			Content:     blockContent,
			Position:    int(block.Rank),
			BlockConfig: block.Metadata,
			CreatedAt:   block.CreatedAt,
			UpdatedAt:   block.UpdatedAt,
		}, nil
	})
}

func (c *blockController) ListBlocks(ctx *gin.Context) {
	api.HandleQueryRequestWithMessage(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, blockMessages.MsgBlocksListed, func(ctx *gin.Context) (interface{}, error) {
		pageID := ctx.Query("pageId")
		if pageID == "" {
			return nil, errors.NewValidationError(
				errors.CodeMissingField,
				blockMessages.MsgInvalidPageID,
				nil,
			)
		}

		pageUUID, err := uuid.Parse(pageID)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				blockMessages.MsgInvalidPageID,
				err,
			)
		}

		blocks, err := c.blockService.ListBlocksByPage(ctx, pageUUID)
		if err != nil {
			c.logger.Error("Failed to list blocks", zap.Error(err), zap.String("page_id", pageID))
			return nil, errors.NewInternalError(
				errors.CodeFailedToListBlocks,
				blockMessages.MsgFailedToListBlocks,
				err,
			)
		}

		responses := make([]dto.BlockResponse, len(blocks))
		for i, block := range blocks {
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

			responses[i] = dto.BlockResponse{
				ID:            block.ID.String(),
				PageID:        block.PageID.String(),
				Type:          typeName,
				Content:       blockContent,
				Position:      int(block.Rank),
				ParentBlockID: parentID,
				BlockConfig:   block.Metadata,
				CreatedAt:     block.CreatedAt,
				UpdatedAt:     block.UpdatedAt,
			}
		}

		return responses, nil
	})
}

func (c *blockController) UpdateBlock(ctx *gin.Context) {
	id, ok := api.GetParam(ctx, "id", blockMessages.MsgInvalidBlockID)
	if !ok {
		return
	}

	req, ok := api.BindRequest[dto.UpdateBlockRequest](ctx, c.logger)
	if !ok {
		return
	}

	blockID, err := uuid.Parse(id)
	if err != nil {
		api.HandleError(ctx, errors.NewValidationError(
			errors.CodeInvalidID,
			blockMessages.MsgInvalidBlockID,
			err,
		), c.logger)
		return
	}

	var content *string
	if req.Content != "" {
		content = &req.Content
	}

	block, err := c.blockService.UpdateBlock(ctx, blockID, content)
	if err != nil {
		// Let central error handler map domain errors correctly (e.g. BLOCK_NOT_FOUND -> 404)
		c.logger.Error("Failed to update block", zap.Error(err), zap.String("block_id", id))
		api.HandleError(ctx, err, c.logger)
		return
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

	api.SendSuccess(ctx, 200, blockMessages.MsgBlockUpdated, dto.BlockResponse{
		ID:            block.ID.String(),
		PageID:        block.PageID.String(),
		Type:          typeName,
		Content:       blockContent,
		Position:      int(block.Rank),
		ParentBlockID: parentID,
		BlockConfig:   block.Metadata,
		CreatedAt:     block.CreatedAt,
		UpdatedAt:     block.UpdatedAt,
	})
}

func (c *blockController) DeleteBlock(ctx *gin.Context) {
	api.HandleParamRequestWithMessage(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, blockMessages.MsgBlockDeleted, func(ctx *gin.Context, id string) (interface{}, error) {
		blockID, err := uuid.Parse(id)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				blockMessages.MsgInvalidBlockID,
				err,
			)
		}

		if err := c.blockService.DeleteBlock(ctx, blockID); err != nil {
			c.logger.Error("Failed to delete block", zap.Error(err), zap.String("block_id", id))
			return nil, errors.NewInternalError(
				errors.CodeFailedToDeleteBlock,
				blockMessages.MsgFailedToDeleteBlock,
				err,
			)
		}
		return nil, nil
	})
}

func (c *blockController) ReorderBlocks(ctx *gin.Context) {
	api.HandleRequestWithStatus(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, 200, blockMessages.MsgBlocksReordered, func(ctx *gin.Context, req dto.ReorderBlocksRequest) (interface{}, error) {
		// Get pageId from query parameters
		pageIDStr := api.GetQuery(ctx, "pageId", "")
		if pageIDStr == "" {
			return nil, errors.NewValidationError(
				errors.CodeMissingRequired,
				blockMessages.MsgPageIDRequired,
				nil,
			)
		}

		pageID, err := uuid.Parse(pageIDStr)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				blockMessages.MsgInvalidPageID,
				err,
			)
		}

		var blockIDs []uuid.UUID
		for _, id := range req.BlockIDs {
			blockID, err := uuid.Parse(id)
			if err != nil {
				return nil, errors.NewValidationError(
					errors.CodeInvalidID,
					blockMessages.MsgInvalidBlockID,
					err,
				)
			}
			blockIDs = append(blockIDs, blockID)
		}

		if err := c.blockService.ReorderBlocks(ctx, pageID, blockIDs); err != nil {
			c.logger.Error("Failed to reorder blocks", zap.Error(err), zap.String("page_id", pageIDStr))
			return nil, errors.NewInternalError(
				errors.CodeFailedToReorderBlocks,
				blockMessages.MsgFailedToReorderBlocks,
				err,
			)
		}

		return nil, nil
	})
}

func (c *blockController) BatchSync(ctx *gin.Context) {
	api.HandleRequest(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context, req dto.BatchSyncRequest) (interface{}, error) {
		response, err := c.blockService.BatchSync(ctx, &req)
		if err != nil {
			c.logger.Error("Failed to batch sync blocks", zap.Error(err))
			return nil, errors.NewInternalError(
				errors.CodeFailedToUpdateBlock, // Reuse existing error code
				blockMessages.MsgFailedToSyncBlocks,
				err,
			)
		}
		return response, nil
	})
}
