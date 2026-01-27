package controller

import (
	"goalkeeper-plan/internal/api"
	"goalkeeper-plan/internal/errors"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/workspace/dto"
	workspaceMessages "goalkeeper-plan/internal/workspace/messages"
	"goalkeeper-plan/internal/workspace/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WorkspaceController interface {
	CreateWorkspace(*gin.Context)
	GetWorkspace(*gin.Context)
	ListWorkspaces(*gin.Context)
	UpdateWorkspace(*gin.Context)
	DeleteWorkspace(*gin.Context)
}

type workspaceController struct {
	service service.WorkspaceService
	logger  logger.Logger
}

func NewWorkspaceController(s service.WorkspaceService, l logger.Logger) WorkspaceController {
	return &workspaceController{
		service: s,
		logger:  l,
	}
}

func (c *workspaceController) CreateWorkspace(ctx *gin.Context) {
	api.HandleRequestWithStatus(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, 201, workspaceMessages.MsgWorkspaceCreated, func(ctx *gin.Context, req dto.CreateWorkspaceRequest) (interface{}, error) {
		// Extract user ID from context (set by auth middleware)
		userID, err := api.GetUserIDFromContext(ctx)
		if err != nil {
			// If no user ID in context, return error (authentication required)
			c.logger.Error("Failed to get user ID from context", zap.Error(err))
			return nil, errors.NewAuthenticationError(
				errors.CodeAuthRequired,
				"Authentication required",
				err,
			)
		}

		var description *string
		if req.Description != "" {
			description = &req.Description
		}

		ws, err := c.service.CreateWorkspace(ctx, userID, req.Name, description)
		if err != nil {
			c.logger.Error("Failed to create workspace", zap.Error(err), zap.String("name", req.Name))
			return nil, errors.NewInternalError(
				errors.CodeFailedToCreateWorkspace,
				workspaceMessages.MsgFailedToCreateWorkspace,
				err,
			)
		}

		desc := ""
		if ws.Description != nil {
			desc = *ws.Description
		}

		return dto.WorkspaceResponse{
			ID:          ws.ID.String(),
			Name:        ws.Name,
			Description: desc,
			CreatedAt:   ws.CreatedAt,
			UpdatedAt:   ws.UpdatedAt,
		}, nil
	})
}

func (c *workspaceController) GetWorkspace(ctx *gin.Context) {
	api.HandleParamRequest(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context, id string) (interface{}, error) {
		workspaceID, err := uuid.Parse(id)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				workspaceMessages.MsgInvalidWorkspaceID,
				err,
			)
		}

		ws, err := c.service.GetWorkspace(ctx, workspaceID)
		if err != nil {
			c.logger.Error("Failed to get workspace", zap.Error(err), zap.String("workspace_id", id))
			return nil, errors.NewNotFoundError(
				errors.CodeWorkspaceNotFound,
				workspaceMessages.MsgWorkspaceNotFound,
				err,
			)
		}

		desc := ""
		if ws.Description != nil {
			desc = *ws.Description
		}

		return dto.WorkspaceResponse{
			ID:          ws.ID.String(),
			Name:        ws.Name,
			Description: desc,
			CreatedAt:   ws.CreatedAt,
			UpdatedAt:   ws.UpdatedAt,
		}, nil
	})
}

func (c *workspaceController) ListWorkspaces(ctx *gin.Context) {
	api.HandleQueryRequestWithMessage(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, workspaceMessages.MsgWorkspacesListed, func(ctx *gin.Context) (interface{}, error) {
		workspaces, err := c.service.ListWorkspaces(ctx)
		if err != nil {
			c.logger.Error("Failed to list workspaces", zap.Error(err))
			return nil, errors.NewInternalError(
				errors.CodeFailedToListWorkspaces,
				workspaceMessages.MsgFailedToListWorkspaces,
				err,
			)
		}

		responses := make([]dto.WorkspaceResponse, len(workspaces))
		for i, ws := range workspaces {
			desc := ""
			if ws.Description != nil {
				desc = *ws.Description
			}

			responses[i] = dto.WorkspaceResponse{
				ID:          ws.ID.String(),
				Name:        ws.Name,
				Description: desc,
				CreatedAt:   ws.CreatedAt,
				UpdatedAt:   ws.UpdatedAt,
			}
		}

		return responses, nil
	})
}

func (c *workspaceController) UpdateWorkspace(ctx *gin.Context) {
	id, ok := api.GetParam(ctx, "id", workspaceMessages.MsgInvalidWorkspaceID)
	if !ok {
		return
	}

	req, ok := api.BindRequest[dto.UpdateWorkspaceRequest](ctx, c.logger)
	if !ok {
		return
	}

	workspaceID, err := uuid.Parse(id)
	if err != nil {
		api.HandleError(ctx, errors.NewValidationError(
			errors.CodeInvalidID,
			workspaceMessages.MsgInvalidWorkspaceID,
			err,
		), c.logger)
		return
	}

	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	ws, err := c.service.UpdateWorkspace(ctx, workspaceID, req.Name, description)
	if err != nil {
		c.logger.Error("Failed to update workspace", zap.Error(err), zap.String("workspace_id", id))
		api.HandleError(ctx, errors.NewInternalError(
			errors.CodeFailedToUpdateWorkspace,
			workspaceMessages.MsgFailedToUpdateWorkspace,
			err,
		), c.logger)
		return
	}

	desc := ""
	if ws.Description != nil {
		desc = *ws.Description
	}

	api.SendSuccess(ctx, 200, workspaceMessages.MsgWorkspaceUpdated, dto.WorkspaceResponse{
		ID:          ws.ID.String(),
		Name:        ws.Name,
		Description: desc,
		CreatedAt:   ws.CreatedAt,
		UpdatedAt:   ws.UpdatedAt,
	})
}

func (c *workspaceController) DeleteWorkspace(ctx *gin.Context) {
	api.HandleParamRequestWithMessage(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, workspaceMessages.MsgWorkspaceDeleted, func(ctx *gin.Context, id string) (interface{}, error) {
		workspaceID, err := uuid.Parse(id)
		if err != nil {
			return nil, errors.NewValidationError(
				errors.CodeInvalidID,
				workspaceMessages.MsgInvalidWorkspaceID,
				err,
			)
		}

		if err := c.service.DeleteWorkspace(ctx, workspaceID); err != nil {
			c.logger.Error("Failed to delete workspace", zap.Error(err), zap.String("workspace_id", id))
			return nil, errors.NewInternalError(
				errors.CodeFailedToDeleteWorkspace,
				workspaceMessages.MsgFailedToDeleteWorkspace,
				err,
			)
		}
		return nil, nil
	})
}
