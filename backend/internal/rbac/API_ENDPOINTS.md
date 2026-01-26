# RBAC API Endpoints

This document describes all available API endpoints for Role and Permission management.

## Base URL
All endpoints are prefixed with `/api/v1`

## Role Management API

### Create Role
- **POST** `/roles`
- **Description**: Create a new role
- **Request Body**:
  ```json
  {
    "name": "editor",
    "description": "Editor role with content management permissions"
  }
  ```
- **Response**: `201 Created`
  ```json
  {
    "status": "success",
    "message": "Role created successfully",
    "data": {
      "id": "uuid",
      "name": "editor",
      "description": "Editor role with content management permissions",
      "is_system": false,
      "created_at": "2026-01-26T00:00:00Z",
      "updated_at": "2026-01-26T00:00:00Z"
    }
  }
  ```

### List Roles
- **GET** `/roles?limit=10&offset=0`
- **Description**: Get a paginated list of roles
- **Query Parameters**:
  - `limit` (optional, default: 10): Number of roles to return
  - `offset` (optional, default: 0): Number of roles to skip
- **Response**: `200 OK`

### Get Role by ID
- **GET** `/roles/{id}`
- **Description**: Get role information by ID
- **Response**: `200 OK`

### Get Role by Name
- **GET** `/roles/name?name=admin`
- **Description**: Get role information by name
- **Query Parameters**:
  - `name` (required): Role name
- **Response**: `200 OK`

### Get Role with Permissions
- **GET** `/roles/{id}/permissions`
- **Description**: Get role information with all assigned permissions
- **Response**: `200 OK`
  ```json
  {
    "status": "success",
    "message": "Role retrieved successfully",
    "data": {
      "id": "uuid",
      "name": "admin",
      "description": "Administrator role",
      "is_system": true,
      "permissions": [
        {
          "id": "uuid",
          "name": "users.create",
          "resource": "users",
          "action": "create",
          "description": "Create new users",
          "is_system": true,
          "created_at": "2026-01-26T00:00:00Z",
          "updated_at": "2026-01-26T00:00:00Z"
        }
      ],
      "created_at": "2026-01-26T00:00:00Z",
      "updated_at": "2026-01-26T00:00:00Z"
    }
  }
  ```

### Update Role
- **PUT** `/roles/{id}`
- **PATCH** `/roles/{id}`
- **Description**: Update role information
- **Request Body**:
  ```json
  {
    "name": "updated_name",
    "description": "Updated description"
  }
  ```
- **Response**: `200 OK`

### Delete Role
- **DELETE** `/roles/{id}`
- **Description**: Delete a role (system roles cannot be deleted)
- **Response**: `200 OK`

### Assign Permission to Role
- **POST** `/roles/{id}/permissions`
- **Description**: Assign a permission to a role
- **Request Body**:
  ```json
  {
    "permission_name": "users.create"
  }
  ```
- **Response**: `200 OK`

### Remove Permission from Role
- **DELETE** `/roles/{id}/permissions`
- **Description**: Remove a permission from a role
- **Request Body**:
  ```json
  {
    "permission_name": "users.create"
  }
  ```
- **Response**: `200 OK`

## Permission Management API

### Create Permission
- **POST** `/permissions`
- **Description**: Create a new permission
- **Request Body**:
  ```json
  {
    "name": "posts.create",
    "resource": "posts",
    "action": "create",
    "description": "Create new posts"
  }
  ```
- **Response**: `201 Created`
  ```json
  {
    "status": "success",
    "message": "Permission created successfully",
    "data": {
      "id": "uuid",
      "name": "posts.create",
      "resource": "posts",
      "action": "create",
      "description": "Create new posts",
      "is_system": false,
      "created_at": "2026-01-26T00:00:00Z",
      "updated_at": "2026-01-26T00:00:00Z"
    }
  }
  ```

### List Permissions
- **GET** `/permissions?limit=10&offset=0`
- **Description**: Get a paginated list of permissions
- **Query Parameters**:
  - `limit` (optional, default: 10): Number of permissions to return
  - `offset` (optional, default: 0): Number of permissions to skip
- **Response**: `200 OK`

### Get Permission by ID
- **GET** `/permissions/{id}`
- **Description**: Get permission information by ID
- **Response**: `200 OK`

### Get Permission by Name
- **GET** `/permissions/name?name=users.create`
- **Description**: Get permission information by name
- **Query Parameters**:
  - `name` (required): Permission name
- **Response**: `200 OK`

### Get Permission by Resource and Action
- **GET** `/permissions/resource-action?resource=users&action=create`
- **Description**: Get permission information by resource and action
- **Query Parameters**:
  - `resource` (required): Resource name
  - `action` (required): Action name
- **Response**: `200 OK`

### List Permissions by Resource
- **GET** `/permissions/resource?resource=users`
- **Description**: Get all permissions for a specific resource
- **Query Parameters**:
  - `resource` (required): Resource name
- **Response**: `200 OK`

### Update Permission
- **PUT** `/permissions/{id}`
- **PATCH** `/permissions/{id}`
- **Description**: Update permission information (system permissions cannot be updated)
- **Request Body**:
  ```json
  {
    "name": "updated_name",
    "resource": "updated_resource",
    "action": "updated_action",
    "description": "Updated description"
  }
  ```
- **Response**: `200 OK`

### Delete Permission
- **DELETE** `/permissions/{id}`
- **Description**: Delete a permission (system permissions cannot be deleted)
- **Response**: `200 OK`

## Error Responses

All endpoints return errors in the following format:

```json
{
  "status": "error",
  "type": "validation|not_found|conflict|internal",
  "code": "ERROR_CODE",
  "message": "Error message",
  "details": "Additional error details",
  "timestamp": "2026-01-26T00:00:00Z"
}
```

### Common Error Codes

- `ROLE_NOT_FOUND`: Role does not exist
- `ROLE_EXISTS`: Role with the same name already exists
- `SYSTEM_ROLE_CANNOT_DELETE`: Attempted to delete a system role
- `PERMISSION_NOT_FOUND`: Permission does not exist
- `PERMISSION_EXISTS`: Permission with the same name already exists
- `SYSTEM_PERMISSION_CANNOT_DELETE`: Attempted to delete a system permission
- `FAILED_TO_ASSIGN_PERMISSION`: Failed to assign permission to role
- `FAILED_TO_REMOVE_PERMISSION`: Failed to remove permission from role

## Notes

1. **System Roles/Permissions**: System roles and permissions (marked with `is_system: true`) cannot be deleted or have their names changed. They are protected from modification.

2. **Authentication**: All endpoints require authentication. Use the `Authorization: Bearer <token>` header.

3. **Authorization**: Consider protecting these endpoints with RBAC middleware:
   - Role management: `roles.create`, `roles.read`, `roles.update`, `roles.delete`, `roles.assign_permissions`
   - Permission management: `permissions.read` (create/update/delete can be restricted to admins)

4. **Validation**: All request bodies are validated. Invalid requests return `400 Bad Request` with validation error details.
