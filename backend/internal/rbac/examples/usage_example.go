package examples

// This file contains examples of how to use RBAC in your routes.
// These are examples only and should not be compiled.

/*
import (
	"goalkeeper-plan/internal/auth/middleware"
	rbacApp "goalkeeper-plan/internal/rbac/app"
	rbacMiddleware "goalkeeper-plan/internal/rbac/middleware"
	"goalkeeper-plan/internal/logger"
	
	"github.com/gin-gonic/gin"
)

// Example: Setting up RBAC in your router
func ExampleSetupRBAC(router *gin.Engine, db *gorm.DB, logger logger.Logger) {
	// Initialize RBAC app
	rbacAppInstance, err := rbacApp.NewRBACApp(db)
	if err != nil {
		log.Fatal("Failed to initialize RBAC:", err)
	}

	// Initialize auth middleware (required before RBAC)
	authMiddleware := middleware.AuthMiddleware(jwtService, logger)

	// Public routes
	router.POST("/auth/signup", authController.SignUp)
	router.POST("/auth/login", authController.Login)

	// Protected API routes
	api := router.Group("/api/v1")
	api.Use(authMiddleware) // All routes require authentication

	// Example 1: Require specific permission
	api.GET("/users",
		rbacMiddleware.RequirePermission(rbacAppInstance.RBACService, "users.read", logger),
		userController.ListUsers,
	)

	// Example 2: Require any of multiple permissions
	api.POST("/users",
		rbacMiddleware.RequireAnyPermission(
			rbacAppInstance.RBACService,
			"users.create",
			"users.manage",
		),
		userController.CreateUser,
	)

	// Example 3: Require all permissions
	api.DELETE("/users/:id",
		rbacMiddleware.RequireAllPermissions(
			rbacAppInstance.RBACService,
			"users.delete",
			"users.manage",
		),
		userController.DeleteUser,
	)

	// Example 4: Require specific role
	admin := api.Group("/admin")
	admin.Use(rbacMiddleware.RequireRole(rbacAppInstance.RBACService, "admin", logger))
	{
		admin.GET("/dashboard", adminController.Dashboard)
		admin.POST("/roles", adminController.CreateRole)
	}

	// Example 5: Require permission by resource and action
	api.PUT("/goals/:id",
		rbacMiddleware.RequirePermissionByResourceAction(
			rbacAppInstance.RBACService,
			"goals",
			"update",
			logger,
		),
		goalController.UpdateGoal,
	)

	// Example 6: Using RBAC service directly in controller
	// In your controller:
	func (c *UserController) UpdateUser(ctx *gin.Context) {
		userID := ctx.GetString("user_id")
		
		// Check permission programmatically
		hasPermission, err := rbacAppInstance.RBACService.HasPermission(userID, "users.update")
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		
		if !hasPermission {
			ctx.JSON(403, gin.H{"error": "Forbidden"})
			return
		}
		
		// Continue with update logic...
	}
}
*/
