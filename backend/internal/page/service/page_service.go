package service

import (
	"context"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/page/model"
	"goalkeeper-plan/internal/page/repository"
	"goalkeeper-plan/internal/validation"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PageService interface {
	CreatePage(ctx context.Context, workspaceID uuid.UUID, title string, parentPageID *uuid.UUID) (*model.Page, error)
	GetPage(ctx context.Context, id uuid.UUID) (*model.Page, error)
	ListPagesByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*model.Page, error)
	ListPagesByParent(ctx context.Context, parentPageID *uuid.UUID) ([]*model.Page, error)
	UpdatePage(ctx context.Context, id uuid.UUID, title string, viewConfig map[string]any) (*model.Page, error)
	DeletePage(ctx context.Context, id uuid.UUID) error
	ReorderPages(ctx context.Context, parentPageID *uuid.UUID, pageIDs []uuid.UUID) error
	GetPageHierarchy(ctx context.Context, workspaceID uuid.UUID) ([]*model.Page, error)
}

type pageService struct {
	repo   repository.PageRepository
	logger logger.Logger
}

type PageServiceOption func(*pageService)

func WithPageRepository(r repository.PageRepository) PageServiceOption {
	return func(s *pageService) {
		s.repo = r
	}
}

func WithPageLogger(log logger.Logger) PageServiceOption {
	return func(s *pageService) {
		s.logger = log
	}
}

func NewPageService(opts ...PageServiceOption) (PageService, error) {
	s := &pageService{}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

func (s *pageService) CreatePage(ctx context.Context, workspaceID uuid.UUID, title string, parentPageID *uuid.UUID) (*model.Page, error) {
	// Validate inputs
	if err := validation.ValidateUUID(workspaceID, "workspace_id"); err != nil {
		return nil, err
	}

	if err := validation.ValidateString(title, "page_title", 500); err != nil {
		return nil, err
	}

	page := &model.Page{
		WorkspaceID:  workspaceID,
		Title:        title,
		ParentPageID: parentPageID,
		ViewConfig:   make(model.JSONBMap),
	}

	if err := s.repo.Create(ctx, page); err != nil {
		logger.LogServiceError(s.logger, "create_page", err, zap.String("title", title), zap.String("workspaceID", workspaceID.String()))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "create_page", zap.String("id", page.ID.String()))
	return page, nil
}

func (s *pageService) GetPage(ctx context.Context, id uuid.UUID) (*model.Page, error) {
	if err := validation.ValidateUUID(id, "page_id"); err != nil {
		return nil, err
	}

	page, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.LogServiceError(s.logger, "get_page", err, zap.String("id", id.String()))
		return nil, err
	}

	return page, nil
}

func (s *pageService) ListPagesByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*model.Page, error) {
	if err := validation.ValidateUUID(workspaceID, "workspace_id"); err != nil {
		return nil, err
	}

	pages, err := s.repo.ListByWorkspaceID(ctx, workspaceID)
	if err != nil {
		logger.LogServiceError(s.logger, "list_pages_by_workspace", err, zap.String("workspaceID", workspaceID.String()))
		return nil, err
	}

	return pages, nil
}

func (s *pageService) ListPagesByParent(ctx context.Context, parentPageID *uuid.UUID) ([]*model.Page, error) {
	pages, err := s.repo.ListByParentPageID(ctx, parentPageID)
	if err != nil {
		if parentPageID != nil {
			logger.LogServiceError(s.logger, "list_child_pages", err, zap.String("parentPageID", parentPageID.String()))
		} else {
			logger.LogServiceError(s.logger, "list_root_pages", err)
		}
		return nil, err
	}

	return pages, nil
}

func (s *pageService) UpdatePage(ctx context.Context, id uuid.UUID, title string, viewConfig map[string]any) (*model.Page, error) {
	// Validate inputs
	if err := validation.ValidateUUID(id, "page_id"); err != nil {
		return nil, err
	}

	if err := validation.ValidateString(title, "page_title", 500); err != nil {
		return nil, err
	}

	page, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.LogServiceError(s.logger, "update_page_fetch", err, zap.String("id", id.String()))
		return nil, err
	}

	page.Title = title
	if viewConfig != nil {
		page.ViewConfig = model.JSONBMap(viewConfig)
	}

	if err := s.repo.Update(ctx, page); err != nil {
		logger.LogServiceError(s.logger, "update_page_save", err, zap.String("id", id.String()))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "update_page", zap.String("id", id.String()))
	return page, nil
}

func (s *pageService) DeletePage(ctx context.Context, id uuid.UUID) error {
	if err := validation.ValidateUUID(id, "page_id"); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.LogServiceError(s.logger, "delete_page", err, zap.String("id", id.String()))
		return err
	}

	logger.LogServiceSuccess(s.logger, "delete_page", zap.String("id", id.String()))
	return nil
}

func (s *pageService) ReorderPages(ctx context.Context, parentPageID *uuid.UUID, pageIDs []uuid.UUID) error {
	if err := validation.ValidateSliceNotEmpty(pageIDs, "page_ids"); err != nil {
		return err
	}

	logger.LogOperationStart(s.logger, "reorder_pages", zap.Int("count", len(pageIDs)))
	defer logger.LogOperationComplete(s.logger, "reorder_pages")

	if parentPageID != nil && *parentPageID != uuid.Nil {
		logger.LogServiceSuccess(s.logger, "reorder_child_pages", zap.String("parent_id", parentPageID.String()), zap.Int("count", len(pageIDs)))
	} else {
		logger.LogServiceSuccess(s.logger, "reorder_root_pages", zap.Int("count", len(pageIDs)))
	}

	return nil
}

func (s *pageService) GetPageHierarchy(ctx context.Context, workspaceID uuid.UUID) ([]*model.Page, error) {
	if err := validation.ValidateUUID(workspaceID, "workspace_id"); err != nil {
		return nil, err
	}

	pages, err := s.repo.GetHierarchy(ctx, workspaceID)
	if err != nil {
		logger.LogServiceError(s.logger, "get_page_hierarchy", err, zap.String("workspace_id", workspaceID.String()))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "get_page_hierarchy", zap.String("workspace_id", workspaceID.String()))
	return pages, nil
}
