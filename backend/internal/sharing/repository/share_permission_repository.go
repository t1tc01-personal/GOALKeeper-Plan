package repository

import (
	"context"
	"goalkeeper-plan/internal/sharing/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SharePermissionRepository interface {
	Create(ctx context.Context, permission *model.SharePermission) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.SharePermission, error)
	GetByPageAndUser(ctx context.Context, pageID, userID uuid.UUID) (*model.SharePermission, error)
	ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*model.SharePermission, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.SharePermission, error)
	Update(ctx context.Context, permission *model.SharePermission) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByPageAndUser(ctx context.Context, pageID, userID uuid.UUID) error
}

type sharePermissionRepository struct {
	db *gorm.DB
}

func NewSharePermissionRepository(db *gorm.DB) SharePermissionRepository {
	return &sharePermissionRepository{db: db}
}

func (r *sharePermissionRepository) Create(ctx context.Context, permission *model.SharePermission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *sharePermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.SharePermission, error) {
	var permission *model.SharePermission
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return permission, nil
}

func (r *sharePermissionRepository) GetByPageAndUser(ctx context.Context, pageID, userID uuid.UUID) (*model.SharePermission, error) {
	var permission *model.SharePermission
	err := r.db.WithContext(ctx).
		Where("page_id = ? AND user_id = ?", pageID, userID).
		First(&permission).Error
	if err != nil {
		return nil, err
	}
	return permission, nil
}

func (r *sharePermissionRepository) ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*model.SharePermission, error) {
	var permissions []*model.SharePermission
	err := r.db.WithContext(ctx).
		Where("page_id = ?", pageID).
		Order("created_at ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *sharePermissionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.SharePermission, error) {
	var permissions []*model.SharePermission
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *sharePermissionRepository) Update(ctx context.Context, permission *model.SharePermission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *sharePermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.SharePermission{}).Error
}

func (r *sharePermissionRepository) DeleteByPageAndUser(ctx context.Context, pageID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("page_id = ? AND user_id = ?", pageID, userID).
		Delete(&model.SharePermission{}).Error
}
