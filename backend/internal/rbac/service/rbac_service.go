package service

import (
	"errors"

	"go.uber.org/zap"

	"goalkeeper-plan/internal/logger"
	rbacRepo "goalkeeper-plan/internal/rbac/repository"
	userRepo "goalkeeper-plan/internal/user/repository"
	"goalkeeper-plan/internal/validation"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrRoleNotFound     = errors.New("role not found")
	ErrUserNotFound     = errors.New("user not found")
)

type RBACService interface {
	// Permission checking
	HasPermission(userID, permissionName string) (bool, error)
	HasAnyPermission(userID string, permissionNames ...string) (bool, error)
	HasAllPermissions(userID string, permissionNames ...string) (bool, error)
	HasPermissionByResourceAction(userID, resource, action string) (bool, error)

	// Role checking
	HasRole(userID, roleName string) (bool, error)
	HasAnyRole(userID string, roleNames ...string) (bool, error)

	// Get user permissions
	GetUserPermissions(userID string) ([]string, error)
	GetUserRoles(userID string) ([]string, error)

	// Role management
	AssignRoleToUser(userID, roleName string) error
	RemoveRoleFromUser(userID, roleName string) error

	// Permission management for roles
	AssignPermissionToRole(roleName, permissionName string) error
	RemovePermissionFromRole(roleName, permissionName string) error
}

type rbacService struct {
	roleRepository       rbacRepo.RoleRepository
	permissionRepository rbacRepo.PermissionRepository
	userRepository       userRepo.UserRepository
	logger               logger.Logger
}

type RBACConfiguration func(s *rbacService) error

func NewRBACService(configs ...RBACConfiguration) (RBACService, error) {
	s := &rbacService{}
	for _, cfg := range configs {
		if err := cfg(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func WithRoleRepository(r rbacRepo.RoleRepository) RBACConfiguration {
	return func(s *rbacService) error {
		s.roleRepository = r
		return nil
	}
}

func WithPermissionRepository(p rbacRepo.PermissionRepository) RBACConfiguration {
	return func(s *rbacService) error {
		s.permissionRepository = p
		return nil
	}
}

func WithUserRepository(u userRepo.UserRepository) RBACConfiguration {
	return func(s *rbacService) error {
		s.userRepository = u
		return nil
	}
}

func WithLogger(l logger.Logger) RBACConfiguration {
	return func(s *rbacService) error {
		s.logger = l
		return nil
	}
}

// HasPermission checks if a user has a specific permission
func (s *rbacService) HasPermission(userID, permissionName string) (bool, error) {
	if err := validation.ValidateString(userID, "user_id", 255); err != nil {
		return false, err
	}

	if err := validation.ValidateString(permissionName, "permission_name", 255); err != nil {
		return false, err
	}

	// Get user roles
	roles, err := s.roleRepository.GetUserRoles(userID)
	if err != nil {
		logger.LogServiceError(s.logger, "has_permission_fetch_roles", err, zap.String("user_id", userID), zap.String("permission", permissionName))
		return false, err
	}

	// Check if any role has the permission
	for _, role := range roles {
		for _, permission := range role.Permissions {
			if permission.Name == permissionName {
				logger.LogServiceSuccess(s.logger, "has_permission_check", zap.String("user_id", userID), zap.String("permission", permissionName), zap.Bool("granted", true))
				return true, nil
			}
		}
	}

	return false, nil
}

// HasAnyPermission checks if a user has at least one of the specified permissions
func (s *rbacService) HasAnyPermission(userID string, permissionNames ...string) (bool, error) {
	for _, permissionName := range permissionNames {
		has, err := s.HasPermission(userID, permissionName)
		if err != nil {
			return false, err
		}
		if has {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if a user has all of the specified permissions
func (s *rbacService) HasAllPermissions(userID string, permissionNames ...string) (bool, error) {
	for _, permissionName := range permissionNames {
		has, err := s.HasPermission(userID, permissionName)
		if err != nil {
			return false, err
		}
		if !has {
			return false, nil
		}
	}
	return true, nil
}

// HasPermissionByResourceAction checks if a user has permission for a resource and action
func (s *rbacService) HasPermissionByResourceAction(userID, resource, action string) (bool, error) {
	if err := validation.ValidateString(userID, "user_id", 255); err != nil {
		return false, err
	}

	if err := validation.ValidateString(resource, "resource", 255); err != nil {
		return false, err
	}

	if err := validation.ValidateString(action, "action", 255); err != nil {
		return false, err
	}

	// Get user roles
	roles, err := s.roleRepository.GetUserRoles(userID)
	if err != nil {
		logger.LogServiceError(s.logger, "has_permission_by_resource", err, zap.String("user_id", userID), zap.String("resource", resource), zap.String("action", action))
		return false, err
	}

	// Check if any role has the permission
	for _, role := range roles {
		for _, permission := range role.Permissions {
			if permission.Resource == resource && permission.Action == action {
				logger.LogServiceSuccess(s.logger, "has_permission_by_resource", zap.String("user_id", userID), zap.String("resource", resource), zap.String("action", action))
				return true, nil
			}
		}
	}

	return false, nil
}

// HasRole checks if a user has a specific role
func (s *rbacService) HasRole(userID, roleName string) (bool, error) {
	if err := validation.ValidateString(userID, "user_id", 255); err != nil {
		return false, err
	}

	if err := validation.ValidateString(roleName, "role_name", 255); err != nil {
		return false, err
	}

	roles, err := s.roleRepository.GetUserRoles(userID)
	if err != nil {
		logger.LogServiceError(s.logger, "has_role_check", err, zap.String("user_id", userID), zap.String("role", roleName))
		return false, err
	}

	for _, role := range roles {
		if role.Name == roleName {
			logger.LogServiceSuccess(s.logger, "has_role_check", zap.String("user_id", userID), zap.String("role", roleName))
			return true, nil
		}
	}

	return false, nil
}

// HasAnyRole checks if a user has at least one of the specified roles
func (s *rbacService) HasAnyRole(userID string, roleNames ...string) (bool, error) {
	roles, err := s.roleRepository.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	roleMap := make(map[string]bool)
	for _, role := range roles {
		roleMap[role.Name] = true
	}

	for _, roleName := range roleNames {
		if roleMap[roleName] {
			return true, nil
		}
	}

	return false, nil
}

// GetUserPermissions returns all permission names for a user
func (s *rbacService) GetUserPermissions(userID string) ([]string, error) {
	if err := validation.ValidateString(userID, "user_id", 255); err != nil {
		return nil, err
	}

	roles, err := s.roleRepository.GetUserRoles(userID)
	if err != nil {
		logger.LogServiceError(s.logger, "get_user_permissions", err, zap.String("user_id", userID))
		return nil, err
	}

	permissionMap := make(map[string]bool)
	for _, role := range roles {
		for _, permission := range role.Permissions {
			permissionMap[permission.Name] = true
		}
	}

	permissions := make([]string, 0, len(permissionMap))
	for permission := range permissionMap {
		permissions = append(permissions, permission)
	}

	logger.LogServiceSuccess(s.logger, "get_user_permissions", zap.String("user_id", userID), zap.Int("count", len(permissions)))
	return permissions, nil
}

// GetUserRoles returns all role names for a user
func (s *rbacService) GetUserRoles(userID string) ([]string, error) {
	if err := validation.ValidateString(userID, "user_id", 255); err != nil {
		return nil, err
	}

	roles, err := s.roleRepository.GetUserRoles(userID)
	if err != nil {
		logger.LogServiceError(s.logger, "get_user_roles", err, zap.String("user_id", userID))
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	logger.LogServiceSuccess(s.logger, "get_user_roles", zap.String("user_id", userID), zap.Int("count", len(roleNames)))
	return roleNames, nil
}

// AssignRoleToUser assigns a role to a user
func (s *rbacService) AssignRoleToUser(userID, roleName string) error {
	if err := validation.ValidateString(userID, "user_id", 255); err != nil {
		return err
	}

	if err := validation.ValidateString(roleName, "role_name", 255); err != nil {
		return err
	}

	// Verify user exists
	_, err := s.userRepository.GetByID(userID)
	if err != nil {
		logger.LogServiceError(s.logger, "assign_role_user_check", ErrUserNotFound, zap.String("user_id", userID), zap.String("role", roleName))
		return ErrUserNotFound
	}

	// Get role by name
	role, err := s.roleRepository.GetByName(roleName)
	if err != nil {
		logger.LogServiceError(s.logger, "assign_role_fetch_role", ErrRoleNotFound, zap.String("role", roleName))
		return ErrRoleNotFound
	}

	if err := s.roleRepository.AssignRoleToUser(userID, role.ID.String()); err != nil {
		logger.LogServiceError(s.logger, "assign_role_save", err, zap.String("user_id", userID), zap.String("role", roleName))
		return err
	}

	logger.LogServiceSuccess(s.logger, "assign_role_to_user", zap.String("user_id", userID), zap.String("role", roleName))
	return nil
}

// RemoveRoleFromUser removes a role from a user
func (s *rbacService) RemoveRoleFromUser(userID, roleName string) error {
	if err := validation.ValidateString(userID, "user_id", 255); err != nil {
		return err
	}

	if err := validation.ValidateString(roleName, "role_name", 255); err != nil {
		return err
	}

	// Get role by name
	role, err := s.roleRepository.GetByName(roleName)
	if err != nil {
		logger.LogServiceError(s.logger, "remove_role_fetch_role", ErrRoleNotFound, zap.String("role", roleName))
		return ErrRoleNotFound
	}

	if err := s.roleRepository.RemoveRoleFromUser(userID, role.ID.String()); err != nil {
		logger.LogServiceError(s.logger, "remove_role_from_user", err, zap.String("user_id", userID), zap.String("role", roleName))
		return err
	}

	logger.LogServiceSuccess(s.logger, "remove_role_from_user", zap.String("user_id", userID), zap.String("role", roleName))
	return nil
}

// AssignPermissionToRole assigns a permission to a role
func (s *rbacService) AssignPermissionToRole(roleName, permissionName string) error {
	if err := validation.ValidateString(roleName, "role_name", 255); err != nil {
		return err
	}

	if err := validation.ValidateString(permissionName, "permission_name", 255); err != nil {
		return err
	}

	// Get role by name
	role, err := s.roleRepository.GetByName(roleName)
	if err != nil {
		logger.LogServiceError(s.logger, "assign_permission_fetch_role", err, zap.String("role", roleName), zap.String("permission", permissionName))
		return err
	}

	// Get permission by name
	permission, err := s.permissionRepository.GetByName(permissionName)
	if err != nil {
		logger.LogServiceError(s.logger, "assign_permission_fetch_permission", err, zap.String("permission", permissionName))
		return err
	}

	if err := s.roleRepository.AssignPermission(role.ID.String(), permission.ID.String()); err != nil {
		logger.LogServiceError(s.logger, "assign_permission_save", err, zap.String("role", roleName), zap.String("permission", permissionName))
		return err
	}

	logger.LogServiceSuccess(s.logger, "assign_permission_to_role", zap.String("role", roleName), zap.String("permission", permissionName))
	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (s *rbacService) RemovePermissionFromRole(roleName, permissionName string) error {
	if err := validation.ValidateString(roleName, "role_name", 255); err != nil {
		return err
	}

	if err := validation.ValidateString(permissionName, "permission_name", 255); err != nil {
		return err
	}

	// Get role by name
	role, err := s.roleRepository.GetByName(roleName)
	if err != nil {
		logger.LogServiceError(s.logger, "remove_permission_fetch_role", err, zap.String("role", roleName), zap.String("permission", permissionName))
		return err
	}

	// Get permission by name
	permission, err := s.permissionRepository.GetByName(permissionName)
	if err != nil {
		logger.LogServiceError(s.logger, "remove_permission_fetch_permission", err, zap.String("permission", permissionName))
		return err
	}

	if err := s.roleRepository.RemovePermission(role.ID.String(), permission.ID.String()); err != nil {
		logger.LogServiceError(s.logger, "remove_permission_save", err, zap.String("role", roleName), zap.String("permission", permissionName))
		return err
	}

	logger.LogServiceSuccess(s.logger, "remove_permission_from_role", zap.String("role", roleName), zap.String("permission", permissionName))
	return nil
}
