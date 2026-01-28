# Architecture Guide: GOALKeeper Notion-like Workspace

**Last Updated**: January 27, 2026
**Status**: MVP Complete
**Audience**: Developers, architects, maintainers

---

## Table of Contents

1. [Overview](#overview)
2. [High-Level Architecture](#high-level-architecture)
3. [Backend Architecture](#backend-architecture)
4. [Frontend Architecture](#frontend-architecture)
5. [Data Model](#data-model)
6. [Communication Patterns](#communication-patterns)
7. [Deployment Architecture](#deployment-architecture)

---

## Overview

GOALKeeper is a web-based workspace application that allows users to create pages, add blocks (content items), and collaborate with others. It implements a layered architecture with clear separation of concerns.

**Core Principles**:
- **Layered Architecture**: Repository → Service → Controller pattern
- **Dependency Injection**: Services and controllers receive dependencies
- **REST API**: Standard HTTP verbs for CRUD operations
- **Concurrent-Safe**: Last-write-wins strategy for conflict resolution
- **Permission-Based**: Role-based access control (RBAC) at the page level

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       Frontend (React)                       │
│         (Next.js, TypeScript, Vitest, TailwindCSS)          │
└─────────────────┬───────────────────────────────────────────┘
                  │ JSON over HTTP/REST
                  │
┌─────────────────▼───────────────────────────────────────────┐
│                   API Gateway (Gin)                          │
│  (Middleware, routing, request/response handling)           │
└─────────────────┬───────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────────┐
│             Backend Services (Go Packages)                   │
│  ┌───────────┐ ┌───────────┐ ┌──────────┐ ┌──────────────┐  │
│  │ Workspace │ │   Page    │ │  Block   │ │  Permission  │  │
│  │ Service   │ │ Service   │ │ Service  │ │  Service     │  │
│  └────┬──────┘ └────┬──────┘ └─────┬────┘ └──────┬───────┘  │
│       │             │              │             │          │
│  ┌────▼──────────────▼──────────────▼─────────────▼────────┐ │
│  │   Repository Layer (GORM)                             │ │
│  │   (Database queries, transaction management)          │ │
│  └────┬──────────────────────────────────────────────────┘ │
└───────┼────────────────────────────────────────────────────┘
        │
┌───────▼──────────────────────────────────────────────────┐
│              Data Store (PostgreSQL)                      │
│  Tables: workspaces, pages, blocks, permissions, etc.   │
└───────────────────────────────────────────────────────────┘
```

---

## Backend Architecture

### Package Structure

```
backend/
├── cmd/
│   └── main.go           # Entry point, bootstraps application
├── internal/
│   ├── api/              # HTTP handlers and response formatting
│   │   ├── handler.go    # Base handler
│   │   └── helpers.go    # Response building utilities
│   ├── workspace/        # Workspace entity module
│   │   ├── model/        # Data models
│   │   ├── repository/   # Database layer
│   │   ├── service/      # Business logic
│   │   ├── controller/   # HTTP handlers
│   │   ├── dto/          # Request/response DTOs
│   │   └── router/       # Route registration
│   ├── page/             # Page entity module
│   │   ├── model/
│   │   ├── repository/
│   │   ├── service/
│   │   ├── controller/
│   │   ├── dto/
│   │   └── router/
│   ├── block/            # Block entity module
│   │   ├── model/
│   │   ├── repository/
│   │   ├── service/
│   │   ├── controller/
│   │   ├── dto/
│   │   └── router/
│   ├── rbac/             # Role-based access control
│   │   ├── model/        # Role, permission models
│   │   ├── repository/   # Permission queries
│   │   ├── service/      # Permission logic
│   │   └── middleware/   # Permission enforcement
│   ├── auth/             # Authentication (if extended)
│   ├── middleware/       # Global middleware
│   ├── logger/           # Logging
│   ├── errors/           # Error types and codes
│   ├── messages/         # Message constants
│   └── redis/            # Caching (optional)
├── config/               # Configuration loading
├── migrations/           # Database migrations (SQL)
└── tests/                # Test files
    ├── unit/
    ├── integration/
    └── contract/
```

### Layered Pattern: Repository → Service → Controller

Each entity (Workspace, Page, Block) follows this pattern:

#### 1. **Model Layer** (`model/`)
Defines data structures:
```go
type Workspace struct {
  ID        uuid.UUID `gorm:"primaryKey"`
  Name      string
  CreatedAt time.Time
  UpdatedAt time.Time
}
```

#### 2. **Repository Layer** (`repository/`)
Database operations:
```go
type WorkspaceRepository interface {
  Create(ctx context.Context, workspace *Workspace) error
  FindByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
  List(ctx context.Context, limit, offset int) ([]*Workspace, error)
  Update(ctx context.Context, workspace *Workspace) error
  Delete(ctx context.Context, id uuid.UUID) error
}
```

**Responsibilities**:
- Direct database queries via GORM
- Transaction management
- Query optimization
- No business logic

#### 3. **Service Layer** (`service/`)
Business logic:
```go
type WorkspaceService interface {
  CreateWorkspace(ctx context.Context, name string) (*Workspace, error)
  GetWorkspace(ctx context.Context, id uuid.UUID) (*Workspace, error)
  ListWorkspaces(ctx context.Context, limit, offset int) ([]*Workspace, error)
}
```

**Responsibilities**:
- Business logic validation
- Orchestration of multiple repositories
- Transaction coordination
- Permission checks (via RBAC service)

#### 4. **Controller Layer** (`controller/`)
HTTP handlers:
```go
func (c *WorkspaceController) Create(ctx *gin.Context) {
  // Request parsing
  var req CreateWorkspaceRequest
  
  // Business logic (via service)
  workspace, err := c.service.CreateWorkspace(ctx.Context, req.Name)
  
  // Response formatting
  ctx.JSON(201, workspace)
}
```

**Responsibilities**:
- HTTP request parsing
- Calling services
- Response formatting
- Error handling

#### 5. **Router Layer** (`router/`)
Route registration:
```go
func RegisterWorkspaceRoutes(engine *gin.Engine, controller *WorkspaceController) {
  group := engine.Group("/api/v1/notion/workspaces")
  group.POST("", controller.Create)
  group.GET("", controller.List)
  group.GET("/:id", controller.Get)
}
```

### Dependency Injection

Dependencies are injected at application startup:

```go
// In main.go
db := setupDatabase()
wsRepo := workspace.NewRepository(db)
wsService := workspace.NewService(wsRepo)
wsController := workspace.NewController(wsService)
workspace.RegisterRoutes(engine, wsController)
```

**Benefits**:
- Testability (inject mocks)
- Flexibility (swap implementations)
- Clear dependencies

### RBAC (Role-Based Access Control)

Permission checking happens in the service layer:

```go
// In BlockService.UpdateBlock()
if err := c.rbac.CheckPermission(ctx, blockID, userID, "edit"); err != nil {
  return err // 403 Forbidden
}
```

**Permission Levels**:
- `viewer`: Read-only
- `editor`: Read + Write
- `owner`: Full control

**Enforcement Points**:
- Service methods validate permission before operation
- Repository never bypasses permission (only called from service)
- Controller validates user identity via header

---

## Frontend Architecture

### Directory Structure

```
frontend/src/
├── app/                    # Next.js app directory
│   ├── layout.tsx          # Root layout
│   ├── page.tsx            # Home page
│   ├── middleware.ts       # Request middleware
│   ├── api/                # API routes (optional)
│   ├── auth/               # Auth pages
│   ├── workspace/          # Workspace pages
│   │   ├── [id]/           # Dynamic workspace
│   │   └── layout.tsx      # Workspace layout
│   └── (auth)/             # Grouped auth routes
├── components/             # Reusable UI components
│   ├── BlockList.tsx       # Block list component
│   ├── BlockEditor.tsx     # Block editor
│   ├── WorkspaceSidebar.tsx
│   ├── WorkspacePageShell.tsx
│   └── ui/                 # Base UI components (buttons, inputs)
├── features/               # Feature-specific logic
│   ├── auth/               # Authentication feature
│   │   ├── hooks/          # useAuth, useSession
│   │   └── context/        # Auth context
│   ├── workspace/
│   ├── block/
│   └── page/
├── pages/                  # Page components (not app dir)
│   └── WorkspacePageView.tsx
├── shared/                 # Shared utilities
│   ├── constant/           # Constants
│   ├── hooks/              # useQuery, useMutation hooks
│   ├── types/              # TypeScript types
│   ├── api/                # API client
│   └── utils/              # Utilities
└── tests/
    ├── unit/
    ├── integration/
    └── contract/
```

### Component Architecture

**Hierarchical Structure**:
```
<App>
  ├── <WorkspaceSidebar>         # List workspaces, pages
  │   └── <PageTree>             # Hierarchical page display
  │       └── <PageTreeItem>     # Individual page
  └── <WorkspacePageShell>       # Main content area
      └── <BlockList>            # Editable blocks
          ├── <BlockItem>        # Single block
          │   ├── <BlockEditor>  # Edit mode
          │   └── <BlockView>    # View mode
          └── <BlockControls>    # Add, delete, reorder
```

### State Management

- **React Context**: Authentication, workspace selection
- **React Query/SWR**: Server state (fetching, caching)
- **useState**: Local component state

### Data Flow

1. **User Action** (e.g., edit block)
2. **Event Handler** calls mutation
3. **API Client** sends request to backend
4. **Service Response** returned with new state
5. **UI Rerender** with updated data

---

## Data Model

### Entity Relationship Diagram

```
Workspace (1) ──────────────── (N) Page
    │                            │
    └──────────────────────┬─────┘
                           │
                           │ (1) ──────────────── (N) Block
                           │
                      Page
                    (1) │
                        │ (N)
                        ▼
                      Page (child)

User (1) ──────────────── (N) Permission
            user_id       page_id, role

Page (1) ──────────────── (N) Permission
```

### Tables

#### Workspaces
```sql
CREATE TABLE workspaces (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Pages
```sql
CREATE TABLE pages (
  id UUID PRIMARY KEY,
  workspace_id UUID NOT NULL REFERENCES workspaces(id),
  parent_id UUID REFERENCES pages(id),
  title VARCHAR(500) NOT NULL,
  order INTEGER NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Blocks
```sql
CREATE TABLE blocks (
  id UUID PRIMARY KEY,
  page_id UUID NOT NULL REFERENCES pages(id),
  type VARCHAR(50) NOT NULL,
  content TEXT,
  checked BOOLEAN,
  order INTEGER NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Permissions
```sql
CREATE TABLE page_permissions (
  id UUID PRIMARY KEY,
  page_id UUID NOT NULL REFERENCES pages(id),
  user_id UUID NOT NULL,
  role VARCHAR(20) NOT NULL, -- viewer, editor, owner
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Communication Patterns

### Request/Response Flow

```
Frontend                      Backend
   │                             │
   ├─ POST /pages/create ───────>│
   │  (Create page request)       │
   │                             │
   │                      ┌──────────────┐
   │                      │ Controller   │
   │                      ├──────────────┤
   │                      │ Service      │ ← Permission check
   │                      ├──────────────┤
   │                      │ Repository   │ ← Database query
   │                      └──────────────┘
   │                             │
   │<─ 201 Created ────────────┤
   │  (Created page data)       │
   │                             │
```

### Error Handling

All errors follow standard HTTP status codes:

```
400 Bad Request      ← Invalid input
401 Unauthorized     ← Missing X-User-ID
403 Forbidden        ← Permission denied
404 Not Found        ← Resource not found
500 Server Error     ← Unexpected error
```

Error response format:
```json
{
  "error": "NO_PERMISSION",
  "message": "User lacks edit permission",
  "details": { "resource": "page", "required_role": "editor" }
}
```

---

## Deployment Architecture

### Local Development

```
Frontend (localhost:3000)
    ↓ (REST API calls)
Backend API (localhost:8080)
    ↓ (SQL queries)
PostgreSQL (localhost:5432)
```

### Docker Compose

```yaml
services:
  postgres:
    image: postgres:15
    ports: 5432
  backend:
    build: ./backend
    ports: 8080
    depends_on: postgres
  frontend:
    build: ./frontend
    ports: 3000
    depends_on: backend
```

### Production Deployment

**Recommended**:
- **Frontend**: Deploy to Vercel, Netlify, or cloud CDN
- **Backend**: Deploy to cloud run, Kubernetes, or managed app platform
- **Database**: Managed PostgreSQL (RDS, Cloud SQL, Azure Database)

**Environment Configuration**:
- API base URL injected into frontend
- Database credentials via environment variables
- CORS settings configured per environment

---

## Performance Considerations

### Query Optimization

1. **Index Strategy**:
   - `pages.workspace_id` (list pages by workspace)
   - `pages.parent_id` (hierarchical queries)
   - `blocks.page_id` (list blocks by page)
   - `permissions.page_id, user_id` (permission checks)

2. **Pagination**:
   - List endpoints support limit/offset
   - Default limit: 50, max: 500

3. **Eager Loading**:
   - Pages load with child pages (nested)
   - Avoid N+1 queries in hierarchical queries

### Caching Strategy

- **Frontend**: React Query with 5-minute stale time
- **Backend**: Redis (optional, for permission checks)
- **CDN**: Static assets (CSS, JavaScript)

### Concurrency Handling

- **Strategy**: Last-write-wins (no conflict merging)
- **Rationale**: Simpler implementation, good for single-user and low-contention scenarios
- **Future**: Operational transformation (OT) or CRDT if needed

---

## Security Architecture

### Authentication

- User identity via `X-User-ID` header (UUID)
- No session/JWT in MVP (trust infrastructure)
- Future: Add JWT/OAuth2 for multi-tenant security

### Authorization (RBAC)

- Permissions stored per page
- Checked in service layer before data access
- Cascade: Child pages inherit parent permissions

### Data Protection

- Passwords hashed (if auth added)
- No sensitive data in URLs
- HTTPS in production (configured at reverse proxy)

---

## Testing Strategy

### Unit Tests
- Test individual functions (repositories, services)
- Mock external dependencies
- Example: `TestCreateWorkspace`

### Integration Tests
- Test full flow (controller → service → repository → database)
- Use test database
- Example: `TestCreatePageWithBlocks`

### Contract Tests
- Test API against specification
- Verify request/response formats
- Example: `TestPageCreateResponse`

---

## Extensibility

### Adding a New Entity (e.g., Team)

1. Create `internal/team/model/team.go`
2. Create `internal/team/repository/repository.go`
3. Create `internal/team/service/service.go`
4. Create `internal/team/controller/controller.go`
5. Create `internal/team/router/router.go`
6. Register routes in `main.go`
7. Add database migration

### Adding a New Middleware

1. Create middleware in `internal/middleware/`
2. Register in router setup in `main.go`

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Layered architecture | Clear separation, testability, maintainability |
| GORM ORM | Type-safe, expressive queries, built-in hooks |
| REST API | Standard, HTTP caching, easy to test |
| Last-write-wins | Simplicity, good for MVP, low contention |
| Role-based access control | Scalable, easy to understand, standard |
| Context for cancellation | Proper request lifecycle, timeout handling |
| Dependency injection | Testability, flexibility |

---

## Future Improvements

1. **Performance**: Add caching layer (Redis)
2. **Concurrency**: Implement CRDT for collaborative editing
3. **Security**: Add JWT/OAuth2 for proper auth
4. **Monitoring**: Add metrics, tracing, structured logging
5. **Features**: Real-time collaboration, comments, version history
6. **Scale**: Read replicas, query optimization, horizontal scaling

