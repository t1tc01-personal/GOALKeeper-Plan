package router

import (
	"goalkeeper-plan/internal/user/controller"

	"github.com/gin-gonic/gin"
)

func NewRouter(baseRouter interface{}, c controller.UserController) {
	// Type assertion to get the Gin router group
	group, ok := baseRouter.(*gin.RouterGroup)
	if !ok {
		panic("baseRouter must be *gin.RouterGroup")
	}

	// User routes
	users := group.Group("/users")
	{
		users.POST("", c.CreateUser)
		users.GET("", c.ListUsers)
		users.GET("/:id", c.GetUserByID)
		users.PUT("/:id", c.UpdateUser)
		users.PATCH("/:id", c.UpdateUser)
		users.DELETE("/:id", c.DeleteUser)
	}
}

