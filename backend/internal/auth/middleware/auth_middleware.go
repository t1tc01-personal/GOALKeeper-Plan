package middleware

import (
	"goalkeeper-plan/internal/auth/jwt"
	"goalkeeper-plan/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthMiddleware(jwtService *jwt.JWTService, logger logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			logger.Error("Missing authorization header")
			ctx.JSON(401, gin.H{"error": "Authorization header is required"})
			ctx.Abort()
			return
		}

		tokenString := jwt.ExtractTokenFromHeader(authHeader)
		if tokenString == "" {
			logger.Error("Invalid authorization header format")
			ctx.JSON(401, gin.H{"error": "Invalid authorization header format"})
			ctx.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				logger.Error("Token expired")
				ctx.JSON(401, gin.H{"error": "Token has expired"})
			} else {
				logger.Error("Invalid token", zap.Error(err))
				ctx.JSON(401, gin.H{"error": "Invalid token"})
			}
			ctx.Abort()
			return
		}

		// Set user info in context
		ctx.Set("user_id", claims.UserID)
		ctx.Set("email", claims.Email)

		ctx.Next()
	}
}

