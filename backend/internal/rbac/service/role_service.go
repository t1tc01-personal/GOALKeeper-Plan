package service

import (
	"errors"
	"fmt"

	"goalkeeper-plan/internal/rbac/model"
	rbacRepo "goalkeeper-plan/internal/rbac/repository"

	"gorm.io/gorm"
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
}

func NewRoleService(roleRepo rbacRepo.RoleRepository, permissionRepo rbacRepo.PermissionRepository) RoleService {
	return &roleService{
		roleRepository:       roleRepo,
		permissionRepository: permissionRepo,
	}
}

func (s *roleService) CreateRole(name, description string) (*model.Role, error) {
	// Check if role already exists
	existing, _ := s.roleRepository.GetByName(name)
	if existing != nil {
		return nil, errors.New("role already exists")
	}

	role := &model.Role{
		Name:        name,
		Description: description,
		IsSystem:    false,
	}

	if err := s.roleRepository.Create(role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

func (s *roleService) GetRoleByID(id string) (*model.Role, error) {
	role, err := s.roleRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

func (s *roleService) GetRoleByName(name string) (*model.Role, error) {
	role, err := s.roleRepository.GetByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

func (s *roleService) GetRoleByIDWithPermissions(id string) (*model.Role, error) {
	role, err := s.roleRepository.GetByIDWithPermissions(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

func (s *roleService) UpdateRole(id string, name, description *string) (*model.Role, error) {
	role, err := s.roleRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Check if trying to update system role name
	if role.IsSystem && name != nil && *name != role.Name {
		return nil, errors.New("cannot change name of system role")
	}

	// Check if new name already exists (if name is being changed)
	if name != nil && *name != role.Name {
		existing, _ := s.roleRepository.GetByName(*name)
		if existing != nil {
			return nil, errors.New("role name already exists")
		}
		role.Name = *name
	}

	if description != nil {
		role.Description = *description
	}

	if err := s.roleRepository.Update(role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

func (s *roleService) DeleteRole(id string) error {
	role, err := s.roleRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	if role.IsSystem {
		return errors.New("system roles cannot be deleted")
	}

	if err := s.roleRepository.Delete(id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

func (s *roleService) ListRoles(limit, offset int) ([]*model.Role, error) {
	roles, err := s.roleRepository.List(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return roles, nil
}

func (s *roleService) AssignPermissionToRole(roleID, permissionID string) error {
	if err := s.roleRepository.AssignPermission(roleID, permissionID); err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}
	return nil
}

func (s *roleService) RemovePermissionFromRole(roleID, permissionID string) error {
	if err := s.roleRepository.RemovePermission(roleID, permissionID); err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}
	return nil
}
