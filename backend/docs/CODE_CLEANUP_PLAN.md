# Code Cleanup & Refactoring Plan (T045)

**Status**: In Progress
**Goal**: Improve code quality, maintainability, and reduce duplication
**Target**: 10-15% reduction in lines of code, improved test coverage
**Modules**: Workspace, Page, Block, RBAC, API handlers

---

## Overview

This document outlines refactoring opportunities identified in the MVP codebase. The focus is on:

1. **Reduce Duplication**: Eliminate repeated patterns in services, controllers, and error handling
2. **Improve Readability**: Consolidate validation logic, improve naming
3. **Enhance Maintainability**: Extract common patterns into utilities
4. **Strengthen Error Handling**: Leverage existing AppError type consistently
5. **Optimize Dependencies**: Clean up unnecessary imports and fields

---

## Current Code Patterns & Issues

### Issue 1: Repetitive Logging in Services

**Current Pattern** (found in multiple services):
```go
if s.logger != nil {
  s.logger.Error("failed to create workspace", zap.Error(err), zap.String("name", name))
}
return nil, fmt.Errorf("failed to create workspace: %w", err)
```

**Problem**:
- Repeated `if s.logger != nil` checks
- Error wrapping always uses same pattern: `fmt.Errorf("failed to ...: %w", err)`
- No structured error types being used
- Logger field might be nil (defensive programming, not guaranteed)

**Solution**:
- Create helper method: `LogError(operationName, err, fields...)`
- Make logger mandatory (not optional pointer)
- Return structured AppError types

### Issue 2: Inconsistent Validation Logic

**Current Pattern**:
```go
if id == uuid.Nil {
  return nil, fmt.Errorf("workspace ID cannot be nil")
}

if name == "" {
  return nil, fmt.Errorf("workspace name cannot be empty")
}
```

**Problem**:
- Validation scattered across multiple methods
- String error messages instead of typed errors
- Repeated UUID nil checks
- No consistent error codes

**Solution**:
- Create validation package with reusable validators
- Use AppError with consistent error codes
- Create helper: `ValidateUUID(id, "workspace_id")`

### Issue 3: Repetitive Dependency Injection Options

**Current Pattern** (multiple services):
```go
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
```

**Problem**:
- Every service repeats this exact pattern
- Boilerplate code, ~20 lines per service
- Pattern doesn't validate required fields

**Solution**:
- Create generic DI helper or use struct embedding
- Simplify constructor validation
- Reusable pattern across services

### Issue 4: Missing Error Type Consistency

**Current Pattern**:
- Services return `fmt.Errorf("...")`  
- Controllers don't convert to AppError consistently
- No error code tracking for monitoring

**Solution**:
- All services return AppError
- Controllers simply convert and respond
- Structured error codes for tracking

### Issue 5: Controller Boilerplate

**Current Pattern** (every controller method):
```go
func (c *WorkspaceController) Create(ctx *gin.Context) {
  var req CreateWorkspaceRequest
  if err := ctx.ShouldBindJSON(&req); err != nil {
    ctx.JSON(400, gin.H{"error": err.Error()})
    return
  }

  ws, err := c.service.CreateWorkspace(ctx.Context, req.Name, req.Description)
  if err != nil {
    ctx.JSON(500, gin.H{"error": err.Error()})
    return
  }

  ctx.JSON(201, ws)
}
```

**Problem**:
- Repeated error handling logic
- No structured error responses
- Manual status code selection
- No request validation framework

**Solution**:
- Create controller base with error handling middleware
- Use request validators (struct tags)
- Leverage AppError.GetHTTPStatusCode()

---

## Refactoring Tasks

### Task 1: Create Validation Package

**File**: `internal/validation/validation.go`

**Contents**:
```go
package validation

import (
  "github.com/google/uuid"
  "goalkeeper-plan/internal/errors"
)

// ValidateUUID validates UUID is not nil
func ValidateUUID(id uuid.UUID, fieldName string) *errors.AppError {
  if id == uuid.Nil {
    return &errors.AppError{
      Type: errors.ErrorTypeValidation,
      Code: "INVALID_UUID",
      Message: fieldName + " cannot be nil",
    }
  }
  return nil
}

// ValidateString validates non-empty string
func ValidateString(value string, fieldName string, maxLength ...int) *errors.AppError {
  if value == "" {
    return &errors.AppError{
      Type: errors.ErrorTypeValidation,
      Code: "EMPTY_STRING",
      Message: fieldName + " cannot be empty",
    }
  }
  
  if len(maxLength) > 0 && len(value) > maxLength[0] {
    return &errors.AppError{
      Type: errors.ErrorTypeValidation,
      Code: "STRING_TOO_LONG",
      Message: fieldName + " exceeds max length",
    }
  }
  
  return nil
}

// ValidateRequired validates not nil/empty
func ValidateRequired(value any, fieldName string) *errors.AppError {
  if value == nil {
    return &errors.AppError{
      Type: errors.ErrorTypeValidation,
      Code: "REQUIRED_FIELD",
      Message: fieldName + " is required",
    }
  }
  return nil
}
```

**Impact**: Reduces validation code by ~30 lines per service

---

### Task 2: Create Logger Helper

**File**: `internal/logger/helpers.go`

**Contents**:
```go
package logger

import (
  "go.uber.org/zap"
  "goalkeeper-plan/internal/errors"
)

// LogServiceError logs an error with service context
func (l *zapLogger) LogServiceError(operation string, err error, fields ...zap.Field) {
  if err == nil {
    return
  }
  
  l.Error(
    operation + " failed",
    append(fields, zap.Error(err))...,
  )
}

// LogServiceSuccess logs successful operation
func (l *zapLogger) LogServiceSuccess(operation string, fields ...zap.Field) {
  l.Info(operation + " succeeded", fields...)
}

// LogAppError logs structured AppError
func (l *zapLogger) LogAppError(appErr *errors.AppError) {
  fields := []zap.Field{
    zap.String("error_type", string(appErr.Type)),
    zap.String("error_code", appErr.Code),
    zap.String("message", appErr.Message),
  }
  
  if appErr.Details != "" {
    fields = append(fields, zap.String("details", appErr.Details))
  }
  
  l.Error(appErr.Message, fields...)
}
```

**Impact**: Reduces logging code by ~50% in services

---

### Task 3: Refactor Workspace Service

**File**: `internal/workspace/service/workspace_service.go`

**Before** (153 lines):
```go
func (s *workspaceService) CreateWorkspace(ctx context.Context, name string, description *string) (*model.Workspace, error) {
  if name == "" {
    return nil, fmt.Errorf("workspace name cannot be empty")
  }

  ws := &model.Workspace{
    Name:        name,
    Description: description,
  }

  if err := s.repo.Create(ctx, ws); err != nil {
    if s.logger != nil {
      s.logger.Error("failed to create workspace", zap.Error(err), zap.String("name", name))
    }
    return nil, fmt.Errorf("failed to create workspace: %w", err)
  }

  if s.logger != nil {
    s.logger.Info("workspace created successfully", zap.String("id", ws.ID.String()), zap.String("name", name))
  }

  return ws, nil
}
```

**After** (~50 lines):
```go
func (s *workspaceService) CreateWorkspace(ctx context.Context, name string, description *string) (*model.Workspace, error) {
  // Validate input
  if err := validation.ValidateString(name, "workspace_name", 255); err != nil {
    return nil, err
  }

  // Create workspace
  ws := &model.Workspace{Name: name, Description: description}
  if err := s.repo.Create(ctx, ws); err != nil {
    s.logger.LogServiceError("create_workspace", err, zap.String("name", name))
    return nil, err
  }

  s.logger.LogServiceSuccess("create_workspace", zap.String("id", ws.ID.String()))
  return ws, nil
}
```

**Changes**:
- Use `validation.ValidateString()` instead of inline `if name == ""`
- Use `logger.LogServiceError()` and `logger.LogServiceSuccess()`
- Return AppError directly from validation
- Remove nil checks on logger

**Lines Saved**: ~40 lines for single method × 5 methods = ~200 lines per service

---

### Task 4: Refactor Block Service

Similar refactoring as workspace service (Task 3).

**Current**: ~199 lines
**Target**: ~100 lines
**Changes**: Same as Task 3

---

### Task 5: Refactor Page Service

Similar refactoring as workspace and block services.

**Current**: ~180 lines (estimated)
**Target**: ~90 lines

---

### Task 6: Create Controller Base Class

**File**: `internal/api/base_controller.go`

**Contents**:
```go
package api

import (
  "github.com/gin-gonic/gin"
  "goalkeeper-plan/internal/errors"
  "goalkeeper-plan/internal/logger"
)

type BaseController struct {
  Logger logger.Logger
}

// RespondError responds with structured error
func (bc *BaseController) RespondError(ctx *gin.Context, err error) {
  var appErr *errors.AppError
  
  // Convert to AppError if needed
  if appErr, ok := err.(*errors.AppError); !ok {
    appErr = &errors.AppError{
      Type:    errors.ErrorTypeInternal,
      Code:    "INTERNAL_ERROR",
      Message: "An unexpected error occurred",
      Err:     err,
    }
  }
  
  bc.Logger.LogAppError(appErr)
  ctx.JSON(appErr.GetHTTPStatusCode(), appErr)
}

// RespondSuccess responds with data and status code
func (bc *BaseController) RespondSuccess(ctx *gin.Context, statusCode int, data any) {
  ctx.JSON(statusCode, data)
}

// RespondCreated shortcut for 201 Created
func (bc *BaseController) RespondCreated(ctx *gin.Context, data any) {
  bc.RespondSuccess(ctx, 201, data)
}

// RespondOK shortcut for 200 OK
func (bc *BaseController) RespondOK(ctx *gin.Context, data any) {
  bc.RespondSuccess(ctx, 200, data)
}
```

**Impact**: Standardizes error handling across all controllers

---

### Task 7: Refactor Controllers

**Before**:
```go
func (c *WorkspaceController) Create(ctx *gin.Context) {
  var req CreateWorkspaceRequest
  if err := ctx.ShouldBindJSON(&req); err != nil {
    ctx.JSON(400, gin.H{"error": err.Error()})
    return
  }

  ws, err := c.service.CreateWorkspace(ctx.Context, req.Name, req.Description)
  if err != nil {
    ctx.JSON(500, gin.H{"error": err.Error()})
    return
  }

  ctx.JSON(201, ws)
}
```

**After**:
```go
func (c *WorkspaceController) Create(ctx *gin.Context) {
  var req CreateWorkspaceRequest
  if err := ctx.ShouldBindJSON(&req); err != nil {
    c.RespondError(ctx, err)
    return
  }

  ws, err := c.service.CreateWorkspace(ctx.Context, req.Name, req.Description)
  if err != nil {
    c.RespondError(ctx, err)
    return
  }

  c.RespondCreated(ctx, ws)
}
```

**Impact**: Cleaner, more consistent error handling

---

### Task 8: Extract Common Patterns to Utilities

**File**: `internal/service/base_service.go`

**Contents**:
```go
package service

// ServiceDependencies holds common service dependencies
type ServiceDependencies struct {
  Logger logger.Logger
}

// BaseService provides common functionality
type BaseService struct {
  Dependencies ServiceDependencies
}

// Logger getter with fallback
func (bs *BaseService) Log() logger.Logger {
  if bs.Dependencies.Logger == nil {
    // Return no-op logger if nil
    return logger.NewNopLogger()
  }
  return bs.Dependencies.Logger
}
```

**Impact**: Eliminates nil checks throughout services

---

## Refactoring Checklist

### Phase 1: Utilities & Helpers
- [ ] Create `internal/validation/validation.go`
- [ ] Create `internal/logger/helpers.go`
- [ ] Create `internal/api/base_controller.go`
- [ ] Create `internal/service/base_service.go` (if needed)

### Phase 2: Service Refactoring (Sequential)
- [ ] Refactor workspace service (~30 min)
  - [ ] Update imports
  - [ ] Replace validation patterns
  - [ ] Replace logging patterns
  - [ ] Update tests
  - [ ] Run tests: `go test ./internal/workspace/service`

- [ ] Refactor block service (~30 min)
  - [ ] Same steps as workspace
  - [ ] Run tests: `go test ./internal/block/service`

- [ ] Refactor page service (~30 min)
  - [ ] Same steps as workspace
  - [ ] Run tests: `go test ./internal/page/service`

### Phase 3: Controller Refactoring (Parallel)
- [ ] Update workspace controller (~15 min)
  - [ ] Extend BaseController
  - [ ] Update error handling
  - [ ] Update response methods
  - [ ] Run tests: `go test ./internal/workspace/controller`

- [ ] Update block controller (~15 min)
  - [ ] Same steps as workspace
  - [ ] Run tests: `go test ./internal/block/controller`

- [ ] Update page controller (~15 min)
  - [ ] Same steps as workspace
  - [ ] Run tests: `go test ./internal/page/controller`

- [ ] Update RBAC controller (~10 min)
  - [ ] Same steps as workspace

### Phase 4: RBAC Service Refactoring
- [ ] Refactor RBAC service (~20 min)
  - [ ] Apply same patterns as other services

### Phase 5: Testing & Validation
- [ ] Run full test suite: `go test ./...`
- [ ] Type check: `go vet ./...`
- [ ] Format: `go fmt ./...`
- [ ] Build: `go build ./cmd/main.go`
- [ ] Verify API still works with manual tests

---

## Code Metrics Targets

| Metric | Current | Target | % Reduction |
|--------|---------|--------|-------------|
| Service files lines | ~550 | ~350 | 36% |
| Logging statements | ~80 | ~40 | 50% |
| Error handling code | ~60 | ~20 | 67% |
| Validation code | ~50 | ~10 | 80% |
| Total service code | ~660 | ~420 | 36% |

---

## Benefits After Refactoring

1. **Maintainability**: Single source of truth for patterns
2. **Consistency**: All services follow same approach
3. **Reduced Bugs**: Centralized validation and error handling
4. **Testability**: Easier to mock and test
5. **Readability**: Less boilerplate, clearer intent
6. **Scalability**: Adding new services is now faster
7. **Monitoring**: Structured error codes for tracking

---

## Rollback Plan

Each refactored file has:
1. Original backed up in git
2. Tests verify functionality unchanged
3. Git history available for comparison
4. If issues found: `git revert <commit>`

---

## Success Criteria

- [ ] All tests pass: `go test ./...`
- [ ] No compilation errors: `go build ./cmd/main.go`
- [ ] Manual API testing successful
- [ ] Code coverage maintained or improved
- [ ] Performance metrics unchanged
- [ ] PR reviewed and approved

---

## Time Estimate

- Phase 1 (Utilities): 1 hour
- Phase 2 (Services): 1.5 hours  
- Phase 3 (Controllers): 1 hour
- Phase 4 (RBAC): 30 minutes
- Phase 5 (Testing): 1 hour
- **Total**: ~5 hours

---

## Files to Modify

### New Files
- `internal/validation/validation.go`
- `internal/logger/helpers.go`
- `internal/api/base_controller.go`

### Modified Files
- `internal/workspace/service/workspace_service.go`
- `internal/workspace/controller/controller.go`
- `internal/block/service/block_service.go`
- `internal/block/controller/controller.go`
- `internal/page/service/service.go`
- `internal/page/controller/controller.go`
- `internal/rbac/service/service.go`
- `internal/rbac/controller/controller.go`

### Potentially Modified
- `internal/workspace/service/sharing_service.go`
- Constructor functions in each module's `app/` folder

---

## Implementation Strategy

1. **Create utilities first** - No risk, foundational
2. **Refactor services** - Lower risk, isolated logic
3. **Refactor controllers** - Lower risk, HTTP layer only
4. **Test thoroughly** - Verify behavior unchanged
5. **Deploy carefully** - One module at a time

