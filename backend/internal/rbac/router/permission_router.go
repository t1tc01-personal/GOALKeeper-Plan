package router

import (
	"goalkeeper-plan/internal/rbac/controller"

	"github.com/gin-gonic/gin"
)

func NewPermissionRouter(baseRouter interface{}, c controller.PermissionController) {
	// Type assertion to get the Gin router group
	group, ok := baseRouter.(*gin.RouterGroup)
	if !ok {
		panic("baseRouter must be *gin.RouterGroup")
	}

	// Permission routes
	permissions := group.Group("/permissions")
	{
		permissions.POST("", c.CreatePermission)
		permissions.GET("", c.ListPermissions)
		permissions.GET("/name", c.GetPermissionByName)
		permissions.GET("/resource", c.ListPermissionsByResource)
		permissions.GET("/resource-action", c.GetPermissionByResourceAndAction)
		permissions.GET("/:id", c.GetPermissionByID)
		permissions.PUT("/:id", c.UpdatePermission)
		permissions.PATCH("/:id", c.UpdatePermission)
		permissions.DELETE("/:id", c.DeletePermission)
	}
}
