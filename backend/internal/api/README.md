# API Handler Helpers

Package này cung cấp các helper functions để giảm boilerplate code trong controllers, giữ code clean và consistent.

## Features

- ✅ Generic request binding và validation
- ✅ Automatic error handling
- ✅ Standardized response format
- ✅ Type-safe với Go generics
- ✅ Flexible - có thể customize khi cần

## Usage Examples

### 1. Handle POST Request với JSON Body

```go
func (c *userController) CreateUser(ctx *gin.Context) {
    api.HandleRequestWithStatus(ctx, api.HandlerConfig{
        Logger: c.logger,
    }, 201, userMessages.MsgUserCreated, func(ctx *gin.Context, req dto.CreateUserRequest) (interface{}, error) {
        user, err := c.userService.CreateUser(req.Email, req.Name)
        if err != nil {
            return nil, errors.NewInternalError(
                errors.CodeFailedToCreateUser,
                userMessages.MsgFailedToCreateUser,
                err,
            )
        }
        return dto.UserResponse{...}, nil
    })
}
```

### 2. Handle GET Request với URL Parameter

```go
func (c *userController) GetUserByID(ctx *gin.Context) {
    api.HandleParamRequest(ctx, "id", api.HandlerConfig{
        Logger: c.logger,
    }, func(ctx *gin.Context, id string) (interface{}, error) {
        user, err := c.userService.GetUserByID(id)
        if err != nil {
            return nil, errors.NewNotFoundError(
                errors.CodeUserNotFound,
                userMessages.MsgUserNotFound,
                err,
            )
        }
        return dto.UserResponse{...}, nil
    })
}
```

### 3. Handle GET Request với Query Parameters

```go
func (c *userController) ListUsers(ctx *gin.Context) {
    api.HandleQueryRequestWithMessage(ctx, api.HandlerConfig{
        Logger: c.logger,
    }, userMessages.MsgUsersListed, func(ctx *gin.Context) (interface{}, error) {
        limitStr := api.GetQuery(ctx, "limit", "10")
        offsetStr := api.GetQuery(ctx, "offset", "0")
        // ... process query params
        users, err := c.userService.ListUsers(limit, offset)
        if err != nil {
            return nil, errors.NewInternalError(...)
        }
        return responses, nil
    })
}
```

### 4. Manual Approach (cho complex cases)

```go
func (c *userController) UpdateUser(ctx *gin.Context) {
    // Get param manually
    id, ok := api.GetParam(ctx, "id", userMessages.MsgInvalidUserID)
    if !ok {
        return
    }

    // Bind request manually
    req, ok := api.BindRequest[dto.UpdateUserRequest](ctx, c.logger)
    if !ok {
        return
    }

    // Business logic
    user, err := c.userService.GetUserByID(id)
    if err != nil {
        api.HandleError(ctx, errors.NewNotFoundError(...), c.logger)
        return
    }

    // Update logic...
    
    // Send success
    api.SendSuccess(ctx, 200, userMessages.MsgUserUpdated, dto.UserResponse{...})
}
```

## Available Functions

### High-level Handlers

- `HandleRequest[T]` - Handle POST/PUT với JSON body
- `HandleRequestWithStatus[T]` - Handle request với custom status code
- `HandleParamRequest` - Handle GET với URL parameter
- `HandleParamRequestWithMessage` - Handle param request với custom message
- `HandleQueryRequest` - Handle GET với query parameters
- `HandleQueryRequestWithMessage` - Handle query request với custom message

### Low-level Helpers

- `BindRequest[T]` - Bind và validate request, return error nếu failed
- `GetParam` - Get URL parameter, return error nếu missing
- `GetQuery` - Get query parameter với default value
- `HandleError` - Process error và send response
- `SendSuccess` - Send success response
- `SendError` - Send error response

## Benefits

1. **DRY**: Giảm code duplication
2. **Consistent**: Tất cả API có cùng error handling pattern
3. **Type-safe**: Sử dụng Go generics
4. **Flexible**: Có thể mix manual và helper functions
5. **Maintainable**: Dễ maintain và update

## Migration Guide

### Before (Old Pattern)
```go
func (c *controller) Create(ctx *gin.Context) {
    var req dto.Request
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.logger.Error("Failed to bind", zap.Error(err))
        appErr := errors.NewValidationError(...)
        response.ErrorResponse(ctx, appErr)
        return
    }
    
    result, err := c.service.Create(req)
    if err != nil {
        c.logger.Error("Failed", zap.Error(err))
        appErr := errors.NewInternalError(...)
        response.ErrorResponse(ctx, appErr)
        return
    }
    
    response.SuccessResponse(ctx, 201, "Created", result)
}
```

### After (New Pattern)
```go
func (c *controller) Create(ctx *gin.Context) {
    api.HandleRequestWithStatus(ctx, api.HandlerConfig{
        Logger: c.logger,
    }, 201, "Created", func(ctx *gin.Context, req dto.Request) (interface{}, error) {
        result, err := c.service.Create(req)
        if err != nil {
            return nil, errors.NewInternalError(...)
        }
        return result, nil
    })
}
```

## Notes

- Helper functions tự động handle request binding và validation
- Error handling tự động convert AppError thành HTTP response
- Có thể return AppError từ handler function, sẽ được tự động handle
- Logger được pass vào config để log errors

