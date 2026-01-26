package middleware

import (
	"goalkeeper-plan/internal/rbac/service"
	"goalkeeper-plan/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequirePermission creates a middleware that requires a specific permission
func RequirePermission(rbacService service.RBACService, permissionName string, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			logger.Error("User ID not found in context")
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			logger.Error("Invalid user ID type in context")
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		hasPermission, err := rbacService.HasPermission(userIDStr, permissionName)
		if err != nil {
			logger.Error("Error checking permission", zap.Error(err))
			c.JSON(500, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		if !hasPermission {
			logger.Warn("Permission denied",
				zap.String("user_id", userIDStr),
				zap.String("permission", permissionName),
			)
			c.JSON(403, gin.H{
				"error":      "Forbidden",
				"message":    "You do not have permission to perform this action",
				"permission": permissionName,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission creates a middleware that requires at least one of the specified permissions
func RequireAnyPermission(rbacService service.RBACService, permissionNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		hasPermission, err := rbacService.HasAnyPermission(userIDStr, permissionNames...)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(403, gin.H{
				"error":      "Forbidden",
				"message":    "You do not have permission to perform this action",
				"permissions": permissionNames,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissions creates a middleware that requires all of the specified permissions
func RequireAllPermissions(rbacService service.RBACService, permissionNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		hasPermission, err := rbacService.HasAllPermissions(userIDStr, permissionNames...)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(403, gin.H{
				"error":      "Forbidden",
				"message":    "You do not have all required permissions to perform this action",
				"permissions": permissionNames,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(rbacService service.RBACService, roleName string, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			logger.Error("User ID not found in context")
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			logger.Error("Invalid user ID type in context")
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		hasRole, err := rbacService.HasRole(userIDStr, roleName)
		if err != nil {
			logger.Error("Error checking role", zap.Error(err))
			c.JSON(500, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		if !hasRole {
			logger.Warn("Role denied",
				zap.String("user_id", userIDStr),
				zap.String("role", roleName),
			)
			c.JSON(403, gin.H{
				"error":   "Forbidden",
				"message": "You do not have the required role to perform this action",
				"role":    roleName,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole creates a middleware that requires at least one of the specified roles
func RequireAnyRole(rbacService service.RBACService, roleNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		hasRole, err := rbacService.HasAnyRole(userIDStr, roleNames...)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		if !hasRole {
			c.JSON(403, gin.H{
				"error":   "Forbidden",
				"message": "You do not have any of the required roles to perform this action",
				"roles":   roleNames,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermissionByResourceAction creates a middleware that requires permission for a resource and action
func RequirePermissionByResourceAction(rbacService service.RBACService, resource, action string, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			logger.Error("User ID not found in context")
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			logger.Error("Invalid user ID type in context")
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		hasPermission, err := rbacService.HasPermissionByResourceAction(userIDStr, resource, action)
		if err != nil {
			logger.Error("Error checking permission", zap.Error(err))
			c.JSON(500, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		if !hasPermission {
			logger.Warn("Permission denied",
				zap.String("user_id", userIDStr),
				zap.String("resource", resource),
				zap.String("action", action),
			)
			c.JSON(403, gin.H{
				"error":    "Forbidden",
				"message":  "You do not have permission to perform this action",
				"resource": resource,
				"action":   action,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
