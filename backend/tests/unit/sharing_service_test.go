package unit

import (
	"context"
	"errors"
	"testing"

	"goalkeeper-plan/internal/workspace/model"
	"goalkeeper-plan/internal/workspace/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSharePermissionRepository is a mock implementation of SharePermissionRepository
type MockSharePermissionRepository struct {
	mock.Mock
}

func (m *MockSharePermissionRepository) Create(ctx context.Context, perm *model.SharePermission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockSharePermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.SharePermission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SharePermission), args.Error(1)
}

func (m *MockSharePermissionRepository) GetByPageAndUser(ctx context.Context, pageID, userID uuid.UUID) (*model.SharePermission, error) {
	args := m.Called(ctx, pageID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SharePermission), args.Error(1)
}

func (m *MockSharePermissionRepository) ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*model.SharePermission, error) {
	args := m.Called(ctx, pageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.SharePermission), args.Error(1)
}

func (m *MockSharePermissionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.SharePermission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.SharePermission), args.Error(1)
}

func (m *MockSharePermissionRepository) Update(ctx context.Context, perm *model.SharePermission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockSharePermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSharePermissionRepository) DeleteByPageAndUser(ctx context.Context, pageID, userID uuid.UUID) error {
	args := m.Called(ctx, pageID, userID)
	return args.Error(0)
}

// TestSharingServiceGrantAccessNewSuccess tests granting access for the first time
func TestSharingServiceGrantAccessNewSuccess(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, err := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)
	require.NoError(t, err)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	// Expect permission doesn't exist
	mockRepo.On("GetByPageAndUser", ctx, pageID, userID).Return(nil, errors.New("not found"))

	// Expect creation
	mockRepo.On("Create", ctx, mock.MatchedBy(func(perm *model.SharePermission) bool {
		return perm.PageID == pageID && perm.UserID == userID && perm.Role == model.RoleEditor
	})).Return(nil)

	// Execute
	err = svc.GrantAccess(ctx, pageID, userID, model.RoleEditor)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestSharingServiceGrantAccessUpdate tests updating existing access permission
func TestSharingServiceGrantAccessUpdate(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	existing := &model.SharePermission{
		ID:     uuid.New(),
		PageID: pageID,
		UserID: userID,
		Role:   model.RoleViewer,
	}

	// Expect permission exists
	mockRepo.On("GetByPageAndUser", ctx, pageID, userID).Return(existing, nil)

	// Expect update to Editor
	mockRepo.On("Update", ctx, mock.MatchedBy(func(perm *model.SharePermission) bool {
		return perm.ID == existing.ID && perm.Role == model.RoleEditor
	})).Return(nil)

	// Execute
	err := svc.GrantAccess(ctx, pageID, userID, model.RoleEditor)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestSharingServiceGrantAccessInvalidRole tests granting with invalid role
func TestSharingServiceGrantAccessInvalidRole(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	// Execute with invalid role
	err := svc.GrantAccess(ctx, pageID, userID, "invalid_role")

	// Assert
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "GetByPageAndUser")
}

// TestSharingServiceGrantAccessCreateError tests creation failure
func TestSharingServiceGrantAccessCreateError(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	mockRepo.On("GetByPageAndUser", ctx, pageID, userID).Return(nil, errors.New("not found"))
	mockRepo.On("Create", ctx, mock.Anything).Return(errors.New("database error"))

	// Execute
	err := svc.GrantAccess(ctx, pageID, userID, model.RoleViewer)

	// Assert
	assert.Error(t, err)
}

// TestSharingServiceRevokeAccessSuccess tests revoking access successfully
func TestSharingServiceRevokeAccessSuccess(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	mockRepo.On("DeleteByPageAndUser", ctx, pageID, userID).Return(nil)

	// Execute
	err := svc.RevokeAccess(ctx, pageID, userID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestSharingServiceRevokeAccessNotFound tests revoking access that doesn't exist
func TestSharingServiceRevokeAccessNotFound(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	mockRepo.On("DeleteByPageAndUser", ctx, pageID, userID).Return(errors.New("permission not found"))

	// Execute
	err := svc.RevokeAccess(ctx, pageID, userID)

	// Assert
	assert.Error(t, err)
}

// TestSharingServiceListCollaboratorsSuccess tests listing collaborators
func TestSharingServiceListCollaboratorsSuccess(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()

	permissions := []*model.SharePermission{
		{ID: uuid.New(), PageID: pageID, UserID: uuid.New(), Role: model.RoleEditor},
		{ID: uuid.New(), PageID: pageID, UserID: uuid.New(), Role: model.RoleViewer},
	}

	mockRepo.On("ListByPageID", ctx, pageID).Return(permissions, nil)

	// Execute
	result, err := svc.ListCollaborators(ctx, pageID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, model.RoleEditor, result[0].Role)
	assert.Equal(t, model.RoleViewer, result[1].Role)
}

// TestSharingServiceListCollaboratorsEmpty tests listing collaborators when none exist
func TestSharingServiceListCollaboratorsEmpty(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()

	mockRepo.On("ListByPageID", ctx, pageID).Return([]*model.SharePermission{}, nil)

	// Execute
	result, err := svc.ListCollaborators(ctx, pageID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

// TestSharingServiceHasAccessYes tests checking access when user has access
func TestSharingServiceHasAccessYes(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	perm := &model.SharePermission{
		ID:     uuid.New(),
		PageID: pageID,
		UserID: userID,
		Role:   model.RoleViewer,
	}

	mockRepo.On("GetByPageAndUser", ctx, pageID, userID).Return(perm, nil)

	// Execute
	hasAccess, err := svc.HasAccess(ctx, pageID, userID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, hasAccess)
}

// TestSharingServiceHasAccessNo tests checking access when user doesn't have access
func TestSharingServiceHasAccessNo(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	mockRepo.On("GetByPageAndUser", ctx, pageID, userID).Return(nil, errors.New("not found"))

	// Execute
	hasAccess, err := svc.HasAccess(ctx, pageID, userID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, hasAccess)
}

// TestSharingServiceCanEditAsEditor tests edit permission for editor
func TestSharingServiceCanEditAsEditor(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	perm := &model.SharePermission{
		ID:     uuid.New(),
		PageID: pageID,
		UserID: userID,
		Role:   model.RoleEditor,
	}

	mockRepo.On("GetByPageAndUser", ctx, pageID, userID).Return(perm, nil)

	// Execute
	canEdit, err := svc.CanEdit(ctx, pageID, userID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, canEdit)
}

// TestSharingServiceCanEditAsOwner tests edit permission for owner
func TestSharingServiceCanEditAsOwner(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	perm := &model.SharePermission{
		ID:     uuid.New(),
		PageID: pageID,
		UserID: userID,
		Role:   model.RoleOwner,
	}

	mockRepo.On("GetByPageAndUser", ctx, pageID, userID).Return(perm, nil)

	// Execute
	canEdit, err := svc.CanEdit(ctx, pageID, userID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, canEdit)
}

// TestSharingServiceCanEditAsViewer tests edit permission for viewer
func TestSharingServiceCanEditAsViewer(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	perm := &model.SharePermission{
		ID:     uuid.New(),
		PageID: pageID,
		UserID: userID,
		Role:   model.RoleViewer,
	}

	mockRepo.On("GetByPageAndUser", ctx, pageID, userID).Return(perm, nil)

	// Execute
	canEdit, err := svc.CanEdit(ctx, pageID, userID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, canEdit)
}

// TestSharingServiceCanEditNoAccess tests edit permission when no access
func TestSharingServiceCanEditNoAccess(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	userID := uuid.New()

	mockRepo.On("GetByPageAndUser", ctx, pageID, userID).Return(nil, errors.New("not found"))

	// Execute
	canEdit, err := svc.CanEdit(ctx, pageID, userID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, canEdit)
}

// TestSharingServiceGetUserPagesSuccess tests getting pages for a user
func TestSharingServiceGetUserPagesSuccess(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	userID := uuid.New()

	permissions := []*model.SharePermission{
		{ID: uuid.New(), PageID: uuid.New(), UserID: userID, Role: model.RoleEditor},
		{ID: uuid.New(), PageID: uuid.New(), UserID: userID, Role: model.RoleViewer},
	}

	mockRepo.On("ListByUserID", ctx, userID).Return(permissions, nil)

	// Execute
	result, err := svc.GetUserPages(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
}

// TestSharingServiceGetUserPagesEmpty tests getting pages for user with no access
func TestSharingServiceGetUserPagesEmpty(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)
	log := NewNoopLogger()
	svc, _ := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
		service.WithSharingLogger(log),
	)

	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("ListByUserID", ctx, userID).Return([]*model.SharePermission{}, nil)

	// Execute
	result, err := svc.GetUserPages(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

// TestSharingServiceNewNoRepository tests initialization without repository
func TestSharingServiceNewNoRepository(t *testing.T) {
	log := NewNoopLogger()

	// Execute
	svc, err := service.NewSharingService(
		service.WithSharingLogger(log),
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, svc)
}

// TestSharingServiceNewNoLogger tests initialization without logger
func TestSharingServiceNewNoLogger(t *testing.T) {
	mockRepo := new(MockSharePermissionRepository)

	// Execute
	svc, err := service.NewSharingService(
		service.WithSharingRepository(mockRepo),
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, svc)
}
