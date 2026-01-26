package controller

import (
	"goalkeeper-plan/internal/api"
	"goalkeeper-plan/internal/errors"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/rbac/dto"
	rbacMessages "goalkeeper-plan/internal/rbac/messages"
	"goalkeeper-plan/internal/rbac/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PermissionController interface {
	CreatePermission(*gin.Context)
	GetPermissionByID(*gin.Context)
	GetPermissionByName(*gin.Context)
	GetPermissionByResourceAndAction(*gin.Context)
	UpdatePermission(*gin.Context)
	DeletePermission(*gin.Context)
	ListPermissions(*gin.Context)
	ListPermissionsByResource(*gin.Context)
}

type permissionController struct {
	permissionService service.PermissionService
	logger            logger.Logger
}

func NewPermissionController(permissionSvc service.PermissionService, l logger.Logger) PermissionController {
	return &permissionController{
		permissionService: permissionSvc,
		logger:            l,
	}
}

// CreatePermission godoc
// @Summary      Create a new permission
// @Description  Create a new permission with name, resource, action, and description
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        request   body      dto.CreatePermissionRequest  true  "Create permission request"
// @Success      201       {object}  response.Response{data=dto.PermissionResponse}
// @Failure      400       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /permissions [post]
func (c *permissionController) CreatePermission(ctx *gin.Context) {
	api.HandleRequestWithStatus(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, 201, rbacMessages.MsgPermissionCreated, func(ctx *gin.Context, req dto.CreatePermissionRequest) (interface{}, error) {
		permission, err := c.permissionService.CreatePermission(req.Name, req.Resource, req.Action, req.Description)
		if err != nil {
			c.logger.Error("Failed to create permission", zap.Error(err), zap.String("name", req.Name))
			if strings.Contains(err.Error(), "already exists") {
				return nil, errors.NewConflictError(
					errors.CodePermissionExists,
					rbacMessages.MsgPermissionAlreadyExists,
					err,
				)
			}
			return nil, errors.NewInternalError(
				errors.CodeFailedToCreatePermission,
				rbacMessages.MsgFailedToCreatePermission,
				err,
			)
		}

		return dto.PermissionResponse{
			ID:          permission.ID.String(),
			Name:        permission.Name,
			Resource:    permission.Resource,
			Action:      permission.Action,
			Description: permission.Description,
			IsSystem:    permission.IsSystem,
			CreatedAt:   permission.CreatedAt,
			UpdatedAt:   permission.UpdatedAt,
		}, nil
	})
}

// GetPermissionByID godoc
// @Summary      Get permission by ID
// @Description  Get permission information by permission ID
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "Permission ID"
// @Success      200       {object}  response.Response{data=dto.PermissionResponse}
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Router       /permissions/{id} [get]
func (c *permissionController) GetPermissionByID(ctx *gin.Context) {
	api.HandleParamRequest(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context, id string) (interface{}, error) {
		permission, err := c.permissionService.GetPermissionByID(id)
		if err != nil {
			c.logger.Error("Failed to get permission", zap.Error(err), zap.String("permission_id", id))
			if strings.Contains(err.Error(), "not found") {
				return nil, errors.NewNotFoundError(
					errors.CodePermissionNotFound,
					rbacMessages.MsgPermissionNotFound,
					err,
				)
			}
			return nil, errors.NewInternalError(
				errors.CodeFailedToGetPermission,
				rbacMessages.MsgFailedToRetrievePermission,
				err,
			)
		}

		return dto.PermissionResponse{
			ID:          permission.ID.String(),
			Name:        permission.Name,
			Resource:    permission.Resource,
			Action:      permission.Action,
			Description: permission.Description,
			IsSystem:    permission.IsSystem,
			CreatedAt:   permission.CreatedAt,
			UpdatedAt:   permission.UpdatedAt,
		}, nil
	})
}

// GetPermissionByName godoc
// @Summary      Get permission by name
// @Description  Get permission information by permission name
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        name      query     string  true  "Permission name"
// @Success      200       {object}  response.Response{data=dto.PermissionResponse}
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Router       /permissions/name [get]
func (c *permissionController) GetPermissionByName(ctx *gin.Context) {
	api.HandleQueryRequest(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context) (interface{}, error) {
		name := ctx.Query("name")
		if name == "" {
			return nil, errors.NewValidationError(
				errors.CodeMissingField,
				"Permission name is required",
				nil,
			)
		}

		permission, err := c.permissionService.GetPermissionByName(name)
		if err != nil {
			c.logger.Error("Failed to get permission", zap.Error(err), zap.String("name", name))
			if strings.Contains(err.Error(), "not found") {
				return nil, errors.NewNotFoundError(
					errors.CodePermissionNotFound,
					rbacMessages.MsgPermissionNotFound,
					err,
				)
			}
			return nil, errors.NewInternalError(
				errors.CodeFailedToGetPermission,
				rbacMessages.MsgFailedToRetrievePermission,
				err,
			)
		}

		return dto.PermissionResponse{
			ID:          permission.ID.String(),
			Name:        permission.Name,
			Resource:    permission.Resource,
			Action:      permission.Action,
			Description: permission.Description,
			IsSystem:    permission.IsSystem,
			CreatedAt:   permission.CreatedAt,
			UpdatedAt:   permission.UpdatedAt,
		}, nil
	})
}

// GetPermissionByResourceAndAction godoc
// @Summary      Get permission by resource and action
// @Description  Get permission information by resource and action
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        resource  query     string  true  "Resource name"
// @Param        action    query     string  true  "Action name"
// @Success      200       {object}  response.Response{data=dto.PermissionResponse}
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Router       /permissions/resource-action [get]
func (c *permissionController) GetPermissionByResourceAndAction(ctx *gin.Context) {
	api.HandleQueryRequest(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context) (interface{}, error) {
		resource := ctx.Query("resource")
		action := ctx.Query("action")

		if resource == "" || action == "" {
			return nil, errors.NewValidationError(
				errors.CodeMissingField,
				"Resource and action are required",
				nil,
			)
		}

		permission, err := c.permissionService.GetPermissionByResourceAndAction(resource, action)
		if err != nil {
			c.logger.Error("Failed to get permission", zap.Error(err), zap.String("resource", resource), zap.String("action", action))
			if strings.Contains(err.Error(), "not found") {
				return nil, errors.NewNotFoundError(
					errors.CodePermissionNotFound,
					rbacMessages.MsgPermissionNotFound,
					err,
				)
			}
			return nil, errors.NewInternalError(
				errors.CodeFailedToGetPermission,
				rbacMessages.MsgFailedToRetrievePermission,
				err,
			)
		}

		return dto.PermissionResponse{
			ID:          permission.ID.String(),
			Name:        permission.Name,
			Resource:    permission.Resource,
			Action:      permission.Action,
			Description: permission.Description,
			IsSystem:    permission.IsSystem,
			CreatedAt:   permission.CreatedAt,
			UpdatedAt:   permission.UpdatedAt,
		}, nil
	})
}

// UpdatePermission godoc
// @Summary      Update permission
// @Description  Update permission information
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        id        path      string              true  "Permission ID"
// @Param        request   body      dto.UpdatePermissionRequest  true  "Update permission request"
// @Success      200       {object}  response.Response{data=dto.PermissionResponse}
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /permissions/{id} [put]
func (c *permissionController) UpdatePermission(ctx *gin.Context) {
	id, ok := api.GetParam(ctx, "id", rbacMessages.MsgInvalidPermissionID)
	if !ok {
		return
	}

	req, ok := api.BindRequest[dto.UpdatePermissionRequest](ctx, c.logger)
	if !ok {
		return
	}

	var namePtr, resourcePtr, actionPtr, descPtr *string
	if req.Name != "" {
		namePtr = &req.Name
	}
	if req.Resource != "" {
		resourcePtr = &req.Resource
	}
	if req.Action != "" {
		actionPtr = &req.Action
	}
	if req.Description != "" {
		descPtr = &req.Description
	}

	permission, err := c.permissionService.UpdatePermission(id, namePtr, resourcePtr, actionPtr, descPtr)
	if err != nil {
		c.logger.Error("Failed to update permission", zap.Error(err), zap.String("permission_id", id))
		if strings.Contains(err.Error(), "not found") {
			api.HandleError(ctx, errors.NewNotFoundError(
				errors.CodePermissionNotFound,
				rbacMessages.MsgPermissionNotFound,
				err,
			), c.logger)
			return
		}
		if strings.Contains(err.Error(), "cannot be updated") {
			api.HandleError(ctx, errors.NewValidationError(
				errors.CodeSystemPermissionCannotDelete,
				rbacMessages.MsgSystemPermissionCannotDelete,
				err,
			), c.logger)
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			api.HandleError(ctx, errors.NewConflictError(
				errors.CodePermissionExists,
				rbacMessages.MsgPermissionAlreadyExists,
				err,
			), c.logger)
			return
		}
		api.HandleError(ctx, errors.NewInternalError(
			errors.CodeFailedToUpdatePermission,
			rbacMessages.MsgFailedToUpdatePermission,
			err,
		), c.logger)
		return
	}

	api.SendSuccess(ctx, 200, rbacMessages.MsgPermissionUpdated, dto.PermissionResponse{
		ID:          permission.ID.String(),
		Name:        permission.Name,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Description: permission.Description,
		IsSystem:    permission.IsSystem,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	})
}

// DeletePermission godoc
// @Summary      Delete permission
// @Description  Delete a permission by ID
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "Permission ID"
// @Success      200       {object}  response.Response
// @Failure      400       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /permissions/{id} [delete]
func (c *permissionController) DeletePermission(ctx *gin.Context) {
	api.HandleParamRequestWithMessage(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, rbacMessages.MsgPermissionDeleted, func(ctx *gin.Context, id string) (interface{}, error) {
		if err := c.permissionService.DeletePermission(id); err != nil {
			c.logger.Error("Failed to delete permission", zap.Error(err), zap.String("permission_id", id))
			if strings.Contains(err.Error(), "not found") {
				return nil, errors.NewNotFoundError(
					errors.CodePermissionNotFound,
					rbacMessages.MsgPermissionNotFound,
					err,
				)
			}
			if strings.Contains(err.Error(), "cannot be deleted") {
				return nil, errors.NewValidationError(
					errors.CodeSystemPermissionCannotDelete,
					rbacMessages.MsgSystemPermissionCannotDelete,
					err,
				)
			}
			return nil, errors.NewInternalError(
				errors.CodeFailedToDeletePermission,
				rbacMessages.MsgFailedToDeletePermission,
				err,
			)
		}
		return nil, nil
	})
}

// ListPermissions godoc
// @Summary      List permissions
// @Description  Get a list of permissions with pagination
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        limit     query     int     false  "Limit (default: 10)"
// @Param        offset    query     int     false  "Offset (default: 0)"
// @Success      200       {object}  response.Response{data=[]dto.PermissionResponse}
// @Failure      500       {object}  response.Response
// @Router       /permissions [get]
func (c *permissionController) ListPermissions(ctx *gin.Context) {
	api.HandleQueryRequestWithMessage(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, rbacMessages.MsgPermissionsListed, func(ctx *gin.Context) (interface{}, error) {
		limitStr := api.GetQuery(ctx, "limit", "10")
		offsetStr := api.GetQuery(ctx, "offset", "0")

		limit := 10
		offset := 0

		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}

		permissions, err := c.permissionService.ListPermissions(limit, offset)
		if err != nil {
			c.logger.Error("Failed to list permissions", zap.Error(err))
			return nil, errors.NewInternalError(
				errors.CodeFailedToListPermissions,
				rbacMessages.MsgFailedToListPermissions,
				err,
			)
		}

		responses := make([]dto.PermissionResponse, len(permissions))
		for i, permission := range permissions {
			responses[i] = dto.PermissionResponse{
				ID:          permission.ID.String(),
				Name:        permission.Name,
				Resource:    permission.Resource,
				Action:      permission.Action,
				Description: permission.Description,
				IsSystem:    permission.IsSystem,
				CreatedAt:   permission.CreatedAt,
				UpdatedAt:   permission.UpdatedAt,
			}
		}

		return responses, nil
	})
}

// ListPermissionsByResource godoc
// @Summary      List permissions by resource
// @Description  Get a list of permissions filtered by resource
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        resource  query     string  true  "Resource name"
// @Success      200       {object}  response.Response{data=[]dto.PermissionResponse}
// @Failure      400       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /permissions/resource [get]
func (c *permissionController) ListPermissionsByResource(ctx *gin.Context) {
	api.HandleQueryRequest(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context) (interface{}, error) {
		resource := ctx.Query("resource")
		if resource == "" {
			return nil, errors.NewValidationError(
				errors.CodeMissingField,
				"Resource is required",
				nil,
			)
		}

		permissions, err := c.permissionService.ListPermissionsByResource(resource)
		if err != nil {
			c.logger.Error("Failed to list permissions by resource", zap.Error(err), zap.String("resource", resource))
			return nil, errors.NewInternalError(
				errors.CodeFailedToListPermissions,
				rbacMessages.MsgFailedToListPermissions,
				err,
			)
		}

		responses := make([]dto.PermissionResponse, len(permissions))
		for i, permission := range permissions {
			responses[i] = dto.PermissionResponse{
				ID:          permission.ID.String(),
				Name:        permission.Name,
				Resource:    permission.Resource,
				Action:      permission.Action,
				Description: permission.Description,
				IsSystem:    permission.IsSystem,
				CreatedAt:   permission.CreatedAt,
				UpdatedAt:   permission.UpdatedAt,
			}
		}

		return responses, nil
	})
}
