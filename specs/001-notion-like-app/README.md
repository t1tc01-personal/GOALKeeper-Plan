# GOALKeeper: Notion-like Workspace Application

**Status**: MVP Complete ✅
**Version**: 1.0
**Last Updated**: January 27, 2026

---

## Quick Start (5 minutes)

```bash
# Clone repository
git clone https://github.com/your-org/GOALKeeper-Plan.git
cd GOALKeeper-Plan

# Start database
docker-compose up -d postgres

# Start backend (terminal 1)
cd backend && go run ./cmd/main.go

# Start frontend (terminal 2)
cd frontend && npm install && npm run dev

# Open browser
# Frontend: http://localhost:3000
# API Docs: http://localhost:8080/swagger/index.html
```

See [SETUP_GUIDE.md](./SETUP_GUIDE.md) for detailed setup instructions.

---

## What is GOALKeeper?

GOALKeeper is a **web-based workspace application** that allows users to:

✅ **Create and organize pages** in hierarchical structures
✅ **Add rich content blocks** (paragraphs, headings, checklists)
✅ **Collaborate with others** via shared access with different permission levels
✅ **Real-time collaboration** with last-write-wins conflict resolution

**Example Use Cases**:
- Personal knowledge base / note-taking app
- Team wiki and documentation
- Project planning and tracking
- Content management system

---

## Features

### Phase 1: Workspaces & Pages ✅
- Create and manage workspaces
- Organize pages in hierarchical structure
- Rename and reorder pages

### Phase 2: Content Editing ✅
- Create blocks (paragraphs, headings, checklists)
- Edit and delete blocks
- Reorder blocks within pages
- Persistent storage with timestamps

### Phase 3: Sharing & Permissions ✅
- Share pages with specific users
- Three permission levels: **Viewer**, **Editor**, **Owner**
- Role-based access control (RBAC)
- Revoke access at any time

---

## Architecture Overview

### Stack

- **Backend**: Go 1.21, Gin framework, GORM ORM
- **Frontend**: React 18, TypeScript, Next.js, Vitest
- **Database**: PostgreSQL 15
- **API**: REST with JSON
- **Deployment**: Docker, cloud-ready

### Design Pattern

```
Frontend (React/Next.js)
    ↓ JSON over HTTP/REST
Backend API (Go/Gin)
    ├── Controller Layer (HTTP handlers)
    ├── Service Layer (business logic)
    └── Repository Layer (database queries)
        ↓ SQL
Database (PostgreSQL)
```

### Key Architecture Principles

1. **Layered Architecture**: Clean separation of concerns (Repository → Service → Controller)
2. **Dependency Injection**: Testable and maintainable code
3. **Last-Write-Wins**: Simple conflict resolution for concurrent edits
4. **RBAC**: Role-based permissions at page level
5. **REST API**: Standard HTTP verbs for CRUD operations

---

## Documentation

### For Developers

- **[SETUP_GUIDE.md](./SETUP_GUIDE.md)** - Get your development environment running (15 minutes)
- **[API_REFERENCE.md](./API_REFERENCE.md)** - Complete API documentation with examples
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Deep dive into codebase structure and patterns

### For DevOps/Infrastructure

- **[DEPLOYMENT.md](./DEPLOYMENT.md)** - Production deployment on AWS, DigitalOcean, Vercel
- **[TESTING.md](./TESTING.md)** - Test strategies and running tests locally
- **[docs/QUICKSTART_VALIDATION.md](../docs/QUICKSTART_VALIDATION.md)** - End-to-end validation guide

### Technical Specifications

- **[plan.md](./plan.md)** - Technical plan and architecture decisions
- **[data-model.md](./data-model.md)** - Entity relationships and database schema
- **[research.md](./research.md)** - Research decisions and constraints
- **[quickstart.md](./quickstart.md)** - Integration scenarios and examples

---

## API Quick Reference

### Authentication

All requests require `X-User-ID` header:

```bash
curl -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  http://localhost:8080/api/v1/notion/workspaces
```

### Core Endpoints

**Workspaces**:
- `POST /workspaces` - Create workspace
- `GET /workspaces` - List workspaces
- `GET /workspaces/:id` - Get workspace

**Pages**:
- `POST /pages` - Create page
- `GET /pages?workspace_id=...` - List pages
- `GET /pages/:id` - Get page with children
- `PUT /pages/:id` - Update page
- `DELETE /pages/:id` - Delete page

**Blocks**:
- `POST /pages/:page_id/blocks` - Create block
- `GET /pages/:page_id/blocks` - List blocks
- `PUT /pages/:page_id/blocks/:id` - Update block
- `DELETE /pages/:page_id/blocks/:id` - Delete block

**Permissions**:
- `POST /pages/:page_id/share` - Grant access
- `DELETE /pages/:page_id/share/:user_id` - Revoke access
- `GET /pages/:page_id/collaborators` - List collaborators

See [API_REFERENCE.md](./API_REFERENCE.md) for full documentation.

---

## Running Locally

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 13+ (or Docker)
- Git

### Quick Setup (Docker)

```bash
# Start database
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
sleep 10

# Backend (Terminal 1)
cd backend
go mod download
go run ./cmd/main.go

# Frontend (Terminal 2)
cd frontend
npm install
npm run dev

# Open http://localhost:3000
```

### Testing Locally

```bash
# Backend unit tests
cd backend
go test ./...

# Frontend unit tests
cd frontend
npm test

# E2E validation (after starting server)
cd specs/001-notion-like-app
# Follow QUICKSTART_VALIDATION.md
```

See [SETUP_GUIDE.md](./SETUP_GUIDE.md) for detailed instructions.

---

## Testing

### Automated Validation

Run the quickstart validation guide:

```bash
# Start backend and frontend first
# Then follow: docs/QUICKSTART_VALIDATION.md

# Covers:
# - Workspace & page management
# - Block editing and persistence
# - Sharing & permissions
# - Performance baseline
# - Automated checks (build, tests, type-check)
```

### Manual Testing

The application is accessible at:
- **Frontend**: http://localhost:3000
- **Swagger Docs**: http://localhost:8080/swagger/index.html

Test workflows:
1. Create a workspace
2. Create pages with hierarchy
3. Add and edit blocks
4. Share with another user
5. Verify permissions enforcement

---

## Deployment

### Quick Deploy to Production

**Option 1: Vercel + AWS (Recommended)**
```bash
vercel deploy --prod        # Frontend
# Backend: AWS EC2 or ECS
```

**Option 2: DigitalOcean App Platform**
```bash
doctl apps create --spec app.yaml
```

**Option 3: Docker Compose**
```bash
docker-compose -f docker-compose.prod.yaml up -d
```

See [DEPLOYMENT.md](./DEPLOYMENT.md) for detailed instructions, cost estimates, and monitoring setup.

---

## Project Structure

```
GOALKeeper-Plan/
├── README.md                          # This file
├── docker-compose.yaml                # Local development
│
├── backend/                           # Go application
│   ├── cmd/main.go                   # Entry point
│   ├── internal/
│   │   ├── workspace/                # Workspace module
│   │   ├── page/                     # Page module
│   │   ├── block/                    # Block module
│   │   ├── rbac/                     # Permission module
│   │   └── ...
│   ├── migrations/                   # Database migrations
│   └── tests/                        # Unit, integration, contract tests
│
├── frontend/                          # React application
│   ├── src/
│   │   ├── app/                      # Next.js app directory
│   │   ├── components/               # UI components
│   │   ├── features/                 # Feature modules
│   │   └── shared/                   # Shared utilities
│   └── tests/                        # Unit and integration tests
│
└── specs/001-notion-like-app/        # Specification & docs
    ├── README.md                     # Spec overview
    ├── plan.md                       # Technical plan
    ├── data-model.md                 # Database schema
    ├── API_REFERENCE.md              # API endpoints
    ├── ARCHITECTURE.md               # Code architecture
    ├── SETUP_GUIDE.md                # Development setup
    ├── DEPLOYMENT.md                 # Production deployment
    ├── tasks.md                      # Task breakdown
    └── checklists/                   # Quality checklists
```

---

## Development Workflow

### Local Development

```bash
# Create feature branch
git checkout -b feature/my-feature

# Make changes
# ... edit files ...

# Test locally
npm test              # frontend
go test ./...         # backend

# Commit and push
git push origin feature/my-feature

# Create pull request
```

### Code Standards

- **Go**: Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
  - Run `go fmt`, `go vet`, `go test`
- **TypeScript/React**: Follow ESLint configuration
  - Run `npm run lint`, `npm run build`, `npm test`

---

## Monitoring & Observability

### Local Development
- Logs: stdout (backend) and browser console (frontend)
- API Docs: http://localhost:8080/swagger/index.html

### Production (See DEPLOYMENT.md)
- Error tracking: Sentry
- Performance monitoring: Datadog / AWS X-Ray
- Logs: CloudWatch / ELK
- Metrics: Prometheus / AWS CloudWatch

---

## Known Limitations & Future Enhancements

### Current Limitations (MVP)
- Last-write-wins conflict resolution (no merging)
- User identification via header only (no proper auth)
- No real-time collaboration (no WebSocket)
- No version history / audit trail
- No commenting or mentions
- Limited block types (paragraph, heading, checklist)

### Planned Enhancements (Phase 4+)
- [ ] Real-time collaboration with WebSockets
- [ ] Operational Transform (OT) or CRDT for conflict resolution
- [ ] Version history and time travel
- [ ] Comments and mentions
- [ ] Rich text editor (Slate, TipTap)
- [ ] Image and file upload
- [ ] OAuth2 / OIDC authentication
- [ ] Team/organization management
- [ ] Mobile app (React Native)
- [ ] Dark mode

---

## Support

### Getting Help

1. Check documentation: [SETUP_GUIDE.md](./SETUP_GUIDE.md), [API_REFERENCE.md](./API_REFERENCE.md)
2. Search existing issues: https://github.com/your-org/GOALKeeper-Plan/issues
3. Create new issue with:
   - Description of problem
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment (OS, Go version, Node version, etc.)

### Reporting Bugs

```bash
# Include system info
go version
node --version
npm --version
psql --version

# Include error logs
# Copy relevant error messages from console/logs
```

---

## Contributing

1. Read [ARCHITECTURE.md](./ARCHITECTURE.md) to understand the codebase
2. Follow the development workflow above
3. Write tests for new features
4. Update documentation if adding new endpoints/APIs
5. Run local tests before committing

---

## License

This project is licensed under MIT License. See LICENSE file for details.

---

## Changelog

### Version 1.0 (2026-01-27)
✅ **User Story 1**: Workspace and page management
✅ **User Story 2**: Block editing with persistence
✅ **User Story 3**: Sharing and role-based permissions
✅ **Documentation**: Complete API, architecture, and deployment guides
✅ **Testing**: Unit, integration, and contract tests

See [tasks.md](./tasks.md) for detailed implementation timeline.

---

## Quick Links

| Link | Purpose |
|------|---------|
| [Setup Guide](./SETUP_GUIDE.md) | Get started in 15 minutes |
| [API Reference](./API_REFERENCE.md) | API endpoint documentation |
| [Architecture](./ARCHITECTURE.md) | Understanding the codebase |
| [Deployment](./DEPLOYMENT.md) | Production deployment |
| [Quickstart Validation](../docs/QUICKSTART_VALIDATION.md) | E2E validation guide |

---

## Contact & Support

- **Email**: team@goalkeeper.example.com
- **Slack**: #goalkeeper-dev
- **GitHub Issues**: https://github.com/your-org/GOALKeeper-Plan/issues

---

**Made with ❤️ by the GOALKeeper Team**

