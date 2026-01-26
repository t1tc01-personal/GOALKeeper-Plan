package app

import (
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/logger"
	rbacController "goalkeeper-plan/internal/rbac/controller"
	rbacRouter "goalkeeper-plan/internal/rbac/router"
	rbacService "goalkeeper-plan/internal/rbac/service"
	rbacRepo "goalkeeper-plan/internal/rbac/repository"

	"gorm.io/gorm"
)

// NewRBACApplication initializes and registers RBAC routes
func NewRBACApplication(db *gorm.DB, baseRouter interface{}, configs config.Configurations, logger logger.Logger) {
	// Initialize repositories
	roleRepo := rbacRepo.NewRoleRepository(db)
	permissionRepo := rbacRepo.NewPermissionRepository(db)

	// Initialize services
	roleService := rbacService.NewRoleService(roleRepo, permissionRepo)
	permissionService := rbacService.NewPermissionService(permissionRepo)

	// Initialize controllers
	roleController := rbacController.NewRoleController(roleService, permissionService, logger)
	permissionController := rbacController.NewPermissionController(permissionService, logger)

	// Register routes
	rbacRouter.NewRoleRouter(baseRouter, roleController)
	rbacRouter.NewPermissionRouter(baseRouter, permissionController)
}
