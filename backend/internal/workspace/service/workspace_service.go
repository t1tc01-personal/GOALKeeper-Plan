package service

import (
	"context"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/validation"
	"goalkeeper-plan/internal/workspace/model"
	"goalkeeper-plan/internal/workspace/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, ownerID uuid.UUID, name string, description *string) (*model.Workspace, error)
	GetWorkspace(ctx context.Context, id uuid.UUID) (*model.Workspace, error)
	ListWorkspaces(ctx context.Context) ([]*model.Workspace, error)
	UpdateWorkspace(ctx context.Context, id uuid.UUID, name string, description *string) (*model.Workspace, error)
	DeleteWorkspace(ctx context.Context, id uuid.UUID) error
}

type workspaceService struct {
	repo   repository.WorkspaceRepository
	logger logger.Logger
}

type WorkspaceServiceOption func(*workspaceService)

func WithWorkspaceRepository(r repository.WorkspaceRepository) WorkspaceServiceOption {
	return func(s *workspaceService) {
		s.repo = r
	}
}

func WithLogger(log logger.Logger) WorkspaceServiceOption {
	return func(s *workspaceService) {
		s.logger = log
	}
}

func NewWorkspaceService(opts ...WorkspaceServiceOption) (WorkspaceService, error) {
	s := &workspaceService{}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

func (s *workspaceService) CreateWorkspace(ctx context.Context, ownerID uuid.UUID, name string, description *string) (*model.Workspace, error) {
	// Validate input
	if err := validation.ValidateUUID(ownerID, "owner_id"); err != nil {
		return nil, err
	}
	if err := validation.ValidateString(name, "workspace_name", 255); err != nil {
		return nil, err
	}

	ws := &model.Workspace{
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
	}

	if err := s.repo.Create(ctx, ws); err != nil {
		logger.LogServiceError(s.logger, "create_workspace", err, zap.String("name", name))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "create_workspace", zap.String("id", ws.ID.String()))
	return ws, nil
}

func (s *workspaceService) GetWorkspace(ctx context.Context, id uuid.UUID) (*model.Workspace, error) {
	if err := validation.ValidateUUID(id, "workspace_id"); err != nil {
		return nil, err
	}

	ws, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.LogServiceError(s.logger, "get_workspace", err, zap.String("id", id.String()))
		return nil, err
	}

	return ws, nil
}

func (s *workspaceService) ListWorkspaces(ctx context.Context) ([]*model.Workspace, error) {
	workspaces, err := s.repo.List(ctx)
	if err != nil {
		logger.LogServiceError(s.logger, "list_workspaces", err)
		return nil, err
	}

	return workspaces, nil
}

func (s *workspaceService) UpdateWorkspace(ctx context.Context, id uuid.UUID, name string, description *string) (*model.Workspace, error) {
	// Validate inputs
	if err := validation.ValidateUUID(id, "workspace_id"); err != nil {
		return nil, err
	}

	if err := validation.ValidateString(name, "workspace_name", 255); err != nil {
		return nil, err
	}

	ws, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.LogServiceError(s.logger, "update_workspace_fetch", err, zap.String("id", id.String()))
		return nil, err
	}

	ws.Name = name
	ws.Description = description

	if err := s.repo.Update(ctx, ws); err != nil {
		logger.LogServiceError(s.logger, "update_workspace_save", err, zap.String("id", id.String()))
		return nil, err
	}

	logger.LogServiceSuccess(s.logger, "update_workspace", zap.String("id", id.String()))
	return ws, nil
}

func (s *workspaceService) DeleteWorkspace(ctx context.Context, id uuid.UUID) error {
	if err := validation.ValidateUUID(id, "workspace_id"); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.LogServiceError(s.logger, "delete_workspace", err, zap.String("id", id.String()))
		return err
	}

	logger.LogServiceSuccess(s.logger, "delete_workspace", zap.String("id", id.String()))
	return nil
}
