package repository

import (
	"context"
	"goalkeeper-plan/internal/block/dto"
	"goalkeeper-plan/internal/workspace/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *model.Workspace) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Workspace, error)
	List(ctx context.Context) ([]*model.Workspace, error)
	ListWithPagination(ctx context.Context, pagReq *dto.PaginationRequest) ([]*model.Workspace, *dto.PaginationMeta, error)
	Update(ctx context.Context, workspace *model.Workspace) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type workspaceRepository struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) WorkspaceRepository {
	return &workspaceRepository{db: db}
}

func (r *workspaceRepository) Create(ctx context.Context, workspace *model.Workspace) error {
	return r.db.WithContext(ctx).Create(workspace).Error
}

func (r *workspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Workspace, error) {
	var ws *model.Workspace
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&ws).Error
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func (r *workspaceRepository) List(ctx context.Context) ([]*model.Workspace, error) {
	var workspaces []*model.Workspace
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&workspaces).Error
	if err != nil {
		return nil, err
	}
	return workspaces, nil
}

func (r *workspaceRepository) ListWithPagination(ctx context.Context, pagReq *dto.PaginationRequest) ([]*model.Workspace, *dto.PaginationMeta, error) {
	var workspaces []*model.Workspace
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).
		Model(&model.Workspace{}).
		Where("deleted_at IS NULL").
		Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(pagReq.GetLimit()).
		Offset(pagReq.GetOffset()).
		Find(&workspaces).Error
	if err != nil {
		return nil, nil, err
	}

	meta := &dto.PaginationMeta{
		Total:   int(total),
		Limit:   pagReq.GetLimit(),
		Offset:  pagReq.GetOffset(),
		HasMore: int64(pagReq.GetOffset()+pagReq.GetLimit()) < total,
	}

	return workspaces, meta, nil
}

func (r *workspaceRepository) Update(ctx context.Context, workspace *model.Workspace) error {
	return r.db.WithContext(ctx).Save(workspace).Error
}

func (r *workspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Workspace{}).Error
}
