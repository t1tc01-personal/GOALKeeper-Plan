package repository

import (
	"goalkeeper-plan/internal/rbac/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(role *model.Role) error
	GetByID(id string) (*model.Role, error)
	GetByName(name string) (*model.Role, error)
	GetByIDWithPermissions(id string) (*model.Role, error)
	Update(role *model.Role) error
	Delete(id string) error
	List(limit, offset int) ([]*model.Role, error)
	AssignPermission(roleID, permissionID string) error
	RemovePermission(roleID, permissionID string) error
	GetUserRoles(userID string) ([]*model.Role, error)
	AssignRoleToUser(userID, roleID string) error
	RemoveRoleFromUser(userID, roleID string) error
	WithTrx(tx *gorm.DB) RoleRepository
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) WithTrx(tx *gorm.DB) RoleRepository {
	return &roleRepository{db: tx}
}

func (r *roleRepository) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) GetByID(id string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByIDWithPermissions(id string) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id string) error {
	// Check if it's a system role
	var role model.Role
	if err := r.db.Where("id = ?", id).First(&role).Error; err != nil {
		return err
	}
	if role.IsSystem {
		return gorm.ErrRecordNotFound // Prevent deletion of system roles
	}
	return r.db.Delete(&model.Role{}, "id = ?", id).Error
}

func (r *roleRepository) List(limit, offset int) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.Limit(limit).Offset(offset).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) AssignPermission(roleID, permissionID string) error {
	// Check if already assigned
	var count int64
	r.db.Model(&model.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count)
	if count > 0 {
		return nil // Already assigned
	}

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return err
	}
	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return err
	}

	rolePermission := &model.RolePermission{
		RoleID:       roleUUID,
		PermissionID: permissionUUID,
	}
	return r.db.Create(rolePermission).Error
}

func (r *roleRepository) RemovePermission(roleID, permissionID string) error {
	return r.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&model.RolePermission{}).Error
}

func (r *roleRepository) GetUserRoles(userID string) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.Table("roles").
		Joins("INNER JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Preload("Permissions").
		Find(&roles).Error
	return roles, err
}

func (r *roleRepository) AssignRoleToUser(userID, roleID string) error {
	// Check if already assigned
	var count int64
	r.db.Model(&model.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count)
	if count > 0 {
		return nil // Already assigned
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return err
	}

	userRole := &model.UserRole{
		UserID: userUUID,
		RoleID: roleUUID,
	}
	return r.db.Create(userRole).Error
}

func (r *roleRepository) RemoveRoleFromUser(userID, roleID string) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&model.UserRole{}).Error
}
