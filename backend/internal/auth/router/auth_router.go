package router

import (
	"goalkeeper-plan/internal/auth/controller"

	"github.com/gin-gonic/gin"
)

func NewRouter(baseRouter interface{}, c controller.AuthController) {
	// Type assertion to get the Gin router group
	group, ok := baseRouter.(*gin.RouterGroup)
	if !ok {
		panic("baseRouter must be *gin.RouterGroup")
	}

	// Auth routes
	auth := group.Group("/auth")
	{
		// Public routes
		auth.POST("/signup", c.SignUp)
		auth.POST("/login", c.Login)
		auth.POST("/logout", c.Logout)
		auth.POST("/refresh", c.RefreshToken)

		// OAuth routes
		oauth := auth.Group("/oauth")
		{
			oauth.GET("/:provider/url", c.GetOAuthURL)
			oauth.GET("/:provider/callback", c.OAuthCallbackGET)          // GET for OAuth provider redirect
			oauth.POST("/:provider/callback", c.OAuthCallback)            // POST for frontend API call with code
			oauth.POST("/:provider/token-callback", c.OAuthTokenCallback) // POST for frontend API call with token
			oauth.GET("/tokens/:token_id", c.GetOAuthTokens)              // Get tokens by token ID
		}
	}
}
