# Setup Guide: GOALKeeper Development Environment

**Last Updated**: January 27, 2026
**For**: Developers setting up local development
**Time Required**: ~15-20 minutes

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Environment Setup](#environment-setup)
3. [Backend Setup](#backend-setup)
4. [Frontend Setup](#frontend-setup)
5. [Database Setup](#database-setup)
6. [Running the Application](#running-the-application)
7. [Verification](#verification)
8. [Troubleshooting](#troubleshooting)
9. [Development Workflow](#development-workflow)

---

## Prerequisites

### Required Software

- **Go**: 1.21 or higher ([download](https://go.dev/dl))
- **Node.js**: 18.x or higher ([download](https://nodejs.org))
- **PostgreSQL**: 13 or higher ([download](https://www.postgresql.org/download))
- **Docker** (optional, recommended): [download](https://docs.docker.com/get-docker/)
- **Git**: For cloning the repository

### Verification

Check installed versions:

```bash
go version          # Should output Go 1.21+
node --version      # Should output Node 18+
npm --version       # Should output npm 9+
psql --version      # Should output PostgreSQL 13+
```

---

## Environment Setup

### 1. Clone Repository

```bash
git clone https://github.com/your-org/GOALKeeper-Plan.git
cd GOALKeeper-Plan
```

### 2. Set Environment Variables

Create a `.env` file in the project root:

```bash
# Database
DATABASE_URL=postgres://postgres:password@localhost:5432/goalkeeper_dev
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=goalkeeper_dev
DB_SSLMODE=disable

# Backend
BACKEND_PORT=8080
LOG_LEVEL=debug

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080
NODE_ENV=development

# Testing
TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/goalkeeper_test
```

**Security Note**: Never commit `.env` files. They're in `.gitignore`.

---

## Backend Setup

### 1. Install Go Dependencies

```bash
cd backend
go mod download
go mod tidy
```

### 2. Install Development Tools

```bash
# Swagger documentation generation
go install github.com/swaggo/swag/cmd/swag@latest

# Hot reload for development (optional)
go install github.com/cosmtrek/air@latest
```

Add `$GOPATH/bin` to your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### 3. Verify Backend Compilation

```bash
go build -o bin/main ./cmd/main.go
# Should create ./bin/main executable
```

---

## Frontend Setup

### 1. Install Node Dependencies

```bash
cd frontend
npm install
# or yarn install / pnpm install
```

### 2. Generate Types from API (Optional)

If using OpenAPI generator:

```bash
npm run generate-types
```

### 3. Verify Frontend Build

```bash
npm run build
# Should complete without errors
```

---

## Database Setup

### Option A: Docker Compose (Recommended)

```bash
# From project root
docker-compose up -d postgres

# Wait ~10 seconds for PostgreSQL to be ready
sleep 10

# Verify connection
psql $DATABASE_URL -c "SELECT 1;"
```

### Option B: Local PostgreSQL

#### macOS (Homebrew)
```bash
brew install postgresql@15
brew services start postgresql@15

# Create database and user
createuser -P postgres  # Set password
createdb -U postgres goalkeeper_dev
```

#### Ubuntu/Debian
```bash
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql

# Create database and user
sudo -u postgres createuser -P postgres
sudo -u postgres createdb -O postgres goalkeeper_dev
```

#### Windows
- Download from postgresql.org and run installer
- Create database using pgAdmin

### Verify Database Connection

```bash
psql -h localhost -U postgres -d goalkeeper_dev -c "SELECT NOW();"
# Should return current timestamp
```

---

## Running Migrations

### 1. Install Atlas (Schema Migration Tool)

```bash
curl -sSf https://atlasgo.sh | sh
```

### 2. Run Migrations

```bash
cd backend

# Apply all pending migrations
atlas migrate apply \
  --dir file://migrations \
  --url $DATABASE_URL

# Or using Go directly
go run ./migrations/main.go up
```

### Verify Schema

```bash
psql -h localhost -U postgres -d goalkeeper_dev

# In psql:
\dt                          # List tables
SELECT * FROM workspaces;    # Should return empty table
\q                           # Quit
```

---

## Running the Application

### Terminal 1: Backend

```bash
cd backend

# Using Go directly (slow)
go run ./cmd/main.go

# Or using built binary (fast)
go build -o bin/main ./cmd/main.go
./bin/main

# With hot reload (requires air)
air
```

Expected output:
```
[GIN-debug] Loaded HTML Templates (2): 
[GIN-debug] Listening and serving HTTP on :8080
```

### Terminal 2: Frontend

```bash
cd frontend

# Development mode (with hot reload)
npm run dev

# Then open: http://localhost:3000
```

Expected output:
```
  ▲ Next.js 14.x
  - Local:        http://localhost:3000
  - Environments: .env.local

 ✓ Ready in 1234ms
```

### Browser Access

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **API Docs (Swagger)**: http://localhost:8080/swagger/index.html

---

## Verification

### 1. Test API Connection

```bash
# Create workspace
curl -X POST http://localhost:8080/api/v1/notion/workspaces \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{"name": "Test Workspace"}'

# Should return 201 with workspace data
```

### 2. Test Frontend

```bash
# Open browser and navigate to http://localhost:3000
# Should display home page without errors
# Check browser console for any errors (F12)
```

### 3. Run Tests

```bash
# Backend unit tests
cd backend
go test ./...

# Frontend unit tests
cd frontend
npm run test

# Should show green tests passing
```

---

## Development Workflow

### Adding a New Feature

1. Create feature branch: `git checkout -b feature/my-feature`
2. Make changes to backend and/or frontend
3. Test locally
4. Run tests: `npm test` (frontend), `go test ./...` (backend)
5. Commit and push: `git push origin feature/my-feature`
6. Create pull request

### Typical Development Commands

```bash
# Backend
cd backend
go mod tidy          # Update dependencies
go fmt ./...         # Format code
go vet ./...         # Static analysis
go test ./...        # Run tests
go build ./cmd/main.go  # Build

# Frontend
cd frontend
npm run dev          # Start dev server
npm run build        # Build for production
npm run test         # Run tests
npm run lint         # Run linter
```

### Code Style

**Backend**: Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
```bash
go fmt ./...  # Auto-format code
```

**Frontend**: Follow configured ESLint rules
```bash
npm run lint  # Check linting
npm run lint:fix  # Auto-fix issues
```

---

## Troubleshooting

### Database Connection Issues

**Error**: `connection refused` on localhost:5432

**Solution**:
1. Verify PostgreSQL is running: `psql --version`
2. Start PostgreSQL:
   - macOS: `brew services start postgresql@15`
   - Linux: `sudo systemctl start postgresql`
   - Docker: `docker-compose up -d postgres`
3. Verify connection: `psql -h localhost -U postgres -d goalkeeper_dev`

### Port Already in Use

**Error**: `listen tcp :8080: bind: address already in use`

**Solution**:
```bash
# Kill process on port 8080
lsof -i :8080          # Find process ID (macOS/Linux)
kill -9 <PID>

# Or use different port
BACKEND_PORT=8081 go run ./cmd/main.go
```

### Node Dependencies Issues

**Error**: `npm ERR! code ERESOLVE`

**Solution**:
```bash
npm install --legacy-peer-deps
# or
npm install --force
```

### Go Module Issues

**Error**: `go: no such file or directory`

**Solution**:
```bash
cd backend
go mod download
go mod verify
go mod tidy
```

### Database Migration Fails

**Error**: `migration: current version is invalid`

**Solution**:
```bash
# Reset database (⚠️ deletes all data)
dropdb -U postgres goalkeeper_dev
createdb -U postgres goalkeeper_dev

# Re-apply migrations
cd backend
atlas migrate apply --dir file://migrations --url $DATABASE_URL
```

### Tests Fail Locally

**Common Issues**:
1. Database not migrated: Run migrations again
2. Environment variables not set: Check `.env` file
3. Port already in use: Kill existing process

```bash
# Full reset
docker-compose down
rm -rf backend/bin frontend/node_modules
docker-compose up -d postgres
# Wait 10 seconds
cd backend && go build ./cmd/main.go
cd frontend && npm install
```

---

## IDE Setup

### VS Code (Recommended)

**Extensions**:
- Go (golang.go)
- ESLint (dbaeumer.vscode-eslint)
- Prettier (esbenp.prettier-vscode)
- PostgreSQL (ckolkman.vscode-postgres)
- Thunder Client or REST Client for API testing

**Debugging Backend**:
1. Set breakpoint in code
2. Run → Start Debugging (F5)
3. Select "Go" environment

**Debugging Frontend**:
1. Run: `npm run dev`
2. Open Chrome DevTools (F12)
3. Set breakpoints in Sources tab

### GoLand / IntelliJ IDEA

- Open `backend` folder as project
- Configure Go SDK: Preferences → Go → Go Modules
- Run → Edit Configurations → Add Go Application

### WebStorm

- Open `frontend` folder as project
- Configure Node interpreter: Preferences → Node.js and NPM
- Run → Edit Configurations → Add npm script

---

## Next Steps

1. ✅ Database set up locally
2. ✅ Backend running on port 8080
3. ✅ Frontend running on port 3000
4. ✅ API accessible and tested
5. Read [API_REFERENCE.md](./API_REFERENCE.md) for endpoint documentation
6. Read [ARCHITECTURE.md](./ARCHITECTURE.md) for codebase overview
7. Start exploring: Create a workspace and add pages/blocks via the UI

---

## Production Deployment

For deployment instructions, see [DEPLOYMENT.md](./DEPLOYMENT.md).

---

## Getting Help

- Check [troubleshooting](#troubleshooting) section above
- Review error messages carefully
- Search GitHub issues: https://github.com/your-org/GOALKeeper-Plan/issues
- Ask in project Discord/Slack channel

