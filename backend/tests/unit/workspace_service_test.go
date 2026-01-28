package unit

import (
	"context"
	"errors"
	"testing"

	"goalkeeper-plan/internal/block/dto"
	"goalkeeper-plan/internal/workspace/model"
	"goalkeeper-plan/internal/workspace/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockWorkspaceRepository is a mock implementation of WorkspaceRepository
type MockWorkspaceRepository struct {
	mock.Mock
}

func (m *MockWorkspaceRepository) Create(ctx context.Context, ws *model.Workspace) error {
	args := m.Called(ctx, ws)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Workspace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Workspace), args.Error(1)
}

func (m *MockWorkspaceRepository) List(ctx context.Context) ([]*model.Workspace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Workspace), args.Error(1)
}

func (m *MockWorkspaceRepository) ListWithPagination(ctx context.Context, pagReq *dto.PaginationRequest) ([]*model.Workspace, *dto.PaginationMeta, error) {
	args := m.Called(ctx, pagReq)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var meta *dto.PaginationMeta
	if args.Get(1) != nil {
		meta = args.Get(1).(*dto.PaginationMeta)
	}
	return args.Get(0).([]*model.Workspace), meta, args.Error(2)
}

func (m *MockWorkspaceRepository) Update(ctx context.Context, ws *model.Workspace) error {
	args := m.Called(ctx, ws)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestWorkspaceServiceCreateSuccess tests creating a workspace successfully
func TestWorkspaceServiceCreateSuccess(t *testing.T) {
	// Setup
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, err := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)
	require.NoError(t, err)
	require.NotNil(t, svc)

	ctx := context.Background()
	ownerID := uuid.New()
	name := "Test Workspace"
	desc := "Test description"

	// Expect
	mockRepo.On("Create", ctx, mock.MatchedBy(func(ws *model.Workspace) bool {
		return ws.OwnerID == ownerID && ws.Name == name && ws.Description == &desc
	})).Run(func(args mock.Arguments) {
		ws := args.Get(1).(*model.Workspace)
		ws.ID = uuid.New() // Simulate ID generation by repo
	}).Return(nil)

	// Execute
	result, err := svc.CreateWorkspace(ctx, ownerID, name, &desc)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, name, result.Name)
	assert.Equal(t, &desc, result.Description)
	assert.NotEqual(t, uuid.Nil, result.ID)
	mockRepo.AssertExpectations(t)
}

// TestWorkspaceServiceCreateValidationError tests validation failure on empty name
func TestWorkspaceServiceCreateValidationError(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()
	ownerID := uuid.New()

	// Execute with empty name
	result, err := svc.CreateWorkspace(ctx, ownerID, "", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "Create")
}

// TestWorkspaceServiceCreateRepositoryError tests repository failure
func TestWorkspaceServiceCreateRepositoryError(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()
	ownerID := uuid.New()
	name := "Test Workspace"

	// Expect repository error
	mockRepo.On("Create", ctx, mock.Anything).Return(errors.New("database error"))

	// Execute
	result, err := svc.CreateWorkspace(ctx, ownerID, name, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database error", err.Error())
	mockRepo.AssertExpectations(t)
}

// TestWorkspaceServiceGetSuccess tests retrieving a workspace successfully
func TestWorkspaceServiceGetSuccess(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()
	id := uuid.New()
	expected := &model.Workspace{
		ID:   id,
		Name: "Test Workspace",
	}

	mockRepo.On("GetByID", ctx, id).Return(expected, nil)

	// Execute
	result, err := svc.GetWorkspace(ctx, id)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

// TestWorkspaceServiceGetNotFound tests retrieving a non-existent workspace
func TestWorkspaceServiceGetNotFound(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()
	id := uuid.New()

	mockRepo.On("GetByID", ctx, id).Return(nil, errors.New("workspace not found"))

	// Execute
	result, err := svc.GetWorkspace(ctx, id)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// TestWorkspaceServiceGetInvalidID tests validation of invalid UUID
func TestWorkspaceServiceGetInvalidID(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()

	// Execute with nil UUID
	result, err := svc.GetWorkspace(ctx, uuid.Nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "GetByID")
}

// TestWorkspaceServiceListSuccess tests listing workspaces successfully
func TestWorkspaceServiceListSuccess(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()
	expected := []*model.Workspace{
		{ID: uuid.New(), Name: "Workspace 1"},
		{ID: uuid.New(), Name: "Workspace 2"},
	}

	mockRepo.On("List", ctx).Return(expected, nil)

	// Execute
	result, err := svc.ListWorkspaces(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, expected[0].Name, result[0].Name)
	assert.Equal(t, expected[1].Name, result[1].Name)
	mockRepo.AssertExpectations(t)
}

// TestWorkspaceServiceListEmpty tests listing when no workspaces exist
func TestWorkspaceServiceListEmpty(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()

	mockRepo.On("List", ctx).Return([]*model.Workspace{}, nil)

	// Execute
	result, err := svc.ListWorkspaces(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
	mockRepo.AssertExpectations(t)
}

// TestWorkspaceServiceUpdateSuccess tests updating a workspace successfully
func TestWorkspaceServiceUpdateSuccess(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()
	id := uuid.New()
	newName := "Updated Name"
	newDesc := "Updated Description"

	existing := &model.Workspace{
		ID:   id,
		Name: "Old Name",
	}

	mockRepo.On("GetByID", ctx, id).Return(existing, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(ws *model.Workspace) bool {
		return ws.ID == id && ws.Name == newName && ws.Description == &newDesc
	})).Return(nil)

	// Execute
	result, err := svc.UpdateWorkspace(ctx, id, newName, &newDesc)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newName, result.Name)
	assert.Equal(t, &newDesc, result.Description)
	mockRepo.AssertExpectations(t)
}

// TestWorkspaceServiceUpdateNotFound tests updating a non-existent workspace
func TestWorkspaceServiceUpdateNotFound(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()
	id := uuid.New()

	mockRepo.On("GetByID", ctx, id).Return(nil, errors.New("workspace not found"))

	// Execute
	result, err := svc.UpdateWorkspace(ctx, id, "New Name", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "Update")
}

// TestWorkspaceServiceUpdateValidationError tests validation failure
func TestWorkspaceServiceUpdateValidationError(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()

	// Execute with invalid ID
	result, err := svc.UpdateWorkspace(ctx, uuid.Nil, "New Name", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "GetByID")
}

// TestWorkspaceServiceDeleteSuccess tests deleting a workspace successfully
func TestWorkspaceServiceDeleteSuccess(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()
	id := uuid.New()

	mockRepo.On("Delete", ctx, id).Return(nil)

	// Execute
	err := svc.DeleteWorkspace(ctx, id)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestWorkspaceServiceDeleteNotFound tests deleting a non-existent workspace
func TestWorkspaceServiceDeleteNotFound(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()
	id := uuid.New()

	mockRepo.On("Delete", ctx, id).Return(errors.New("workspace not found"))

	// Execute
	err := svc.DeleteWorkspace(ctx, id)

	// Assert
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

// TestWorkspaceServiceDeleteValidationError tests validation failure
func TestWorkspaceServiceDeleteValidationError(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	log := NewNoopLogger()
	svc, _ := service.NewWorkspaceService(
		service.WithWorkspaceRepository(mockRepo),
		service.WithLogger(log),
	)

	ctx := context.Background()

	// Execute with invalid ID
	err := svc.DeleteWorkspace(ctx, uuid.Nil)

	// Assert
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "Delete")
}
