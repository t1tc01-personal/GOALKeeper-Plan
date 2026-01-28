# API Reference: GOALKeeper Notion-like Workspace

**Version**: 1.0
**Base URL**: `http://localhost:8080/api/v1/notion`
**Authentication**: `X-User-ID` header (UUID format)
**Content-Type**: `application/json`

---

## Table of Contents

1. [Workspaces](#workspaces)
2. [Pages](#pages)
3. [Blocks](#blocks)
4. [Sharing & Permissions](#sharing--permissions)
5. [Error Codes](#error-codes)

---

## Workspaces

### Create Workspace

**Endpoint**: `POST /workspaces`

**Request**:
```json
{
  "name": "My Workspace"
}
```

**Response**: `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Workspace",
  "created_at": "2026-01-27T10:00:00Z",
  "updated_at": "2026-01-27T10:00:00Z"
}
```

**Error**: `400 Bad Request` (invalid name), `401 Unauthorized` (no user ID)

---

### List Workspaces

**Endpoint**: `GET /workspaces`

**Query Parameters**:
- `limit` (optional, default: 50): Number of workspaces to return
- `offset` (optional, default: 0): Pagination offset

**Response**: `200 OK`
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "My Workspace",
      "created_at": "2026-01-27T10:00:00Z",
      "updated_at": "2026-01-27T10:00:00Z"
    }
  ],
  "total": 1
}
```

---

### Get Workspace

**Endpoint**: `GET /workspaces/:workspace_id`

**Response**: `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Workspace",
  "created_at": "2026-01-27T10:00:00Z",
  "updated_at": "2026-01-27T10:00:00Z"
}
```

**Error**: `404 Not Found`, `403 Forbidden` (no access)

---

## Pages

### Create Page

**Endpoint**: `POST /pages`

**Request**:
```json
{
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "My Page",
  "parent_id": null
}
```

**Response**: `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "My Page",
  "parent_id": null,
  "order": 1,
  "created_at": "2026-01-27T10:00:00Z",
  "updated_at": "2026-01-27T10:00:00Z"
}
```

**Error**: `400 Bad Request` (missing fields), `403 Forbidden` (no write permission)

---

### List Pages

**Endpoint**: `GET /pages`

**Query Parameters**:
- `workspace_id` (required): Workspace UUID
- `parent_id` (optional): Filter by parent page

**Response**: `200 OK`
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "My Page",
      "parent_id": null,
      "order": 1,
      "created_at": "2026-01-27T10:00:00Z",
      "updated_at": "2026-01-27T10:00:00Z"
    }
  ],
  "total": 1
}
```

---

### Get Page

**Endpoint**: `GET /pages/:page_id`

**Response**: `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "My Page",
  "parent_id": null,
  "order": 1,
  "created_at": "2026-01-27T10:00:00Z",
  "updated_at": "2026-01-27T10:00:00Z",
  "children": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "title": "Subpage",
      "parent_id": "550e8400-e29b-41d4-a716-446655440001",
      "order": 1
    }
  ]
}
```

**Requires**: Read permission on page
**Error**: `403 Forbidden` (no read permission), `404 Not Found`

---

### Update Page

**Endpoint**: `PUT /pages/:page_id`

**Request**:
```json
{
  "title": "Updated Title",
  "order": 2
}
```

**Response**: `200 OK` (returns updated page)

**Requires**: Edit permission on page
**Error**: `403 Forbidden` (no edit permission)

---

### Delete Page

**Endpoint**: `DELETE /pages/:page_id`

**Response**: `204 No Content`

**Cascade**: Deletes all child pages and their blocks

**Requires**: Owner permission on page
**Error**: `403 Forbidden` (not owner)

---

### Reorder Pages

**Endpoint**: `PUT /pages/:page_id/order`

**Request**:
```json
{
  "order": 2
}
```

**Response**: `200 OK`

**Note**: Automatically reorders siblings

---

## Blocks

### Create Block

**Endpoint**: `POST /pages/:page_id/blocks`

**Request**:
```json
{
  "type": "paragraph",
  "content": "Block content",
  "order": 1
}
```

**Block Types**:
- `paragraph`: Plain text content
- `heading`: With optional `level` (1-3)
- `checklist`: With `checked` boolean

**Response**: `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440003",
  "page_id": "550e8400-e29b-41d4-a716-446655440001",
  "type": "paragraph",
  "content": "Block content",
  "order": 1,
  "checked": null,
  "created_at": "2026-01-27T10:00:00Z",
  "updated_at": "2026-01-27T10:00:00Z"
}
```

**Requires**: Edit permission on page
**Error**: `400 Bad Request` (invalid type), `403 Forbidden` (no edit permission)

---

### List Blocks

**Endpoint**: `GET /pages/:page_id/blocks`

**Query Parameters**:
- `limit` (optional, default: 100)
- `offset` (optional, default: 0)

**Response**: `200 OK`
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "page_id": "550e8400-e29b-41d4-a716-446655440001",
      "type": "paragraph",
      "content": "Block content",
      "order": 1,
      "checked": null,
      "created_at": "2026-01-27T10:00:00Z",
      "updated_at": "2026-01-27T10:00:00Z"
    }
  ],
  "total": 1
}
```

**Requires**: Read permission on page

---

### Get Block

**Endpoint**: `GET /pages/:page_id/blocks/:block_id`

**Response**: `200 OK` (single block object)

**Requires**: Read permission on page

---

### Update Block

**Endpoint**: `PUT /pages/:page_id/blocks/:block_id`

**Request**:
```json
{
  "content": "Updated content",
  "checked": true,
  "order": 2
}
```

**Response**: `200 OK`

**Conflict Handling**: Last-write-wins (no merging, last update overwrites)

**Requires**: Edit permission on page
**Error**: `409 Conflict` (if timestamps differ significantly - not applicable in last-write-wins)

---

### Delete Block

**Endpoint**: `DELETE /pages/:page_id/blocks/:block_id`

**Response**: `204 No Content`

**Cascade**: Deletes block history/deltas

**Requires**: Edit permission on page

---

### Reorder Blocks

**Endpoint**: `PUT /pages/:page_id/blocks/:block_id/order`

**Request**:
```json
{
  "order": 3
}
```

**Response**: `200 OK`

**Note**: Automatically reorders siblings within page

---

## Sharing & Permissions

### Grant Access

**Endpoint**: `POST /pages/:page_id/share`

**Request**:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440004",
  "role": "viewer"
}
```

**Roles**:
- `viewer`: Read-only access
- `editor`: Read and write access
- `owner`: Full control (including share/delete)

**Response**: `201 Created` or `200 OK` (if updating)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440005",
  "page_id": "550e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440004",
  "role": "viewer",
  "created_at": "2026-01-27T10:00:00Z",
  "updated_at": "2026-01-27T10:00:00Z"
}
```

**Requires**: Owner permission on page
**Error**: `400 Bad Request` (invalid role), `403 Forbidden` (not owner)

---

### Revoke Access

**Endpoint**: `DELETE /pages/:page_id/share/:user_id`

**Response**: `204 No Content`

**Cascade**: User loses all permissions on page

**Requires**: Owner permission on page
**Error**: `403 Forbidden` (not owner)

---

### List Collaborators

**Endpoint**: `GET /pages/:page_id/collaborators`

**Response**: `200 OK`
```json
[
  {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "role": "owner",
    "created_at": "2026-01-27T10:00:00Z"
  },
  {
    "user_id": "550e8400-e29b-41d4-a716-446655440004",
    "role": "viewer",
    "created_at": "2026-01-27T10:00:00Z"
  }
]
```

**Requires**: Read permission on page

---

## Error Codes

### Standard HTTP Status Codes

- `200 OK`: Request succeeded
- `201 Created`: Resource created
- `204 No Content`: Success, no response body
- `400 Bad Request`: Invalid request (malformed JSON, missing required fields)
- `401 Unauthorized`: Missing or invalid `X-User-ID` header
- `403 Forbidden`: Authenticated but insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Concurrent modification conflict (rare with last-write-wins)
- `500 Internal Server Error`: Server error

---

### Error Response Format

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": {
    "field": "reason"
  }
}
```

### Common Error Codes

| Code | Status | Meaning |
|------|--------|---------|
| `INVALID_REQUEST` | 400 | Malformed request |
| `INVALID_PAGE_ID` | 400 | Page ID not valid UUID |
| `INVALID_ROLE` | 400 | Role not in [viewer, editor, owner] |
| `MISSING_HEADER` | 401 | X-User-ID header missing |
| `NO_PERMISSION` | 403 | User lacks required permission |
| `NOT_OWNER` | 403 | Operation requires owner role |
| `NOT_FOUND` | 404 | Resource doesn't exist |
| `ALREADY_EXISTS` | 409 | Duplicate resource (e.g., permission) |

---

## Authentication

### User Identification

All requests require the `X-User-ID` header with a valid UUID:

```
X-User-ID: 550e8400-e29b-41d4-a716-446655440000
```

**In curl**:
```bash
curl -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  http://localhost:8080/api/v1/notion/pages
```

**In fetch/JavaScript**:
```javascript
fetch('http://localhost:8080/api/v1/notion/pages', {
  headers: {
    'X-User-ID': '550e8400-e29b-41d4-a716-446655440000',
    'Content-Type': 'application/json'
  }
})
```

---

## Permission Model

### Permission Hierarchy

```
Owner    ⊇ Editor   ⊇ Viewer
(full)     (read+write)  (read-only)
```

### Endpoint Permission Requirements

| Endpoint | Method | Viewer | Editor | Owner | Notes |
|----------|--------|--------|--------|-------|-------|
| Pages | GET | ✓ | ✓ | ✓ | Read permission required |
| Pages | POST | ✗ | ✓ | ✓ | Edit permission required |
| Pages | PUT | ✗ | ✓ | ✓ | Edit permission required |
| Pages | DELETE | ✗ | ✗ | ✓ | Owner permission required |
| Blocks | GET | ✓ | ✓ | ✓ | Read permission required |
| Blocks | POST | ✗ | ✓ | ✓ | Edit permission required |
| Blocks | PUT | ✗ | ✓ | ✓ | Edit permission required |
| Blocks | DELETE | ✗ | ✓ | ✓ | Edit permission required |
| Share | GET | ✓ | ✓ | ✓ | Read permission required |
| Share | POST | ✗ | ✗ | ✓ | Owner permission required |
| Share | DELETE | ✗ | ✗ | ✓ | Owner permission required |

---

## Rate Limiting

Not implemented in MVP. Add in future phase if needed.

---

## Pagination

List endpoints support pagination via query parameters:

- `limit`: Items per page (default: 50, max: 500)
- `offset`: Number of items to skip (default: 0)

Example:
```
GET /pages?workspace_id=...&limit=25&offset=50
```

Response includes `total` count.

---

## Timestamps

All timestamps are ISO 8601 format with UTC timezone:

```
"2026-01-27T10:00:00Z"
```

---

## Versioning

API version is specified in the base URL path: `/api/v1/`

Future versions (`/api/v2/`) will be supported alongside `/api/v1/`.

---

## Implementation Notes

- All IDs are UUIDs (36 characters with hyphens)
- All responses are JSON
- All requests expecting JSON body must include `Content-Type: application/json`
- Concurrent edits use last-write-wins strategy (no conflict merging)
- Permissions are inherited from pages to blocks (block permissions follow page permissions)
- Deleting a page deletes all child pages recursively

