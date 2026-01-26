package app

import (
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/auth/controller"
	"goalkeeper-plan/internal/auth/jwt"
	"goalkeeper-plan/internal/auth/oauth"
	"goalkeeper-plan/internal/auth/router"
	"goalkeeper-plan/internal/auth/service"
	tokenStorage "goalkeeper-plan/internal/auth/token_storage"
	"goalkeeper-plan/internal/logger"
	redisClient "goalkeeper-plan/internal/redis"
	"goalkeeper-plan/internal/user/repository"

	"gorm.io/gorm"
)

func NewApplication(db *gorm.DB, baseRouter interface{}, configs config.Configurations, logger logger.Logger) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize Redis client
	redisCli, err := redisClient.GetRedisClient(configs.RedisConfig, logger)
	if err != nil {
		panic(err)
	}

	// Initialize token storage
	tokenStorageService := tokenStorage.NewTokenStorage(redisCli)

	// Initialize JWT service
	jwtService := jwt.NewJWTService(
		configs.AuthConfig.Secret,
		configs.AuthConfig.JWTExpiration,
		configs.AuthConfig.RefreshExpiration,
	)

	// Initialize OAuth service
	oauthService := oauth.NewOAuthService(configs)

	// Initialize auth service
	authService, err := service.NewAuthService(
		service.WithUserRepository(userRepo),
		service.WithJWTService(jwtService),
		service.WithOAuthService(oauthService),
		service.WithConfig(configs),
	)
	if err != nil {
		panic(err)
	}

	// Initialize auth controller
	authController := controller.NewAuthController(authService, tokenStorageService, configs, logger)

	// Register routes
	router.NewRouter(baseRouter, authController)
}

