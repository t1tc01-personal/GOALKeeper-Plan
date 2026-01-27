# Implementation Guide: Notion-like Workspace App

**Version**: 1.0  
**Date**: 2026-01-28  
**Feature**: GOALKeeper Notion-like Workspace  
**Status**: MVP Complete (Phases 1-5), Polish Phase In Progress

## Quick Start

### For Developers

#### Backend Setup
```bash
cd backend

# Install dependencies
go mod download

# Run migrations
make migrate-up

# Start the server
go run ./cmd/api/main.go
```

Server runs on `http://localhost:8080`

#### Frontend Setup
```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

Frontend runs on `http://localhost:3000`

#### Run Tests
```bash
cd backend
make test              # All tests
make test-unit         # Unit tests only
make test-integration  # Integration tests only
make test-contract     # Contract tests only
```

### For End Users

See [Quickstart Guide](./quickstart.md) for user workflows and feature overview.

---

## Architecture Overview

### Backend Structure

```
backend/
├── cmd/api/
│   ├── main.go           # Application entry point
│   └── router.go         # API route configuration
├── internal/
│   ├── api/              # HTTP handlers and middleware
│   │   ├── helpers.go    # api.Handle* functions (request/response)
│   │   └── base_controller.go  # Utilities for all controllers
│   ├── workspace/        # Workspace module
│   │   ├── controller/   # HTTP handlers
│   │   ├── service/      # Business logic
│   │   ├── repository/   # Data access
│   │   ├── model/        # Domain entities
│   │   ├── dto/          # Request/response DTOs
│   │   ├── messages/     # Constants for responses
│   │   ├── router/       # Route registration
│   │   └── app/          # Module initialization
│   ├── block/            # Block module (similar structure)
│   ├── page/             # Page module (similar structure)
│   ├── rbac/             # Role-Based Access Control
│   │   ├── service/      # Role & permission logic
│   │   ├── controller/   # RBAC HTTP handlers
│   │   └── middleware/   # Permission middleware
│   ├── auth/             # Authentication (OAuth, JWT, user sessions)
│   ├── user/             # User management
│   ├── errors/           # Custom error types and codes
│   ├── logger/           # Structured logging
│   ├── validation/       # Input validation utilities
│   ├── redis/            # Redis client (caching)
│   └── response/         # Response formatting
├── migrations/           # Database migrations
└── tests/
    ├── unit/             # Unit tests (services)
    ├── integration/      # End-to-end flow tests
    └── contract/         # API contract tests
```

### Frontend Structure

```
frontend/
├── src/
│   ├── app/
│   │   ├── layout.tsx          # Root layout
│   │   ├── page.tsx            # Home page
│   │   ├── middleware.ts       # Auth middleware
│   │   ├── (auth)/             # Auth routes
│   │   ├── api/                # API routes (Next.js API)
│   │   └── workspace/          # Workspace routes
│   ├── components/
│   │   ├── WorkspaceSidebar.tsx    # Navigation sidebar
│   │   ├── WorkspacePageShell.tsx  # Page container
│   │   ├── BlockList.tsx           # Block list renderer
│   │   ├── BlockEditor.tsx         # Block editing
│   │   ├── ui/                     # Reusable UI components
│   │   └── ShareDialog.tsx         # Sharing dialog
│   ├── features/
│   │   └── auth/                   # Auth logic
│   ├── pages/
│   │   └── WorkspacePageView.tsx   # Page view
│   ├── shared/
│   │   ├── constant/               # Constants
│   │   └── PermissionGuard.tsx     # Permission checks
│   └── services/
│       ├── workspaceApi.ts         # Workspace API client
│       ├── blockApi.ts             # Block API client
│       └── sharingApi.ts           # Sharing API client
└── tests/
    ├── unit/
    ├── integration/
    └── contract/
```

---

## API Endpoints

### Workspace API
```
GET    /api/notion/workspaces              # List all workspaces
POST   /api/notion/workspaces              # Create workspace
GET    /api/notion/workspaces/:id          # Get workspace
PUT    /api/notion/workspaces/:id          # Update workspace
DELETE /api/notion/workspaces/:id          # Delete workspace
```

### Page API
```
GET    /api/notion/pages                   # List pages (filter by workspaceId)
POST   /api/notion/pages                   # Create page
GET    /api/notion/pages/:id               # Get page
PUT    /api/notion/pages/:id               # Update page
DELETE /api/notion/pages/:id               # Delete page
GET    /api/notion/pages/hierarchy         # Get page hierarchy
```

### Block API
```
GET    /api/notion/blocks                  # List blocks (filter by pageId)
POST   /api/notion/blocks                  # Create block
GET    /api/notion/blocks/:id              # Get block
PUT    /api/notion/blocks/:id              # Update block
DELETE /api/notion/blocks/:id              # Delete block
POST   /api/notion/blocks/reorder          # Reorder blocks
```

### Sharing API
```
POST   /api/notion/sharing/grant           # Grant permission
DELETE /api/notion/sharing/revoke/:id      # Revoke permission
GET    /api/notion/sharing/collaborators   # List collaborators
```

---

## Data Model

### Core Entities

**Workspace**
- ID (UUID)
- Name (string)
- Description (string, nullable)
- CreatedAt, UpdatedAt
- Pages (relation: Workspace -> [Page])

**Page**
- ID (UUID)
- WorkspaceID (UUID)
- ParentPageID (UUID, nullable) - for hierarchy
- Title (string)
- ViewConfig (JSON) - for customization
- CreatedAt, UpdatedAt
- Blocks (relation: Page -> [Block])

**Block**
- ID (UUID)
- PageID (UUID)
- TypeID (UUID) - references BlockType
- Content (string, nullable)
- Rank (int64) - ordering
- Metadata (JSON) - block-specific data
- CreatedAt, UpdatedAt
- BlockType (relation: Block -> BlockType)

**BlockType**
- ID (UUID)
- Name (string) - "text", "heading", "checklist", etc.
- Schema (JSON) - metadata structure for this type

**SharePermission**
- ID (UUID)
- PageID (UUID)
- GranteeID (UUID) - user receiving access
- Role (string) - "viewer" or "editor"
- CreatedAt, UpdatedAt

**User**
- ID (UUID)
- Email (string, unique)
- Name (string)
- CreatedAt, UpdatedAt

---

## Key Design Decisions

### 1. Interface-Based Controllers
All controllers are interfaces with a single implementation struct. This enables:
- Easy mocking for tests
- Clear contracts for each endpoint
- Dependency injection of logger and services

### 2. Structured Error Handling
Custom error types (ValidationError, NotFoundError, InternalError) with:
- Error codes for categorization
- HTTP status mapping
- Structured logging

### 3. Type-Safe DTOs
Request/response objects defined in DTO packages:
- Separate from domain models
- Validation tags for input safety
- Decouples API from internal representation

### 4. Message Constants
User-facing messages in `messages.go` per module:
- Enables localization
- Reduces string duplication
- Centralized message management

### 5. Dependency Injection
Controllers and services use constructor injection:
- Logger injected at initialization
- Services injected into controllers
- Testable without global state

### 6. Last-Write-Wins Concurrency
Block edits use last-write-wins at block level:
- Simpler than operational transformation
- User sees updated_at timestamp
- Clear conflict resolution UI

### 7. Hierarchical Pages
Pages support nesting via ParentPageID:
- Enables flexible organization
- Sidebar shows tree structure
- Query returns full hierarchy

---

## Error Handling

### Error Categories

**Validation Errors (400 Bad Request)**
- Invalid UUID format
- Missing required fields
- Invalid input values
```go
errors.NewValidationError(
    errors.CodeInvalidID,
    "Invalid ID format",
    err,
)
```

**Not Found Errors (404 Not Found)**
- Resource doesn't exist
- User lacks access
```go
errors.NewNotFoundError(
    errors.CodeWorkspaceNotFound,
    "Workspace not found",
    err,
)
```

**Internal Errors (500 Server Error)**
- Database failures
- Service failures
- Unexpected conditions
```go
errors.NewInternalError(
    errors.CodeFailedToCreateWorkspace,
    "Failed to create workspace",
    err,
)
```

### Error Codes

See `backend/internal/errors/codes.go` for complete list:
- `INVALID_ID` - Format validation
- `WORKSPACE_NOT_FOUND` - Workspace missing
- `FAILED_TO_CREATE_WORKSPACE` - Service error
- etc.

---

## Testing Strategy

### Unit Tests (Services)
Focus on business logic:
- Input validation
- Business rules
- Error cases
```bash
make test-unit  # Run unit tests
```

### Integration Tests (End-to-End)
Test complete flows:
- Create workspace → create page → create block
- Permission enforcement
- Last-write-wins behavior
```bash
make test-integration
```

### Contract Tests (API)
Verify API contracts:
- Request/response schemas
- Status codes
- Error formats
```bash
make test-contract
```

---

## Performance Considerations

### Database Optimization
- Indexed queries on frequently filtered columns
- Pagination for large result sets
- N+1 query prevention via eager loading

### Frontend Optimization
- Virtual scrolling for large block lists
- Debounced save operations
- Optimistic UI updates

### Caching
- Redis for session data
- In-memory cache for block types
- Future: page content caching

---

## Security

### Authentication
- OAuth2 for login (Google, GitHub)
- JWT tokens for API access
- Secure session management

### Authorization
- Role-based access control (RBAC)
- Permission checking middleware
- Field-level access control

### Data Protection
- Password hashing (bcrypt)
- CORS configuration
- SQL injection prevention (parameterized queries)
- XSS prevention (CSP headers)

---

## Observability

### Logging
- Structured logging with Zap
- Log levels: Debug, Info, Warn, Error
- Error context with user/workspace IDs

### Metrics (Phase N)
- Page load latency
- API error rates
- Permission check failures
- Database query performance

### Debugging
- Request IDs for tracing
- Error stack traces in logs
- Query logging in development

---

## Running the Full Stack

### Docker Compose (Recommended)
```bash
docker-compose up -d
```

This starts:
- PostgreSQL database
- Redis cache
- Backend API (port 8080)
- Frontend app (port 3000)

### Manual Setup
See Backend and Frontend setup sections above.

---

## Troubleshooting

### Backend
- **Port 8080 in use**: `lsof -i :8080` and kill process
- **Database connection failed**: Check DATABASE_URL and postgres running
- **Migration issues**: `make migrate-down && make migrate-up`

### Frontend
- **Port 3000 in use**: `lsof -i :3000` and kill process
- **CORS errors**: Check API_URL env var matches backend
- **Blank page**: Check browser console for JS errors

### Tests
- **Tests fail**: Ensure database is clean - `make migrate-down && make migrate-up`
- **Flaky tests**: May need retry logic for async operations
- **Timeout issues**: Increase test timeout in config

---

## Contributing

### Code Standards
- Format: `make fmt`
- Lint: `make lint`
- Test: `make test`
- All must pass before PR merge

### Branch Naming
- Feature: `feature/description`
- Bugfix: `fix/description`
- Docs: `docs/description`

### Commit Messages
```
[TYPE] Brief description

Longer explanation if needed.

Fixes #123
```

Types: feat, fix, docs, style, refactor, test, chore

---

## Roadmap

### Current (MVP - Complete)
- ✅ User authentication
- ✅ Workspace CRUD
- ✅ Page hierarchy
- ✅ Block editing
- ✅ Basic sharing

### Next Phase (Enhancements)
- [ ] Real-time collaboration (WebSocket)
- [ ] Rich text with formatting
- [ ] Comments and mentions
- [ ] Version history
- [ ] Templates

### Future
- [ ] Mobile apps
- [ ] Offline sync
- [ ] Advanced search
- [ ] Analytics
- [ ] Multi-workspace permissions

---

## Support & Documentation

- **User Guide**: [quickstart.md](./quickstart.md)
- **Specification**: [spec.md](./spec.md)
- **API Docs**: Generated via Swagger in `backend/docs/`
- **Data Model**: [data-model.md](./data-model.md)
- **Architecture**: [plan.md](./plan.md)

