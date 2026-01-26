# RBAC (Role-Based Access Control) System

This package implements a comprehensive Role-Based Access Control system for the GOALKeeper-Plan backend.

## Overview

The RBAC system provides:
- **Roles**: Group permissions together (e.g., `admin`, `user`, `moderator`)
- **Permissions**: Granular access control (e.g., `users.create`, `users.read`)
- **User-Role Assignment**: Assign roles to users
- **Role-Permission Assignment**: Assign permissions to roles

## Architecture

```
Models:
- Role: Represents a role in the system
- Permission: Represents a permission (resource + action)
- UserRole: Many-to-many relationship between users and roles
- RolePermission: Many-to-many relationship between roles and permissions

Repositories:
- RoleRepository: CRUD operations for roles
- PermissionRepository: CRUD operations for permissions

Service:
- RBACService: Business logic for permission checking and role management

Middleware:
- Authorization middleware for protecting routes
```

## Database Migration

Run the migration to create RBAC tables:

```sql
-- Located in: migrations/20260126040000_rbac_tables.sql
```

This migration creates:
- `roles` table
- `permissions` table
- `role_permissions` junction table
- `user_roles` junction table
- Default roles: `admin`, `user`, `moderator`
- Default permissions for users, roles, permissions, goals, and tasks

## Usage

### 1. Initialize RBAC App

```go
import rbacApp "goalkeeper-plan/internal/rbac/app"

// In your main application setup
rbacAppInstance, err := rbacApp.NewRBACApp(db)
if err != nil {
    log.Fatal(err)
}
```

### 2. Using Authorization Middleware

#### Require a specific permission:
```go
import rbacMiddleware "goalkeeper-plan/internal/rbac/middleware"

// Protect a route with a permission
router.GET("/users",
    authMiddleware, // Must be authenticated first
    rbacMiddleware.RequirePermission(rbacAppInstance.RBACService, "users.read", logger),
    userController.ListUsers,
)
```

#### Require any of multiple permissions:
```go
router.POST("/users",
    authMiddleware,
    rbacMiddleware.RequireAnyPermission(
        rbacAppInstance.RBACService,
        "users.create",
        "users.manage",
    ),
    userController.CreateUser,
)
```

#### Require all permissions:
```go
router.DELETE("/users/:id",
    authMiddleware,
    rbacMiddleware.RequireAllPermissions(
        rbacAppInstance.RBACService,
        "users.delete",
        "users.manage",
    ),
    userController.DeleteUser,
)
```

#### Require a role:
```go
router.GET("/admin/dashboard",
    authMiddleware,
    rbacMiddleware.RequireRole(rbacAppInstance.RBACService, "admin", logger),
    adminController.Dashboard,
)
```

#### Require permission by resource and action:
```go
router.PUT("/goals/:id",
    authMiddleware,
    rbacMiddleware.RequirePermissionByResourceAction(
        rbacAppInstance.RBACService,
        "goals",
        "update",
        logger,
    ),
    goalController.UpdateGoal,
)
```

### 3. Using RBAC Service Directly

```go
// Check if user has permission
hasPermission, err := rbacAppInstance.RBACService.HasPermission(userID, "users.create")
if err != nil {
    // Handle error
}
if !hasPermission {
    // Return forbidden
}

// Check if user has role
hasRole, err := rbacAppInstance.RBACService.HasRole(userID, "admin")
if err != nil {
    // Handle error
}

// Get all user permissions
permissions, err := rbacAppInstance.RBACService.GetUserPermissions(userID)
if err != nil {
    // Handle error
}

// Assign role to user
err := rbacAppInstance.RBACService.AssignRoleToUser(userID, "moderator")
if err != nil {
    // Handle error
}

// Assign permission to role
err := rbacAppInstance.RBACService.AssignPermissionToRole("moderator", "users.read")
if err != nil {
    // Handle error
}
```

## Default Roles and Permissions

### Roles

1. **admin**: Full system access (all permissions)
2. **user**: Basic user access (goals and tasks CRUD)
3. **moderator**: Elevated access (goals, tasks, and read/update users)

### Permissions

#### User Management
- `users.create`: Create new users
- `users.read`: View users
- `users.update`: Update users
- `users.delete`: Delete users
- `users.manage_roles`: Assign/remove roles from users

#### Role Management
- `roles.create`: Create new roles
- `roles.read`: View roles
- `roles.update`: Update roles
- `roles.delete`: Delete roles
- `roles.assign_permissions`: Assign/remove permissions from roles

#### Permission Management
- `permissions.read`: View permissions

#### Goals Management
- `goals.create`: Create goals
- `goals.read`: View goals
- `goals.update`: Update goals
- `goals.delete`: Delete goals

#### Tasks Management
- `tasks.create`: Create tasks
- `tasks.read`: View tasks
- `tasks.update`: Update tasks
- `tasks.delete`: Delete tasks

## Best Practices

1. **Always use AuthMiddleware first**: Authorization middleware requires the user to be authenticated first
2. **Use specific permissions**: Prefer specific permissions over roles for fine-grained control
3. **Cache user permissions**: For high-traffic endpoints, consider caching user permissions
4. **System roles/permissions**: System roles and permissions (marked with `is_system=true`) cannot be deleted
5. **Default role assignment**: Consider assigning a default role (e.g., "user") to new users during registration

## Example Route Protection

```go
func setupRoutes(router *gin.Engine, rbacApp *rbacApp.RBACApp, logger logger.Logger) {
    // Public routes
    router.POST("/auth/signup", authController.SignUp)
    router.POST("/auth/login", authController.Login)
    
    // Protected routes with RBAC
    api := router.Group("/api/v1")
    api.Use(authMiddleware) // All routes below require authentication
    
    // User routes
    api.GET("/users",
        rbacMiddleware.RequirePermission(rbacApp.RBACService, "users.read", logger),
        userController.ListUsers,
    )
    api.POST("/users",
        rbacMiddleware.RequirePermission(rbacApp.RBACService, "users.create", logger),
        userController.CreateUser,
    )
    api.PUT("/users/:id",
        rbacMiddleware.RequirePermission(rbacApp.RBACService, "users.update", logger),
        userController.UpdateUser,
    )
    api.DELETE("/users/:id",
        rbacMiddleware.RequirePermission(rbacApp.RBACService, "users.delete", logger),
        userController.DeleteUser,
    )
    
    // Admin-only routes
    admin := api.Group("/admin")
    admin.Use(rbacMiddleware.RequireRole(rbacApp.RBACService, "admin", logger))
    {
        admin.GET("/dashboard", adminController.Dashboard)
        admin.POST("/roles", adminController.CreateRole)
    }
}
```

## Notes

- The RBAC system is integrated with the existing authentication system
- User roles are loaded on-demand when checking permissions
- System roles and permissions are protected from deletion
- The migration includes seed data for default roles and permissions
