package app

import (
	rbacRepo "goalkeeper-plan/internal/rbac/repository"
	"goalkeeper-plan/internal/rbac/service"
	userRepo "goalkeeper-plan/internal/user/repository"

	"gorm.io/gorm"
)

// RBACApp holds all RBAC-related dependencies
type RBACApp struct {
	RoleRepository       rbacRepo.RoleRepository
	PermissionRepository rbacRepo.PermissionRepository
	RBACService          service.RBACService
}

// NewRBACApp initializes and returns a new RBACApp with all dependencies
func NewRBACApp(db *gorm.DB) (*RBACApp, error) {
	// Initialize repositories
	roleRepo := rbacRepo.NewRoleRepository(db)
	permissionRepo := rbacRepo.NewPermissionRepository(db)
	userRepository := userRepo.NewUserRepository(db)

	// Initialize RBAC service
	rbacService, err := service.NewRBACService(
		service.WithRoleRepository(roleRepo),
		service.WithPermissionRepository(permissionRepo),
		service.WithUserRepository(userRepository),
	)
	if err != nil {
		return nil, err
	}

	return &RBACApp{
		RoleRepository:       roleRepo,
		PermissionRepository: permissionRepo,
		RBACService:          rbacService,
	}, nil
}
