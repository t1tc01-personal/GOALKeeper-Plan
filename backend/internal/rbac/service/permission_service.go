package service

import (
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/rbac/model"
	rbacRepo "goalkeeper-plan/internal/rbac/repository"
	"goalkeeper-plan/internal/validation"
)

type PermissionService interface {
	CreatePermission(name, resource, action, description string) (*model.Permission, error)
	GetPermissionByID(id string) (*model.Permission, error)
	GetPermissionByName(name string) (*model.Permission, error)
	GetPermissionByResourceAndAction(resource, action string) (*model.Permission, error)
	UpdatePermission(id string, name, resource, action, description *string) (*model.Permission, error)
	DeletePermission(id string) error
	ListPermissions(limit, offset int) ([]*model.Permission, error)
	ListPermissionsByResource(resource string) ([]*model.Permission, error)
}

type permissionService struct {
	permissionRepository rbacRepo.PermissionRepository
	logger               logger.Logger
}

func NewPermissionService(permissionRepo rbacRepo.PermissionRepository, l logger.Logger) PermissionService {
	return &permissionService{
		permissionRepository: permissionRepo,
		logger:               l,
	}
}

func (s *permissionService) CreatePermission(name, resource, action, description string) (*model.Permission, error) {
	if err := validation.ValidateString(name, "name", 255); err != nil {
		return nil, err
	}

	if err := validation.ValidateString(resource, "resource", 255); err != nil {
		return nil, err
	}

	if err := validation.ValidateString(action, "action", 255); err != nil {
		return nil, err
	}

	// Check if permission already exists
	existing, _ := s.permissionRepository.GetByName(name)
	if existing != nil {
		appErr := errors.New("permission already exists")
		logger.LogServiceError(s.logger, "create_permission_duplicate", appErr, zap.String("name", name))
		return nil, appErr
	}

	permission := &model.Permission{
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: description,
		IsSystem:    false,
	}

	if err := s.permissionRepository.Create(permission); err != nil {
		logger.LogServiceError(s.logger, "create_permission_save", err, zap.String("name", name))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "create_permission", zap.String("name", name), zap.String("id", permission.ID.String()))
	return permission, nil
}

func (s *permissionService) GetPermissionByID(id string) (*model.Permission, error) {
	if err := validation.ValidateString(id, "id", 255); err != nil {
		return nil, err
	}

	permission, err := s.permissionRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := errors.New("permission not found")
			logger.LogServiceError(s.logger, "get_permission_not_found", appErr, zap.String("id", id))
			return nil, appErr
		}
		logger.LogServiceError(s.logger, "get_permission_fetch", err, zap.String("id", id))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "get_permission_by_id", zap.String("id", id), zap.String("name", permission.Name))
	return permission, nil
}

func (s *permissionService) GetPermissionByName(name string) (*model.Permission, error) {
	if err := validation.ValidateString(name, "name", 255); err != nil {
		return nil, err
	}

	permission, err := s.permissionRepository.GetByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := errors.New("permission not found")
			logger.LogServiceError(s.logger, "get_permission_by_name_not_found", appErr, zap.String("name", name))
			return nil, appErr
		}
		logger.LogServiceError(s.logger, "get_permission_by_name_fetch", err, zap.String("name", name))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "get_permission_by_name", zap.String("name", name))
	return permission, nil
}

func (s *permissionService) GetPermissionByResourceAndAction(resource, action string) (*model.Permission, error) {
	if err := validation.ValidateString(resource, "resource", 255); err != nil {
		return nil, err
	}

	if err := validation.ValidateString(action, "action", 255); err != nil {
		return nil, err
	}

	permission, err := s.permissionRepository.GetByResourceAndAction(resource, action)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := errors.New("permission not found")
			logger.LogServiceError(s.logger, "get_permission_by_resource_action_not_found", appErr, zap.String("resource", resource), zap.String("action", action))
			return nil, appErr
		}
		logger.LogServiceError(s.logger, "get_permission_by_resource_action_fetch", err, zap.String("resource", resource), zap.String("action", action))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "get_permission_by_resource_action", zap.String("resource", resource), zap.String("action", action))
	return permission, nil
}

func (s *permissionService) UpdatePermission(id string, name, resource, action, description *string) (*model.Permission, error) {
	if err := validation.ValidateString(id, "id", 255); err != nil {
		return nil, err
	}

	permission, err := s.permissionRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := errors.New("permission not found")
			logger.LogServiceError(s.logger, "update_permission_not_found", appErr, zap.String("id", id))
			return nil, appErr
		}
		logger.LogServiceError(s.logger, "update_permission_fetch", err, zap.String("id", id))
		return nil, err
	}

	// Check if trying to update system permission
	if permission.IsSystem {
		appErr := errors.New("system permissions cannot be updated")
		logger.LogServiceError(s.logger, "update_permission_system", appErr, zap.String("id", id))
		return nil, appErr
	}

	// Check if new name already exists (if name is being changed)
	if name != nil && *name != permission.Name {
		if err := validation.ValidateString(*name, "name", 255); err != nil {
			return nil, err
		}
		existing, _ := s.permissionRepository.GetByName(*name)
		if existing != nil {
			appErr := errors.New("permission name already exists")
			logger.LogServiceError(s.logger, "update_permission_name_exists", appErr, zap.String("new_name", *name))
			return nil, appErr
		}
		permission.Name = *name
	}

	if resource != nil && *resource != "" {
		if err := validation.ValidateString(*resource, "resource", 255); err != nil {
			return nil, err
		}
		permission.Resource = *resource
	}

	if action != nil && *action != "" {
		if err := validation.ValidateString(*action, "action", 255); err != nil {
			return nil, err
		}
		permission.Action = *action
	}

	if description != nil {
		permission.Description = *description
	}

	if err := s.permissionRepository.Update(permission); err != nil {
		logger.LogServiceError(s.logger, "update_permission_save", err, zap.String("id", id))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "update_permission", zap.String("id", id), zap.String("name", permission.Name))
	return permission, nil
}

func (s *permissionService) DeletePermission(id string) error {
	if err := validation.ValidateString(id, "id", 255); err != nil {
		return err
	}

	permission, err := s.permissionRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := errors.New("permission not found")
			logger.LogServiceError(s.logger, "delete_permission_not_found", appErr, zap.String("id", id))
			return appErr
		}
		logger.LogServiceError(s.logger, "delete_permission_fetch", err, zap.String("id", id))
		return err
	}

	if permission.IsSystem {
		appErr := errors.New("system permissions cannot be deleted")
		logger.LogServiceError(s.logger, "delete_permission_system", appErr, zap.String("id", id))
		return appErr
	}

	if err := s.permissionRepository.Delete(id); err != nil {
		logger.LogServiceError(s.logger, "delete_permission_save", err, zap.String("id", id))
		return err
	}

	logger.LogServiceSuccess(s.logger, "delete_permission", zap.String("id", id))
	return nil
}

func (s *permissionService) ListPermissions(limit, offset int) ([]*model.Permission, error) {
	if err := validation.ValidateMinValue(limit, "limit", 1); err != nil {
		return nil, err
	}

	if err := validation.ValidateMinValue(offset, "offset", 0); err != nil {
		return nil, err
	}

	permissions, err := s.permissionRepository.List(limit, offset)
	if err != nil {
		logger.LogServiceError(s.logger, "list_permissions", err, zap.Int("limit", limit), zap.Int("offset", offset))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "list_permissions", zap.Int("limit", limit), zap.Int("offset", offset), zap.Int("count", len(permissions)))
	return permissions, nil
}

func (s *permissionService) ListPermissionsByResource(resource string) ([]*model.Permission, error) {
	if err := validation.ValidateString(resource, "resource", 255); err != nil {
		return nil, err
	}

	permissions, err := s.permissionRepository.ListByResource(resource)
	if err != nil {
		logger.LogServiceError(s.logger, "list_permissions_by_resource", err, zap.String("resource", resource))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "list_permissions_by_resource", zap.String("resource", resource), zap.Int("count", len(permissions)))
	return permissions, nil
}
