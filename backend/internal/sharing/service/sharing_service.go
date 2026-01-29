package service

import (
	"context"
	"errors"
	"fmt"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/sharing/model"
	"goalkeeper-plan/internal/sharing/repository"

	"github.com/google/uuid"
)

type SharingService interface {
	GrantAccess(ctx context.Context, pageID, userID uuid.UUID, role model.SharePermissionRole) error
	RevokeAccess(ctx context.Context, pageID, userID uuid.UUID) error
	ListCollaborators(ctx context.Context, pageID uuid.UUID) ([]*model.SharePermission, error)
	HasAccess(ctx context.Context, pageID, userID uuid.UUID) (bool, error)
	CanEdit(ctx context.Context, pageID, userID uuid.UUID) (bool, error)
	GetUserPages(ctx context.Context, userID uuid.UUID) ([]*model.SharePermission, error)
}

type sharingService struct {
	permRepo repository.SharePermissionRepository
	log      logger.Logger
}

// WithSharingRepository sets the SharePermission repository
func WithSharingRepository(repo repository.SharePermissionRepository) func(*sharingService) error {
	return func(s *sharingService) error {
		s.permRepo = repo
		return nil
	}
}

// WithSharingLogger sets the logger
func WithSharingLogger(log logger.Logger) func(*sharingService) error {
	return func(s *sharingService) error {
		s.log = log
		return nil
	}
}

func NewSharingService(opts ...func(*sharingService) error) (SharingService, error) {
	s := &sharingService{}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	if s.permRepo == nil {
		return nil, errors.New("SharePermissionRepository is required")
	}

	if s.log == nil {
		return nil, errors.New("Logger is required")
	}

	return s, nil
}

func (s *sharingService) GrantAccess(ctx context.Context, pageID, userID uuid.UUID, role model.SharePermissionRole) error {
	// Validate role
	if role != model.RoleViewer && role != model.RoleEditor && role != model.RoleOwner {
		return fmt.Errorf("invalid role: %s", role)
	}

	// Check if permission already exists
	existing, err := s.permRepo.GetByPageAndUser(ctx, pageID, userID)
	if err == nil && existing != nil {
		// Update existing permission
		existing.Role = role
		err = s.permRepo.Update(ctx, existing)
		if err != nil {
			s.log.Error(fmt.Sprintf("failed to update sharing permission: %v", err))
			return err
		}
		s.log.Info(fmt.Sprintf("updated access for user %s to page %s with role %s", userID, pageID, role))
		return nil
	}

	// Create new permission
	perm := &model.SharePermission{
		ID:     uuid.New(),
		PageID: pageID,
		UserID: userID,
		Role:   role,
	}

	err = s.permRepo.Create(ctx, perm)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to create sharing permission: %v", err))
		return err
	}

	s.log.Info(fmt.Sprintf("granted %s access to user %s for page %s", role, userID, pageID))
	return nil
}

func (s *sharingService) RevokeAccess(ctx context.Context, pageID, userID uuid.UUID) error {
	err := s.permRepo.DeleteByPageAndUser(ctx, pageID, userID)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to revoke access: %v", err))
		return err
	}

	s.log.Info(fmt.Sprintf("revoked access for user %s to page %s", userID, pageID))
	return nil
}

func (s *sharingService) ListCollaborators(ctx context.Context, pageID uuid.UUID) ([]*model.SharePermission, error) {
	perms, err := s.permRepo.ListByPageID(ctx, pageID)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to list collaborators: %v", err))
		return nil, err
	}

	return perms, nil
}

func (s *sharingService) HasAccess(ctx context.Context, pageID, userID uuid.UUID) (bool, error) {
	perm, err := s.permRepo.GetByPageAndUser(ctx, pageID, userID)
	if err != nil {
		// If not found, return false (not an error)
		return false, nil
	}

	return perm != nil, nil
}

func (s *sharingService) CanEdit(ctx context.Context, pageID, userID uuid.UUID) (bool, error) {
	perm, err := s.permRepo.GetByPageAndUser(ctx, pageID, userID)
	if err != nil || perm == nil {
		return false, nil
	}

	// Owner and Editor roles can edit
	return perm.Role == model.RoleEditor || perm.Role == model.RoleOwner, nil
}

func (s *sharingService) GetUserPages(ctx context.Context, userID uuid.UUID) ([]*model.SharePermission, error) {
	perms, err := s.permRepo.ListByUserID(ctx, userID)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to get user pages: %v", err))
		return nil, err
	}

	return perms, nil
}


