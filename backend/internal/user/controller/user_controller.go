package controller

import (
	"goalkeeper-plan/internal/api"
	"goalkeeper-plan/internal/errors"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/user/dto"
	userMessages "goalkeeper-plan/internal/user/messages"
	"goalkeeper-plan/internal/user/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserController interface {
	CreateUser(*gin.Context)
	GetUserByID(*gin.Context)
	GetUserByEmail(*gin.Context)
	UpdateUser(*gin.Context)
	DeleteUser(*gin.Context)
	ListUsers(*gin.Context)
}

type userController struct {
	userService service.UserService
	logger      logger.Logger
}

func NewUserController(s service.UserService, l logger.Logger) UserController {
	return &userController{
		userService: s,
		logger:      l,
	}
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Create a new user with email and name
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request   body      dto.CreateUserRequest  true  "Create user request"
// @Success      201       {object}  response.Response{data=dto.UserResponse}
// @Failure      400       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /users [post]
func (c *userController) CreateUser(ctx *gin.Context) {
	api.HandleRequestWithStatus(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, 201, userMessages.MsgUserCreated, func(ctx *gin.Context, req dto.CreateUserRequest) (interface{}, error) {
		user, err := c.userService.CreateUser(req.Email, req.Name)
		if err != nil {
			c.logger.Error("Failed to create user", zap.Error(err), zap.String("email", req.Email))
			return nil, errors.NewInternalError(
				errors.CodeFailedToCreateUser,
				userMessages.MsgFailedToCreateUser,
				err,
			)
		}

		return dto.UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}, nil
	})
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Get user information by user ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "User ID"
// @Success      200       {object}  response.Response{data=dto.UserResponse}
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Router       /users/{id} [get]
func (c *userController) GetUserByID(ctx *gin.Context) {
	api.HandleParamRequest(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context, id string) (interface{}, error) {
		user, err := c.userService.GetUserByID(id)
		if err != nil {
			c.logger.Error("Failed to get user", zap.Error(err), zap.String("user_id", id))
			return nil, errors.NewNotFoundError(
				errors.CodeUserNotFound,
				userMessages.MsgUserNotFound,
				err,
			)
		}

		return dto.UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}, nil
	})
}

// GetUserByEmail godoc
// @Summary      Get user by email
// @Description  Get user information by email address
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        email     query     string  true  "User email"
// @Success      200       {object}  response.Response{data=dto.UserResponse}
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Router       /users/email [get]
func (c *userController) GetUserByEmail(ctx *gin.Context) {
	api.HandleQueryRequest(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, func(ctx *gin.Context) (interface{}, error) {
		email := ctx.Query("email")
		if email == "" {
			return nil, errors.NewValidationError(
				errors.CodeMissingField,
				userMessages.MsgInvalidUserID,
				nil,
			)
		}

		user, err := c.userService.GetUserByEmail(email)
		if err != nil {
			c.logger.Error("Failed to get user", zap.Error(err), zap.String("email", email))
			return nil, errors.NewNotFoundError(
				errors.CodeUserNotFound,
				userMessages.MsgUserNotFound,
				err,
			)
		}

		return dto.UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}, nil
	})
}

// UpdateUser godoc
// @Summary      Update user
// @Description  Update user information
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id        path      string              true  "User ID"
// @Param        request   body      dto.UpdateUserRequest  true  "Update user request"
// @Success      200       {object}  response.Response{data=dto.UserResponse}
// @Failure      400       {object}  response.Response
// @Failure      404       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /users/{id} [put]
func (c *userController) UpdateUser(ctx *gin.Context) {
	id, ok := api.GetParam(ctx, "id", userMessages.MsgInvalidUserID)
	if !ok {
		return
	}

	req, ok := api.BindRequest[dto.UpdateUserRequest](ctx, c.logger)
	if !ok {
		return
	}

	user, err := c.userService.GetUserByID(id)
	if err != nil {
		c.logger.Error("Failed to get user", zap.Error(err), zap.String("user_id", id))
		api.HandleError(ctx, errors.NewNotFoundError(
			errors.CodeUserNotFound,
			userMessages.MsgUserNotFound,
			err,
		), c.logger)
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := c.userService.UpdateUser(user); err != nil {
		c.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", id))
		api.HandleError(ctx, errors.NewInternalError(
			errors.CodeFailedToUpdateUser,
			userMessages.MsgFailedToUpdateUser,
			err,
		), c.logger)
		return
	}

	api.SendSuccess(ctx, 200, userMessages.MsgUserUpdated, dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Delete a user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "User ID"
// @Success      200       {object}  response.Response
// @Failure      400       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /users/{id} [delete]
func (c *userController) DeleteUser(ctx *gin.Context) {
	api.HandleParamRequestWithMessage(ctx, "id", api.HandlerConfig{
		Logger: c.logger,
	}, userMessages.MsgUserDeleted, func(ctx *gin.Context, id string) (interface{}, error) {
		if err := c.userService.DeleteUser(id); err != nil {
			c.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", id))
			return nil, errors.NewInternalError(
				errors.CodeFailedToDeleteUser,
				userMessages.MsgFailedToDeleteUser,
				err,
			)
		}
		return nil, nil
	})
}

// ListUsers godoc
// @Summary      List users
// @Description  Get a list of users with pagination
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        limit     query     int     false  "Limit (default: 10)"
// @Param        offset    query     int     false  "Offset (default: 0)"
// @Success      200       {object}  response.Response{data=[]dto.UserResponse}
// @Failure      500       {object}  response.Response
// @Router       /users [get]
func (c *userController) ListUsers(ctx *gin.Context) {
	api.HandleQueryRequestWithMessage(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, userMessages.MsgUsersListed, func(ctx *gin.Context) (interface{}, error) {
		limitStr := api.GetQuery(ctx, "limit", "10")
		offsetStr := api.GetQuery(ctx, "offset", "0")

		limit := 10
		offset := 0

		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}

		users, err := c.userService.ListUsers(limit, offset)
		if err != nil {
			c.logger.Error("Failed to list users", zap.Error(err))
			return nil, errors.NewInternalError(
				errors.CodeFailedToListUsers,
				userMessages.MsgFailedToListUsers,
				err,
			)
		}

		responses := make([]dto.UserResponse, len(users))
		for i, user := range users {
			responses[i] = dto.UserResponse{
				ID:        user.ID.String(),
				Email:     user.Email,
				Name:      user.Name,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			}
		}

		return responses, nil
	})
}
