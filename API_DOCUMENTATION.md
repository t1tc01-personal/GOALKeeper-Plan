# API Documentation: Notion-like Workspace

**Version**: 1.0  
**Base URL**: `http://localhost:8080/api`  
**Authentication**: Bearer token (JWT)

## Authentication

### Login (OAuth)
```
GET /auth/google
GET /auth/github
POST /auth/callback
```

### Refresh Token
```
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "..."
}
```

---

## Workspaces

### List Workspaces
```
GET /notion/workspaces

Response 200:
{
  "message": "Workspaces retrieved successfully",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "My Workspace",
      "description": "Personal workspace",
      "created_at": "2026-01-28T10:00:00Z",
      "updated_at": "2026-01-28T10:00:00Z"
    }
  ]
}
```

### Create Workspace
```
POST /notion/workspaces
Content-Type: application/json

{
  "name": "New Workspace",
  "description": "Optional description"
}

Response 201:
{
  "message": "Workspace created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "New Workspace",
    "description": "Optional description",
    "created_at": "2026-01-28T10:00:00Z",
    "updated_at": "2026-01-28T10:00:00Z"
  }
}

Error 400:
{
  "code": "INVALID_INPUT",
  "message": "Name is required"
}

Error 500:
{
  "code": "FAILED_TO_CREATE_WORKSPACE",
  "message": "Failed to create workspace"
}
```

### Get Workspace
```
GET /notion/workspaces/:id

Response 200:
{
  "message": "Workspace retrieved successfully",
  "data": { ... }
}

Error 404:
{
  "code": "WORKSPACE_NOT_FOUND",
  "message": "Workspace not found"
}
```

### Update Workspace
```
PUT /notion/workspaces/:id
Content-Type: application/json

{
  "name": "Updated Name",
  "description": "Updated description"
}

Response 200:
{
  "message": "Workspace updated successfully",
  "data": { ... }
}
```

### Delete Workspace
```
DELETE /notion/workspaces/:id

Response 204: No content

Error 404: Workspace not found
```

---

## Pages

### List Pages
```
GET /notion/pages?workspaceId=:id

Query Parameters:
- workspaceId (required): UUID of workspace

Response 200:
{
  "message": "Pages retrieved successfully",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "workspaceId": "550e8400-e29b-41d4-a716-446655440000",
      "title": "My Page",
      "created_at": "2026-01-28T10:00:00Z",
      "updated_at": "2026-01-28T10:00:00Z"
    }
  ]
}

Error 400:
{
  "code": "MISSING_REQUIRED",
  "message": "Workspace ID is required"
}
```

### Create Page
```
POST /notion/pages
Content-Type: application/json

{
  "workspaceId": "550e8400-e29b-41d4-a716-446655440000",
  "title": "New Page",
  "parentPageId": "550e8400-e29b-41d4-a716-446655440001" (optional)
}

Response 201:
{
  "message": "Page created successfully",
  "data": { ... }
}
```

### Get Page
```
GET /notion/pages/:id

Response 200:
{
  "message": "Page retrieved successfully",
  "data": { ... }
}
```

### Update Page
```
PUT /notion/pages/:id
Content-Type: application/json

{
  "title": "Updated Title",
  "viewConfig": {}
}

Response 200:
{
  "message": "Page updated successfully",
  "data": { ... }
}
```

### Delete Page
```
DELETE /notion/pages/:id

Response 204: No content
```

### Get Page Hierarchy
```
GET /notion/pages/hierarchy?workspaceId=:id

Returns hierarchical structure of all pages in workspace.

Response 200:
{
  "message": "Page hierarchy retrieved successfully",
  "data": [
    {
      "id": "...",
      "title": "Root Page",
      "children": [
        {
          "id": "...",
          "title": "Child Page",
          "children": []
        }
      ]
    }
  ]
}
```

---

## Blocks

### List Blocks
```
GET /notion/blocks?pageId=:id

Query Parameters:
- pageId (required): UUID of page

Response 200:
{
  "message": "Blocks retrieved successfully",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "pageId": "550e8400-e29b-41d4-a716-446655440001",
      "type": "text",
      "content": "Block content",
      "position": 0,
      "blockConfig": {},
      "created_at": "2026-01-28T10:00:00Z",
      "updated_at": "2026-01-28T10:00:00Z"
    }
  ]
}
```

### Create Block
```
POST /notion/blocks
Content-Type: application/json

{
  "pageId": "550e8400-e29b-41d4-a716-446655440001",
  "type": "text|heading|checklist",
  "content": "Block content",
  "position": 0,
  "blockConfig": {}
}

Response 201:
{
  "message": "Block created successfully",
  "data": { ... }
}
```

### Get Block
```
GET /notion/blocks/:id

Response 200:
{
  "message": "Block retrieved successfully",
  "data": { ... }
}
```

### Update Block
```
PUT /notion/blocks/:id
Content-Type: application/json

{
  "content": "Updated content",
  "blockConfig": {}
}

Response 200:
{
  "message": "Block updated successfully",
  "data": { ... }
}
```

### Delete Block
```
DELETE /notion/blocks/:id

Response 204: No content
```

### Reorder Blocks
```
POST /notion/blocks/reorder?pageId=:id
Content-Type: application/json

{
  "blockIds": [
    "id1",
    "id2",
    "id3"
  ]
}

Response 200:
{
  "message": "Blocks reordered successfully"
}

Error 400:
{
  "code": "INVALID_INPUT",
  "message": "Invalid block IDs provided"
}
```

---

## Sharing

### Grant Permission
```
POST /notion/sharing/grant
Content-Type: application/json

{
  "pageId": "550e8400-e29b-41d4-a716-446655440001",
  "granteeId": "550e8400-e29b-41d4-a716-446655440003",
  "role": "viewer|editor"
}

Response 201:
{
  "message": "Permission granted successfully",
  "data": {
    "id": "...",
    "pageId": "...",
    "granteeId": "...",
    "role": "editor",
    "created_at": "..."
  }
}

Error 403:
{
  "code": "PERMISSION_DENIED",
  "message": "You don't have permission to share this page"
}
```

### Revoke Permission
```
DELETE /notion/sharing/revoke/:permissionId

Response 204: No content

Error 403: You don't have permission to revoke
```

### List Collaborators
```
GET /notion/sharing/collaborators?pageId=:id

Response 200:
{
  "message": "Collaborators retrieved successfully",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "email": "user@example.com",
      "name": "User Name",
      "role": "editor",
      "sharedAt": "2026-01-28T10:00:00Z"
    }
  ]
}
```

---

## Error Responses

### Standard Error Format
```json
{
  "code": "ERROR_CODE",
  "message": "Human readable message"
}
```

### Common Error Codes

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| INVALID_ID | 400 | Invalid UUID format |
| MISSING_REQUIRED | 400 | Required field missing |
| INVALID_INPUT | 400 | Input validation failed |
| WORKSPACE_NOT_FOUND | 404 | Workspace doesn't exist |
| PAGE_NOT_FOUND | 404 | Page doesn't exist |
| BLOCK_NOT_FOUND | 404 | Block doesn't exist |
| PERMISSION_DENIED | 403 | User lacks permission |
| FAILED_TO_CREATE_* | 500 | Service operation failed |
| INTERNAL_ERROR | 500 | Server error |

---

## Request/Response Schema

### WorkspaceResponse
```typescript
{
  id: string (UUID)
  name: string
  description: string (nullable)
  created_at: ISO8601 timestamp
  updated_at: ISO8601 timestamp
}
```

### PageResponse
```typescript
{
  id: string (UUID)
  workspaceId: string (UUID)
  title: string
  created_at: ISO8601 timestamp
  updated_at: ISO8601 timestamp
}
```

### BlockResponse
```typescript
{
  id: string (UUID)
  pageId: string (UUID)
  type: string
  content: string (nullable)
  position: integer
  blockConfig: object (JSON)
  created_at: ISO8601 timestamp
  updated_at: ISO8601 timestamp
}
```

---

## Rate Limiting

- **API**: 100 requests per minute per user
- **Auth**: 5 login attempts per minute per IP

Headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1706351425
```

---

## CORS

Allowed origins configured in `.env`:
- `ALLOWED_ORIGINS=http://localhost:3000,https://app.example.com`

---

## Pagination (Future)

```
GET /notion/pages?workspaceId=:id&page=1&pageSize=20

Response includes:
{
  "data": [...],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 45,
    "totalPages": 3
  }
}
```

