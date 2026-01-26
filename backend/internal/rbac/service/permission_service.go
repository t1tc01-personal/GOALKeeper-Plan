package service

import (
	"errors"
	"fmt"

	"goalkeeper-plan/internal/rbac/model"
	rbacRepo "goalkeeper-plan/internal/rbac/repository"

	"gorm.io/gorm"
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
}

func NewPermissionService(permissionRepo rbacRepo.PermissionRepository) PermissionService {
	return &permissionService{
		permissionRepository: permissionRepo,
	}
}

func (s *permissionService) CreatePermission(name, resource, action, description string) (*model.Permission, error) {
	// Check if permission already exists
	existing, _ := s.permissionRepository.GetByName(name)
	if existing != nil {
		return nil, errors.New("permission already exists")
	}

	permission := &model.Permission{
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: description,
		IsSystem:    false,
	}

	if err := s.permissionRepository.Create(permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

func (s *permissionService) GetPermissionByID(id string) (*model.Permission, error) {
	permission, err := s.permissionRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return permission, nil
}

func (s *permissionService) GetPermissionByName(name string) (*model.Permission, error) {
	permission, err := s.permissionRepository.GetByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return permission, nil
}

func (s *permissionService) GetPermissionByResourceAndAction(resource, action string) (*model.Permission, error) {
	permission, err := s.permissionRepository.GetByResourceAndAction(resource, action)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return permission, nil
}

func (s *permissionService) UpdatePermission(id string, name, resource, action, description *string) (*model.Permission, error) {
	permission, err := s.permissionRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	// Check if trying to update system permission
	if permission.IsSystem {
		return nil, errors.New("system permissions cannot be updated")
	}

	// Check if new name already exists (if name is being changed)
	if name != nil && *name != permission.Name {
		existing, _ := s.permissionRepository.GetByName(*name)
		if existing != nil {
			return nil, errors.New("permission name already exists")
		}
		permission.Name = *name
	}

	if resource != nil {
		permission.Resource = *resource
	}

	if action != nil {
		permission.Action = *action
	}

	if description != nil {
		permission.Description = *description
	}

	if err := s.permissionRepository.Update(permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return permission, nil
}

func (s *permissionService) DeletePermission(id string) error {
	permission, err := s.permissionRepository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("permission not found")
		}
		return fmt.Errorf("failed to get permission: %w", err)
	}

	if permission.IsSystem {
		return errors.New("system permissions cannot be deleted")
	}

	if err := s.permissionRepository.Delete(id); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

func (s *permissionService) ListPermissions(limit, offset int) ([]*model.Permission, error) {
	permissions, err := s.permissionRepository.List(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	return permissions, nil
}

func (s *permissionService) ListPermissionsByResource(resource string) ([]*model.Permission, error) {
	permissions, err := s.permissionRepository.ListByResource(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions by resource: %w", err)
	}
	return permissions, nil
}
