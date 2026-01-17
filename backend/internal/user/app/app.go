package app

import (
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/logger"
	"goalkeeper-plan/internal/user/controller"
	"goalkeeper-plan/internal/user/repository"
	"goalkeeper-plan/internal/user/router"
	"goalkeeper-plan/internal/user/service"

	"gorm.io/gorm"
)

func NewApplication(db *gorm.DB, baseRouter interface{}, configs config.Configurations, logger logger.Logger) {
	userRepo := repository.NewUserRepository(db)

	userService, err := service.NewUserService(
		service.WithUserRepository(userRepo),
	)
	if err != nil {
		panic(err)
	}

	userController := controller.NewUserController(userService, logger)
	router.NewRouter(baseRouter, userController)
}

