package controller

import (
	"goalkeeper-plan/internal/api"
	"goalkeeper-plan/internal/errors"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/page/dto"
	pageMessages "goalkeeper-plan/internal/page/messages"
	"goalkeeper-plan/internal/page/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PageController interface {
	CreatePage(*gin.Context)
	GetPage(*gin.Context)
	ListPages(*gin.Context)
	UpdatePage(*gin.Context)
	DeletePage(*gin.Context)
	GetHierarchy(*gin.Context)
}

type pageController struct {
	service service.PageService
	logger  logger.Logger
}

func NewPageController(s service.PageService, l logger.Logger) PageController {
	return &pageController{
		service: s,
		logger:  l,
	}
}

func (c *pageController) CreatePage(ctx *gin.Context) {
	api.HandleRequestWithStatus(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, 201, pageMessages.MsgPageCreated, func(ctx *gin.Context, req dto.CreatePageRequest) (interface{}, error) {
		workspaceID, err := uuid.Parse(req.WorkspaceID)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				pageMessages.MsgInvalidWorkspaceID,
				err,
			)
		}

		var parentPageID *uuid.UUID
		if req.ParentPageID != nil {
			id, err := uuid.Parse(*req.ParentPageID)
			if err != nil {
				return nil, errors.NewValidationError(
					errors.CodeInvalidID,
					pageMessages.MsgInvalidPageID,
					err,
				)
			}
			parentPageID = &id
		}

		page, err := c.service.CreatePage(ctx, workspaceID, req.Title, parentPageID)
		if err != nil {
			c.logger.Error("Failed to create page", zap.Error(err), zap.String("workspace_id", req.WorkspaceID))
			return nil, errors.NewInternalError(
				errors.CodeFailedToCreatePage,
				pageMessages.MsgFailedToCreatePage,
				err,
			)
		}

		return dto.PageResponse{
			ID:          page.ID.String(),
			WorkspaceID: page.WorkspaceID.String(),
			Title:       page.Title,
			CreatedAt:   page.CreatedAt,
			UpdatedAt:   page.UpdatedAt,
		}, nil
	})
}

func (c *pageController) GetPage(ctx *gin.Context) {
	api.HandleParamRequest(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context, id string) (interface{}, error) {
		pageID, err := uuid.Parse(id)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				pageMessages.MsgInvalidPageID,
				err,
			)
		}

		page, err := c.service.GetPage(ctx, pageID)
		if err != nil {
			c.logger.Error("Failed to get page", zap.Error(err), zap.String("page_id", id))
			return nil, errors.NewNotFoundError(
				errors.CodePageNotFound,
				pageMessages.MsgPageNotFound,
				err,
			)
		}

		return dto.PageResponse{
			ID:          page.ID.String(),
			WorkspaceID: page.WorkspaceID.String(),
			Title:       page.Title,
			CreatedAt:   page.CreatedAt,
			UpdatedAt:   page.UpdatedAt,
		}, nil
	})
}

func (c *pageController) ListPages(ctx *gin.Context) {
	api.HandleQueryRequestWithMessage(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, pageMessages.MsgPagesFetched, func(ctx *gin.Context) (interface{}, error) {
		workspaceID := api.GetQuery(ctx, "workspaceId", "")
		if workspaceID == "" {
			return nil, errors.NewValidationError(
				errors.CodeMissingRequired,
				pageMessages.MsgWorkspaceIDRequired,
				nil,
			)
		}

		wsID, err := uuid.Parse(workspaceID)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				pageMessages.MsgInvalidWorkspaceID,
				err,
			)
		}

		pages, err := c.service.ListPagesByWorkspace(ctx, wsID)
		if err != nil {
			c.logger.Error("Failed to list pages", zap.Error(err), zap.String("workspace_id", workspaceID))
			return nil, errors.NewInternalError(
				errors.CodeFailedToFetchPages,
				pageMessages.MsgFailedToFetchPages,
				err,
			)
		}

		responses := make([]dto.PageResponse, len(pages))
		for i, p := range pages {
			responses[i] = dto.PageResponse{
				ID:          p.ID.String(),
				WorkspaceID: p.WorkspaceID.String(),
				Title:       p.Title,
				CreatedAt:   p.CreatedAt,
				UpdatedAt:   p.UpdatedAt,
			}
		}

		return responses, nil
	})
}

func (c *pageController) UpdatePage(ctx *gin.Context) {
	id, ok := api.GetParam(ctx, "id", pageMessages.MsgInvalidPageID)
	if !ok {
		return
	}

	pageID, err := uuid.Parse(id)
	if err != nil {
		api.HandleError(ctx, errors.NewValidationError(
			errors.CodeInvalidID,
			pageMessages.MsgInvalidPageID,
			err,
		), c.logger)
		return
	}

	req, ok := api.BindRequest[dto.UpdatePageRequest](ctx, c.logger)
	if !ok {
		return
	}

	page, err := c.service.UpdatePage(ctx, pageID, req.Title, req.ViewConfig)
	if err != nil {
		c.logger.Error("Failed to update page", zap.Error(err), zap.String("page_id", id))
		api.HandleError(ctx, errors.NewInternalError(
			errors.CodeFailedToUpdatePage,
			pageMessages.MsgFailedToUpdatePage,
			err,
		), c.logger)
		return
	}

	api.SendSuccess(ctx, 200, pageMessages.MsgPageUpdated, dto.PageResponse{
		ID:          page.ID.String(),
		WorkspaceID: page.WorkspaceID.String(),
		Title:       page.Title,
		CreatedAt:   page.CreatedAt,
		UpdatedAt:   page.UpdatedAt,
	})
}

func (c *pageController) DeletePage(ctx *gin.Context) {
	api.HandleParamRequestWithMessage(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, pageMessages.MsgPageDeleted, func(ctx *gin.Context, id string) (interface{}, error) {
		pageID, err := uuid.Parse(id)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				pageMessages.MsgInvalidPageID,
				err,
			)
		}

		if err := c.service.DeletePage(ctx, pageID); err != nil {
			c.logger.Error("Failed to delete page", zap.Error(err), zap.String("page_id", id))
			return nil, errors.NewInternalError(
				errors.CodeFailedToDeletePage,
				pageMessages.MsgFailedToDeletePage,
				err,
			)
		}

		return nil, nil
	})
}

func (c *pageController) GetHierarchy(ctx *gin.Context) {
	api.HandleQueryRequestWithMessage(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, pageMessages.MsgHierarchyFetched, func(ctx *gin.Context) (interface{}, error) {
		workspaceID := api.GetQuery(ctx, "workspaceId", "")
		if workspaceID == "" {
			return nil, errors.NewValidationError(
				errors.CodeMissingRequired,
				pageMessages.MsgWorkspaceIDRequired,
				nil,
			)
		}

		wsID, err := uuid.Parse(workspaceID)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				pageMessages.MsgInvalidWorkspaceID,
				err,
			)
		}

		pages, err := c.service.GetPageHierarchy(ctx, wsID)
		if err != nil {
			c.logger.Error("Failed to get page hierarchy", zap.Error(err), zap.String("workspace_id", workspaceID))
			return nil, errors.NewInternalError(
				errors.CodeFailedToFetchHierarchy,
				pageMessages.MsgFailedToFetchHierarchy,
				err,
			)
		}

		responses := make([]dto.PageResponse, len(pages))
		for i, p := range pages {
			responses[i] = dto.PageResponse{
				ID:          p.ID.String(),
				WorkspaceID: p.WorkspaceID.String(),
				Title:       p.Title,
				CreatedAt:   p.CreatedAt,
				UpdatedAt:   p.UpdatedAt,
			}
		}

		return responses, nil
	})
}
