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

type RoleService interface {
	CreateRole(name, description string) (*model.Role, error)
	GetRoleByID(id string) (*model.Role, error)
	GetRoleByName(name string) (*model.Role, error)
	GetRoleByIDWithPermissions(id string) (*model.Role, error)
	UpdateRole(id string, name, description *string) (*model.Role, error)
	DeleteRole(id string) error
	ListRoles(limit, offset int) ([]*model.Role, error)
	AssignPermissionToRole(roleID, permissionID string) error
	RemovePermissionFromRole(roleID, permissionID string) error
}

type roleService struct {
	roleRepository       rbacRepo.RoleRepository
	permissionRepository rbacRepo.PermissionRepository
	logger               logger.Logger
}

func NewRoleService(roleRepo rbacRepo.RoleRepository, permissionRepo rbacRepo.PermissionRepository, l logger.Logger) RoleService {
	return &roleService{
		roleRepository:       roleRepo,
		permissionRepository: permissionRepo,
		logger:               l,
	}
}

func (s *roleService) CreateRole(name, description string) (*model.Role, error) {
	if err := validation.ValidateString(name, "name", 255); err != nil {
		return nil, err
	}

	// Check if role already exists
	existing, _ := s.roleRepository.GetByName(name)
	if existing != nil {
		appErr := errors.New("role already exists")
		logger.LogServiceError(s.logger, "create_role_duplicate", appErr, zap.String("name", name))
		return nil, appErr
	}

	role := &model.Role{
		Name:        name,
		Description: description,
		IsSystem:    false,
	}

	if err := s.roleRepository.Create(role); err != nil {
		logger.LogServiceError(s.logger, "create_role_save", err, zap.String("name", name))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "create_role", zap.String("name", name), zap.String("id", role.ID.String()))
	return role, nil
}

func (s *roleService) GetRoleByID(id string) (*model.Role, error) {
	if err := validation.ValidateString(id, "id", 255); err != nil {
		return nil, err
	}

	role, err := s.roleRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := errors.New("role not found")
			logger.LogServiceError(s.logger, "get_role_not_found", appErr, zap.String("id", id))
			return nil, appErr
		}
		logger.LogServiceError(s.logger, "get_role_fetch", err, zap.String("id", id))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "get_role_by_id", zap.String("id", id), zap.String("name", role.Name))
	return role, nil
}

func (s *roleService) GetRoleByName(name string) (*model.Role, error) {
	if err := validation.ValidateString(name, "name", 255); err != nil {
		return nil, err
	}

	role, err := s.roleRepository.GetByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := errors.New("role not found")
			logger.LogServiceError(s.logger, "get_role_by_name_not_found", appErr, zap.String("name", name))
			return nil, appErr
		}
		logger.LogServiceError(s.logger, "get_role_by_name_fetch", err, zap.String("name", name))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "get_role_by_name", zap.String("name", name))
	return role, nil
}

func (s *roleService) GetRoleByIDWithPermissions(id string) (*model.Role, error) {
	if err := validation.ValidateString(id, "id", 255); err != nil {
		return nil, err
	}

	role, err := s.roleRepository.GetByIDWithPermissions(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := errors.New("role not found")
			logger.LogServiceError(s.logger, "get_role_with_permissions_not_found", appErr, zap.String("id", id))
			return nil, appErr
		}
		logger.LogServiceError(s.logger, "get_role_with_permissions_fetch", err, zap.String("id", id))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "get_role_by_id_with_permissions", zap.String("id", id), zap.String("name", role.Name))
	return role, nil
}

func (s *roleService) UpdateRole(id string, name, description *string) (*model.Role, error) {
	if err := validation.ValidateString(id, "id", 255); err != nil {
		return nil, err
	}

	role, err := s.roleRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := errors.New("role not found")
			logger.LogServiceError(s.logger, "update_role_not_found", appErr, zap.String("id", id))
			return nil, appErr
		}
		logger.LogServiceError(s.logger, "update_role_fetch", err, zap.String("id", id))
		return nil, err
	}

	// Check if trying to update system role name
	if role.IsSystem && name != nil && *name != role.Name {
		appErr := errors.New("cannot change name of system role")
		logger.LogServiceError(s.logger, "update_role_system_name", appErr, zap.String("id", id))
		return nil, appErr
	}

	// Check if new name already exists (if name is being changed)
	if name != nil && *name != role.Name {
		if err := validation.ValidateString(*name, "name", 255); err != nil {
			return nil, err
		}
		existing, _ := s.roleRepository.GetByName(*name)
		if existing != nil {
			appErr := errors.New("role name already exists")
			logger.LogServiceError(s.logger, "update_role_name_exists", appErr, zap.String("new_name", *name))
			return nil, appErr
		}
		role.Name = *name
	}

	if description != nil {
		role.Description = *description
	}

	if err := s.roleRepository.Update(role); err != nil {
		logger.LogServiceError(s.logger, "update_role_save", err, zap.String("id", id))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "update_role", zap.String("id", id), zap.String("name", role.Name))
	return role, nil
}

func (s *roleService) DeleteRole(id string) error {
	if err := validation.ValidateString(id, "id", 255); err != nil {
		return err
	}

	role, err := s.roleRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := errors.New("role not found")
			logger.LogServiceError(s.logger, "delete_role_not_found", appErr, zap.String("id", id))
			return appErr
		}
		logger.LogServiceError(s.logger, "delete_role_fetch", err, zap.String("id", id))
		return err
	}

	if role.IsSystem {
		appErr := errors.New("system roles cannot be deleted")
		logger.LogServiceError(s.logger, "delete_role_system", appErr, zap.String("id", id))
		return appErr
	}

	if err := s.roleRepository.Delete(id); err != nil {
		logger.LogServiceError(s.logger, "delete_role_save", err, zap.String("id", id))
		return err
	}

	logger.LogServiceSuccess(s.logger, "delete_role", zap.String("id", id))
	return nil
}

func (s *roleService) ListRoles(limit, offset int) ([]*model.Role, error) {
	if err := validation.ValidateMinValue(limit, "limit", 1); err != nil {
		return nil, err
	}

	if err := validation.ValidateMinValue(offset, "offset", 0); err != nil {
		return nil, err
	}

	roles, err := s.roleRepository.List(limit, offset)
	if err != nil {
		logger.LogServiceError(s.logger, "list_roles", err, zap.Int("limit", limit), zap.Int("offset", offset))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "list_roles", zap.Int("limit", limit), zap.Int("offset", offset), zap.Int("count", len(roles)))
	return roles, nil
}

func (s *roleService) AssignPermissionToRole(roleID, permissionID string) error {
	if err := validation.ValidateString(roleID, "role_id", 255); err != nil {
		return err
	}

	if err := validation.ValidateString(permissionID, "permission_id", 255); err != nil {
		return err
	}

	if err := s.roleRepository.AssignPermission(roleID, permissionID); err != nil {
		logger.LogServiceError(s.logger, "assign_permission_to_role", err, zap.String("role_id", roleID), zap.String("permission_id", permissionID))
		return err
	}

	logger.LogServiceSuccess(s.logger, "assign_permission_to_role", zap.String("role_id", roleID), zap.String("permission_id", permissionID))
	return nil
}

func (s *roleService) RemovePermissionFromRole(roleID, permissionID string) error {
	if err := validation.ValidateString(roleID, "role_id", 255); err != nil {
		return err
	}

	if err := validation.ValidateString(permissionID, "permission_id", 255); err != nil {
		return err
	}

	if err := s.roleRepository.RemovePermission(roleID, permissionID); err != nil {
		logger.LogServiceError(s.logger, "remove_permission_from_role", err, zap.String("role_id", roleID), zap.String("permission_id", permissionID))
		return err
	}

	logger.LogServiceSuccess(s.logger, "remove_permission_from_role", zap.String("role_id", roleID), zap.String("permission_id", permissionID))
	return nil
}
