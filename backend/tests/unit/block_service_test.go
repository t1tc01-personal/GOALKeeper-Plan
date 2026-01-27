package unit

import (
	"context"
	"errors"
	"testing"

	"goalkeeper-plan/internal/block/model"
	"goalkeeper-plan/internal/block/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockBlockRepository is a mock implementation of BlockRepository
type MockBlockRepository struct {
	mock.Mock
}

func (m *MockBlockRepository) Create(ctx context.Context, block *model.Block) error {
	args := m.Called(ctx, block)
	return args.Error(0)
}

func (m *MockBlockRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Block, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Block), args.Error(1)
}

func (m *MockBlockRepository) ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*model.Block, error) {
	args := m.Called(ctx, pageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Block), args.Error(1)
}

func (m *MockBlockRepository) ListByParentBlockID(ctx context.Context, parentBlockID *uuid.UUID) ([]*model.Block, error) {
	args := m.Called(ctx, parentBlockID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Block), args.Error(1)
}

func (m *MockBlockRepository) Update(ctx context.Context, block *model.Block) error {
	args := m.Called(ctx, block)
	return args.Error(0)
}

func (m *MockBlockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBlockRepository) Reorder(ctx context.Context, blockIDs []uuid.UUID) error {
	args := m.Called(ctx, blockIDs)
	return args.Error(0)
}

// TestBlockServiceCreateSuccess tests creating a block successfully
func TestBlockServiceCreateSuccess(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, err := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)
	require.NoError(t, err)

	ctx := context.Background()
	pageID := uuid.New()
	blockType := &model.BlockType{ID: uuid.New(), Name: "Paragraph"}
	content := "Test content"

	mockRepo.On("Create", ctx, mock.MatchedBy(func(b *model.Block) bool {
		return b.PageID == pageID && b.TypeID == blockType.ID && b.Content == &content
	})).Run(func(args mock.Arguments) {
		b := args.Get(1).(*model.Block)
		b.ID = uuid.New()
	}).Return(nil)

	// Execute
	result, err := svc.CreateBlock(ctx, pageID, blockType, &content, 0)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, pageID, result.PageID)
	assert.Equal(t, blockType.ID, result.TypeID)
	assert.Equal(t, &content, result.Content)
	mockRepo.AssertExpectations(t)
}

// TestBlockServiceCreateValidationError tests creation with invalid pageID
func TestBlockServiceCreateValidationError(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	blockType := &model.BlockType{ID: uuid.New(), Name: "Paragraph"}

	// Execute with nil pageID
	result, err := svc.CreateBlock(ctx, uuid.Nil, blockType, nil, 0)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "Create")
}

// TestBlockServiceCreateMissingBlockType tests creation without block type
func TestBlockServiceCreateMissingBlockType(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	content := "Test"

	// Execute with nil blockType - should get validation error
	result, err := svc.CreateBlock(ctx, pageID, nil, &content, 0)

	// Assert - should fail validation
	assert.Error(t, err, "expected validation error for nil blockType")
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "Create")
}

// TestBlockServiceCreateRepositoryError tests repository failure
func TestBlockServiceCreateRepositoryError(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	blockType := &model.BlockType{ID: uuid.New(), Name: "Paragraph"}

	mockRepo.On("Create", ctx, mock.Anything).Return(errors.New("database error"))

	// Execute
	result, err := svc.CreateBlock(ctx, pageID, blockType, nil, 0)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestBlockServiceGetSuccess tests retrieving a block successfully
func TestBlockServiceGetSuccess(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	blockID := uuid.New()
	expected := &model.Block{
		ID:       blockID,
		PageID:   uuid.New(),
		TypeID:   uuid.New(),
		Content:  nil,
		Rank:     0,
		Metadata: make(map[string]any),
	}

	mockRepo.On("GetByID", ctx, blockID).Return(expected, nil)

	// Execute
	result, err := svc.GetBlock(ctx, blockID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, blockID, result.ID)
	assert.Equal(t, expected.PageID, result.PageID)
	mockRepo.AssertExpectations(t)
}

// TestBlockServiceGetNotFound tests retrieving a non-existent block
func TestBlockServiceGetNotFound(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	blockID := uuid.New()

	mockRepo.On("GetByID", ctx, blockID).Return(nil, errors.New("block not found"))

	// Execute
	result, err := svc.GetBlock(ctx, blockID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestBlockServiceListByPageSuccess tests listing blocks by page
func TestBlockServiceListByPageSuccess(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	blocks := []*model.Block{
		{ID: uuid.New(), PageID: pageID, Rank: 0},
		{ID: uuid.New(), PageID: pageID, Rank: 1},
	}

	mockRepo.On("ListByPageID", ctx, pageID).Return(blocks, nil)

	// Execute
	result, err := svc.ListBlocksByPage(ctx, pageID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

// TestBlockServiceListByPageEmpty tests listing blocks from empty page
func TestBlockServiceListByPageEmpty(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()

	mockRepo.On("ListByPageID", ctx, pageID).Return([]*model.Block{}, nil)

	// Execute
	result, err := svc.ListBlocksByPage(ctx, pageID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

// TestBlockServiceListByParentSuccess tests listing child blocks
func TestBlockServiceListByParentSuccess(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	parentID := uuid.New()
	blocks := []*model.Block{
		{ID: uuid.New(), Rank: 0},
		{ID: uuid.New(), Rank: 1},
	}

	mockRepo.On("ListByParentBlockID", ctx, &parentID).Return(blocks, nil)

	// Execute
	result, err := svc.ListBlocksByParent(ctx, &parentID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
}

// TestBlockServiceUpdateSuccess tests updating a block
func TestBlockServiceUpdateSuccess(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	blockID := uuid.New()
	newContent := "Updated content"

	existing := &model.Block{
		ID:      blockID,
		Content: nil,
		Rank:    0,
	}

	mockRepo.On("GetByID", ctx, blockID).Return(existing, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(b *model.Block) bool {
		return b.ID == blockID && b.Content == &newContent
	})).Return(nil)

	// Execute
	result, err := svc.UpdateBlock(ctx, blockID, &newContent)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, &newContent, result.Content)
	mockRepo.AssertExpectations(t)
}

// TestBlockServiceUpdateNotFound tests updating a non-existent block
func TestBlockServiceUpdateNotFound(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	blockID := uuid.New()

	mockRepo.On("GetByID", ctx, blockID).Return(nil, errors.New("block not found"))

	// Execute
	result, err := svc.UpdateBlock(ctx, blockID, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "Update")
}

// TestBlockServiceDeleteSuccess tests deleting a block
func TestBlockServiceDeleteSuccess(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	blockID := uuid.New()

	mockRepo.On("Delete", ctx, blockID).Return(nil)

	// Execute
	err := svc.DeleteBlock(ctx, blockID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestBlockServiceDeleteNotFound tests deleting a non-existent block
func TestBlockServiceDeleteNotFound(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	blockID := uuid.New()

	mockRepo.On("Delete", ctx, blockID).Return(errors.New("block not found"))

	// Execute
	err := svc.DeleteBlock(ctx, blockID)

	// Assert
	assert.Error(t, err)
}

// TestBlockServiceReorderSuccess tests reordering blocks
func TestBlockServiceReorderSuccess(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	blockIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	mockRepo.On("Reorder", ctx, blockIDs).Return(nil)

	// Execute
	err := svc.ReorderBlocks(ctx, pageID, blockIDs)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestBlockServiceReorderEmpty tests reordering with empty list
func TestBlockServiceReorderEmpty(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()

	mockRepo.On("Reorder", ctx, []uuid.UUID{}).Return(nil)

	// Execute
	err := svc.ReorderBlocks(ctx, pageID, []uuid.UUID{})

	// Assert
	assert.NoError(t, err)
}

// TestBlockServiceReorderRepositoryError tests reorder with repository error
func TestBlockServiceReorderRepositoryError(t *testing.T) {
	mockRepo := new(MockBlockRepository)
	log := NewNoopLogger()
	svc, _ := service.NewBlockService(
		service.WithBlockRepository(mockRepo),
		service.WithBlockLogger(log),
	)

	ctx := context.Background()
	pageID := uuid.New()
	blockIDs := []uuid.UUID{uuid.New()}

	mockRepo.On("Reorder", ctx, blockIDs).Return(errors.New("database error"))

	// Execute
	err := svc.ReorderBlocks(ctx, pageID, blockIDs)

	// Assert
	assert.Error(t, err)
}
