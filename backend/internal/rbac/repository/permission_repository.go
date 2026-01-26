package repository

import (
	"goalkeeper-plan/internal/rbac/model"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	Create(permission *model.Permission) error
	GetByID(id string) (*model.Permission, error)
	GetByName(name string) (*model.Permission, error)
	GetByResourceAndAction(resource, action string) (*model.Permission, error)
	Update(permission *model.Permission) error
	Delete(id string) error
	List(limit, offset int) ([]*model.Permission, error)
	ListByResource(resource string) ([]*model.Permission, error)
	WithTrx(tx *gorm.DB) PermissionRepository
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) WithTrx(tx *gorm.DB) PermissionRepository {
	return &permissionRepository{db: tx}
}

func (r *permissionRepository) Create(permission *model.Permission) error {
	return r.db.Create(permission).Error
}

func (r *permissionRepository) GetByID(id string) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.Where("id = ?", id).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetByName(name string) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.Where("name = ?", name).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetByResourceAndAction(resource, action string) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.Where("resource = ? AND action = ?", resource, action).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) Update(permission *model.Permission) error {
	return r.db.Save(permission).Error
}

func (r *permissionRepository) Delete(id string) error {
	// Check if it's a system permission
	var permission model.Permission
	if err := r.db.Where("id = ?", id).First(&permission).Error; err != nil {
		return err
	}
	if permission.IsSystem {
		return gorm.ErrRecordNotFound // Prevent deletion of system permissions
	}
	return r.db.Delete(&model.Permission{}, "id = ?", id).Error
}

func (r *permissionRepository) List(limit, offset int) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := r.db.Limit(limit).Offset(offset).Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) ListByResource(resource string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := r.db.Where("resource = ?", resource).Find(&permissions).Error
	return permissions, err
}
