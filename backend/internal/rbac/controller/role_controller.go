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

type RoleController interface {
	CreateRole(*gin.Context)
	GetRoleByID(*gin.Context)
	GetRoleByName(*gin.Context)
	GetRoleByIDWithPermissions(*gin.Context)
	UpdateRole(*gin.Context)
	DeleteRole(*gin.Context)
	ListRoles(*gin.Context)
	AssignPermissionToRole(*gin.Context)
	RemovePermissionFromRole(*gin.Context)
}

type roleController struct {
	roleService       service.RoleService
	permissionService service.PermissionService
	logger            logger.Logger
}

func NewRoleController(roleSvc service.RoleService, permissionSvc service.PermissionService, l logger.Logger) RoleController {
	return &roleController{
		roleService:       roleSvc,
		permissionService: permissionSvc,
		logger:            l,
	}
}

// CreateRole godoc
// @Summary      Create a new role
// @Description  Create a new role with name and description
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        request   body      dto.CreateRoleRequest  true  "Create role request"
// @Success      201       {object}  response.Response{data=dto.RoleResponse}
// @Failure      400       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /roles [post]
func (c *roleController) CreateRole(ctx *gin.Context) {
	api.HandleRequestWithStatus(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, 201, rbacMessages.MsgRoleCreated, func(ctx *gin.Context, req dto.CreateRoleRequest) (interface{}, error) {
		role, err := c.roleService.CreateRole(req.Name, req.Description)
		if err != nil {
			c.logger.Error("Failed to create role", zap.Error(err), zap.String("name", req.Name))
			if strings.Contains(err.Error(), "already exists") {
				return nil, errors.NewConflictError(
					errors.CodeRoleExists,
					rbacMessages.MsgRoleAlreadyExists,
					err,
				)
			}
			return nil, errors.NewInternalError(
				errors.CodeFailedToCreateRole,
				rbacMessages.MsgFailedToCreateRole,
				err,
			)
		}

		return dto.RoleResponse{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		}, nil
	})
}

// GetRoleByID godoc
// @Summary      Get role by ID
// @Description  Get role information by role ID
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "Role ID"
// @Success      200       {object}  response.Response{data=dto.RoleResponse}
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Router       /roles/{id} [get]
func (c *roleController) GetRoleByID(ctx *gin.Context) {
	api.HandleParamRequest(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context, id string) (interface{}, error) {
		role, err := c.roleService.GetRoleByID(id)
		if err != nil {
			c.logger.Error("Failed to get role", zap.Error(err), zap.String("role_id", id))
			if strings.Contains(err.Error(), "not found") {
				return nil, errors.NewNotFoundError(
					errors.CodeRoleNotFound,
					rbacMessages.MsgRoleNotFound,
					err,
				)
			}
			return nil, errors.NewInternalError(
				errors.CodeFailedToGetRole,
				rbacMessages.MsgFailedToRetrieveRole,
				err,
			)
		}

		return dto.RoleResponse{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		}, nil
	})
}

// GetRoleByName godoc
// @Summary      Get role by name
// @Description  Get role information by role name
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        name      query     string  true  "Role name"
// @Success      200       {object}  response.Response{data=dto.RoleResponse}
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Router       /roles/name [get]
func (c *roleController) GetRoleByName(ctx *gin.Context) {
	api.HandleQueryRequest(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context) (interface{}, error) {
		name := ctx.Query("name")
		if name == "" {
			return nil, errors.NewValidationError(
				errors.CodeMissingField,
				"Role name is required",
				nil,
			)
		}

		role, err := c.roleService.GetRoleByName(name)
		if err != nil {
			c.logger.Error("Failed to get role", zap.Error(err), zap.String("name", name))
			if strings.Contains(err.Error(), "not found") {
				return nil, errors.NewNotFoundError(
					errors.CodeRoleNotFound,
					rbacMessages.MsgRoleNotFound,
					err,
				)
			}
			return nil, errors.NewInternalError(
				errors.CodeFailedToGetRole,
				rbacMessages.MsgFailedToRetrieveRole,
				err,
			)
		}

		return dto.RoleResponse{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		}, nil
	})
}

// GetRoleByIDWithPermissions godoc
// @Summary      Get role with permissions
// @Description  Get role information with all assigned permissions
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "Role ID"
// @Success      200       {object}  response.Response{data=dto.RoleWithPermissionsResponse}
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Router       /roles/{id}/permissions [get]
func (c *roleController) GetRoleByIDWithPermissions(ctx *gin.Context) {
	api.HandleParamRequest(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context, id string) (interface{}, error) {
		role, err := c.roleService.GetRoleByIDWithPermissions(id)
		if err != nil {
			c.logger.Error("Failed to get role with permissions", zap.Error(err), zap.String("role_id", id))
			if strings.Contains(err.Error(), "not found") {
				return nil, errors.NewNotFoundError(
					errors.CodeRoleNotFound,
					rbacMessages.MsgRoleNotFound,
					err,
				)
			}
			return nil, errors.NewInternalError(
				errors.CodeFailedToGetRole,
				rbacMessages.MsgFailedToRetrieveRole,
				err,
			)
		}

		permissions := make([]dto.PermissionResponse, len(role.Permissions))
		for i, perm := range role.Permissions {
			permissions[i] = dto.PermissionResponse{
				ID:          perm.ID.String(),
				Name:        perm.Name,
				Resource:    perm.Resource,
				Action:      perm.Action,
				Description: perm.Description,
				IsSystem:    perm.IsSystem,
				CreatedAt:   perm.CreatedAt,
				UpdatedAt:   perm.UpdatedAt,
			}
		}

		return dto.RoleWithPermissionsResponse{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			Permissions: permissions,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		}, nil
	})
}

// UpdateRole godoc
// @Summary      Update role
// @Description  Update role information
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id        path      string              true  "Role ID"
// @Param        request   body      dto.UpdateRoleRequest  true  "Update role request"
// @Success      200       {object}  response.Response{data=dto.RoleResponse}
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /roles/{id} [put]
func (c *roleController) UpdateRole(ctx *gin.Context) {
	id, ok := api.GetParam(ctx, "id", rbacMessages.MsgInvalidRoleID)
	if !ok {
		return
	}

	req, ok := api.BindRequest[dto.UpdateRoleRequest](ctx, c.logger)
	if !ok {
		return
	}

	var namePtr, descPtr *string
	if req.Name != "" {
		namePtr = &req.Name
	}
	if req.Description != "" {
		descPtr = &req.Description
	}

	role, err := c.roleService.UpdateRole(id, namePtr, descPtr)
	if err != nil {
		c.logger.Error("Failed to update role", zap.Error(err), zap.String("role_id", id))
		if strings.Contains(err.Error(), "not found") {
			api.HandleError(ctx, errors.NewNotFoundError(
				errors.CodeRoleNotFound,
				rbacMessages.MsgRoleNotFound,
				err,
			), c.logger)
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			api.HandleError(ctx, errors.NewConflictError(
				errors.CodeRoleExists,
				rbacMessages.MsgRoleAlreadyExists,
				err,
			), c.logger)
			return
		}
		api.HandleError(ctx, errors.NewInternalError(
			errors.CodeFailedToUpdateRole,
			rbacMessages.MsgFailedToUpdateRole,
			err,
		), c.logger)
		return
	}

	api.SendSuccess(ctx, 200, rbacMessages.MsgRoleUpdated, dto.RoleResponse{
		ID:          role.ID.String(),
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	})
}

// DeleteRole godoc
// @Summary      Delete role
// @Description  Delete a role by ID
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "Role ID"
// @Success      200       {object}  response.Response
// @Failure      400       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /roles/{id} [delete]
func (c *roleController) DeleteRole(ctx *gin.Context) {
	api.HandleParamRequestWithMessage(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, rbacMessages.MsgRoleDeleted, func(ctx *gin.Context, id string) (interface{}, error) {
		if err := c.roleService.DeleteRole(id); err != nil {
			c.logger.Error("Failed to delete role", zap.Error(err), zap.String("role_id", id))
			if strings.Contains(err.Error(), "not found") {
				return nil, errors.NewNotFoundError(
					errors.CodeRoleNotFound,
					rbacMessages.MsgRoleNotFound,
					err,
				)
			}
			if strings.Contains(err.Error(), "cannot be deleted") {
				return nil, errors.NewValidationError(
					errors.CodeSystemRoleCannotDelete,
					rbacMessages.MsgSystemRoleCannotDelete,
					err,
				)
			}
			return nil, errors.NewInternalError(
				errors.CodeFailedToDeleteRole,
				rbacMessages.MsgFailedToDeleteRole,
				err,
			)
		}
		return nil, nil
	})
}

// ListRoles godoc
// @Summary      List roles
// @Description  Get a list of roles with pagination
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        limit     query     int     false  "Limit (default: 10)"
// @Param        offset    query     int     false  "Offset (default: 0)"
// @Success      200       {object}  response.Response{data=[]dto.RoleResponse}
// @Failure      500       {object}  response.Response
// @Router       /roles [get]
func (c *roleController) ListRoles(ctx *gin.Context) {
	api.HandleQueryRequestWithMessage(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, rbacMessages.MsgRolesListed, func(ctx *gin.Context) (interface{}, error) {
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

		roles, err := c.roleService.ListRoles(limit, offset)
		if err != nil {
			c.logger.Error("Failed to list roles", zap.Error(err))
			return nil, errors.NewInternalError(
				errors.CodeFailedToListRoles,
				rbacMessages.MsgFailedToListRoles,
				err,
			)
		}

		responses := make([]dto.RoleResponse, len(roles))
		for i, role := range roles {
			responses[i] = dto.RoleResponse{
				ID:          role.ID.String(),
				Name:        role.Name,
				Description: role.Description,
				IsSystem:    role.IsSystem,
				CreatedAt:   role.CreatedAt,
				UpdatedAt:   role.UpdatedAt,
			}
		}

		return responses, nil
	})
}

// AssignPermissionToRole godoc
// @Summary      Assign permission to role
// @Description  Assign a permission to a role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "Role ID"
// @Param        request   body      dto.AssignPermissionToRoleRequest  true  "Assign permission request"
// @Success      200       {object}  response.Response
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /roles/{id}/permissions [post]
func (c *roleController) AssignPermissionToRole(ctx *gin.Context) {
	roleID, ok := api.GetParam(ctx, "id", rbacMessages.MsgInvalidRoleID)
	if !ok {
		return
	}

	api.HandleRequest(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context, req dto.AssignPermissionToRoleRequest) (interface{}, error) {
		// Get permission by name
		permission, err := c.permissionService.GetPermissionByName(req.PermissionName)
		if err != nil {
			c.logger.Error("Failed to get permission", zap.Error(err), zap.String("permission_name", req.PermissionName))
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

		if err := c.roleService.AssignPermissionToRole(roleID, permission.ID.String()); err != nil {
			c.logger.Error("Failed to assign permission to role", zap.Error(err))
			return nil, errors.NewInternalError(
				errors.CodeFailedToAssignPermission,
				rbacMessages.MsgFailedToCreateRole,
				err,
			)
		}

		return gin.H{"message": rbacMessages.MsgPermissionAssigned}, nil
	})
}

// RemovePermissionFromRole godoc
// @Summary      Remove permission from role
// @Description  Remove a permission from a role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "Role ID"
// @Param        request   body      dto.RemovePermissionFromRoleRequest  true  "Remove permission request"
// @Success      200       {object}  response.Response
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /roles/{id}/permissions [delete]
func (c *roleController) RemovePermissionFromRole(ctx *gin.Context) {
	roleID, ok := api.GetParam(ctx, "id", rbacMessages.MsgInvalidRoleID)
	if !ok {
		return
	}

	api.HandleRequest(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context, req dto.RemovePermissionFromRoleRequest) (interface{}, error) {
		// Get permission by name
		permission, err := c.permissionService.GetPermissionByName(req.PermissionName)
		if err != nil {
			c.logger.Error("Failed to get permission", zap.Error(err), zap.String("permission_name", req.PermissionName))
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

		if err := c.roleService.RemovePermissionFromRole(roleID, permission.ID.String()); err != nil {
			c.logger.Error("Failed to remove permission from role", zap.Error(err))
			return nil, errors.NewInternalError(
				errors.CodeFailedToRemovePermission,
				rbacMessages.MsgFailedToDeleteRole,
				err,
			)
		}

		return gin.H{"message": rbacMessages.MsgPermissionRemoved}, nil
	})
}
