package repository

import (
	"context"
	"goalkeeper-plan/internal/block/dto"
	"goalkeeper-plan/internal/page/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PageRepository interface {
	Create(ctx context.Context, page *model.Page) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Page, error)
	ListByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*model.Page, error)
	ListByWorkspaceIDWithPagination(ctx context.Context, workspaceID uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Page, *dto.PaginationMeta, error)
	ListByParentPageID(ctx context.Context, parentPageID *uuid.UUID) ([]*model.Page, error)
	Update(ctx context.Context, page *model.Page) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetHierarchy(ctx context.Context, workspaceID uuid.UUID) ([]*model.Page, error)
}

type pageRepository struct {
	db *gorm.DB
}

func NewPageRepository(db *gorm.DB) PageRepository {
	return &pageRepository{db: db}
}

func (r *pageRepository) Create(ctx context.Context, page *model.Page) error {
	return r.db.WithContext(ctx).Create(page).Error
}

func (r *pageRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Page, error) {
	var page *model.Page
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&page).Error
	if err != nil {
		return nil, err
	}
	return page, nil
}

func (r *pageRepository) ListByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*model.Page, error) {
	var pages []*model.Page
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Order("created_at DESC").
		Find(&pages).Error
	if err != nil {
		return nil, err
	}
	return pages, nil
}

func (r *pageRepository) ListByWorkspaceIDWithPagination(ctx context.Context, workspaceID uuid.UUID, pagReq *dto.PaginationRequest) ([]*model.Page, *dto.PaginationMeta, error) {
	var pages []*model.Page
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).
		Model(&model.Page{}).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Order("created_at DESC").
		Limit(pagReq.GetLimit()).
		Offset(pagReq.GetOffset()).
		Find(&pages).Error
	if err != nil {
		return nil, nil, err
	}

	meta := &dto.PaginationMeta{
		Total:   int(total),
		Limit:   pagReq.GetLimit(),
		Offset:  pagReq.GetOffset(),
		HasMore: int64(pagReq.GetOffset()+pagReq.GetLimit()) < total,
	}

	return pages, meta, nil
}

func (r *pageRepository) ListByParentPageID(ctx context.Context, parentPageID *uuid.UUID) ([]*model.Page, error) {
	var pages []*model.Page
	if parentPageID == nil {
		// Get root pages (pages without a parent)
		err := r.db.WithContext(ctx).
			Where("parent_page_id IS NULL AND deleted_at IS NULL").
			Order("created_at DESC").
			Find(&pages).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := r.db.WithContext(ctx).
			Where("parent_page_id = ? AND deleted_at IS NULL", parentPageID).
			Order("created_at DESC").
			Find(&pages).Error
		if err != nil {
			return nil, err
		}
	}
	return pages, nil
}

func (r *pageRepository) Update(ctx context.Context, page *model.Page) error {
	return r.db.WithContext(ctx).Save(page).Error
}

func (r *pageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Page{}).Error
}

// GetHierarchy returns all pages for a workspace with parent-child relationships
func (r *pageRepository) GetHierarchy(ctx context.Context, workspaceID uuid.UUID) ([]*model.Page, error) {
	var pages []*model.Page
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Order("parent_page_id NULLS FIRST, created_at DESC").
		Find(&pages).Error
	if err != nil {
		return nil, err
	}
	return pages, nil
}
