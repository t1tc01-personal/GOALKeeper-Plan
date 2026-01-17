package controller

import (
	"strconv"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/user/dto"
	"goalkeeper-plan/internal/user/service"

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

func (c *userController) CreateUser(ctx *gin.Context) {
	var req dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Failed to decode request", zap.Error(err))
		ctx.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	user, err := c.userService.CreateUser(req.Email, req.Name)
	if err != nil {
		c.logger.Error("Failed to create user", zap.Error(err))
		ctx.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}

	ctx.JSON(201, dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (c *userController) GetUserByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(400, gin.H{"error": "User ID is required"})
		return
	}

	user, err := c.userService.GetUserByID(id)
	if err != nil {
		c.logger.Error("Failed to get user", zap.Error(err))
		ctx.JSON(404, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(200, dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (c *userController) GetUserByEmail(ctx *gin.Context) {
	email := ctx.Query("email")
	if email == "" {
		ctx.JSON(400, gin.H{"error": "Email is required"})
		return
	}

	user, err := c.userService.GetUserByEmail(email)
	if err != nil {
		c.logger.Error("Failed to get user", zap.Error(err))
		ctx.JSON(404, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(200, dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (c *userController) UpdateUser(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(400, gin.H{"error": "User ID is required"})
		return
	}

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Failed to decode request", zap.Error(err))
		ctx.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	user, err := c.userService.GetUserByID(id)
	if err != nil {
		c.logger.Error("Failed to get user", zap.Error(err))
		ctx.JSON(404, gin.H{"error": "User not found"})
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := c.userService.UpdateUser(user); err != nil {
		c.logger.Error("Failed to update user", zap.Error(err))
		ctx.JSON(500, gin.H{"error": "Failed to update user"})
		return
	}

	ctx.JSON(200, dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (c *userController) DeleteUser(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(400, gin.H{"error": "User ID is required"})
		return
	}

	if err := c.userService.DeleteUser(id); err != nil {
		c.logger.Error("Failed to delete user", zap.Error(err))
		ctx.JSON(500, gin.H{"error": "Failed to delete user"})
		return
	}

	ctx.JSON(200, gin.H{"message": "User deleted successfully"})
}

func (c *userController) ListUsers(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "10")
	offsetStr := ctx.DefaultQuery("offset", "0")

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
		ctx.JSON(500, gin.H{"error": "Failed to list users"})
		return
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

	ctx.JSON(200, responses)
}
