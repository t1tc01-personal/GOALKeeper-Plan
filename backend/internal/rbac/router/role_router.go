package router

import (
	"goalkeeper-plan/internal/rbac/controller"

	"github.com/gin-gonic/gin"
)

func NewRoleRouter(baseRouter interface{}, c controller.RoleController) {
	// Type assertion to get the Gin router group
	group, ok := baseRouter.(*gin.RouterGroup)
	if !ok {
		panic("baseRouter must be *gin.RouterGroup")
	}

	// Role routes
	roles := group.Group("/roles")
	{
		roles.POST("", c.CreateRole)
		roles.GET("", c.ListRoles)
		roles.GET("/name", c.GetRoleByName)
		roles.GET("/:id", c.GetRoleByID)
		roles.GET("/:id/permissions", c.GetRoleByIDWithPermissions)
		roles.PUT("/:id", c.UpdateRole)
		roles.PATCH("/:id", c.UpdateRole)
		roles.DELETE("/:id", c.DeleteRole)
		roles.POST("/:id/permissions", c.AssignPermissionToRole)
		roles.DELETE("/:id/permissions", c.RemovePermissionFromRole)
	}
}
