package app

import (
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/logger"
	rbacController "goalkeeper-plan/internal/rbac/controller"
	rbacRepo "goalkeeper-plan/internal/rbac/repository"
	rbacRouter "goalkeeper-plan/internal/rbac/router"
	rbacService "goalkeeper-plan/internal/rbac/service"

	"gorm.io/gorm"
)

// NewRBACApplication initializes and registers RBAC routes
func NewRBACApplication(db *gorm.DB, baseRouter interface{}, configs config.Configurations, logger logger.Logger) {
	// Initialize repositories
	roleRepo := rbacRepo.NewRoleRepository(db)
	permissionRepo := rbacRepo.NewPermissionRepository(db)

	// Initialize services
	roleService := rbacService.NewRoleService(roleRepo, permissionRepo, logger)
	permissionService := rbacService.NewPermissionService(permissionRepo, logger)

	// Initialize controllers
	roleController := rbacController.NewRoleController(roleService, permissionService, logger)
	permissionController := rbacController.NewPermissionController(permissionService, logger)

	// Register routes
	rbacRouter.NewRoleRouter(baseRouter, roleController)
	rbacRouter.NewPermissionRouter(baseRouter, permissionController)
}
