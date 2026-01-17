# GOALKeeper-Plan Backend

Backend API server for GOALKeeper-Plan, built with Go 1.22+ using the standard library's `net/http` package and the new `ServeMux`.

## Project Structure

```
backend/
├── cmd/
│   └── api/              # Application entry point
│       ├── main.go       # CLI setup with cobra
│       ├── server.go     # Server initialization
│       └── router.go     # HTTP router configuration
├── config/               # Configuration management
│   ├── config.go         # Config structs and initialization
│   └── config.yaml       # Configuration file template
├── internal/             # Private application code
│   ├── logger/           # Logging utilities
│   └── user/             # User domain module
│       ├── app/          # Application initialization
│       ├── controller/   # HTTP handlers
│       ├── dto/          # Data transfer objects
│       ├── model/        # Domain models
│       ├── repository/   # Data access layer
│       ├── router/       # Route registration
│       └── service/      # Business logic
└── go.mod                # Go module definition
```

## Architecture

The backend follows a clean architecture pattern with clear separation of concerns:

- **cmd/api**: Entry point and server setup
- **config**: Configuration management using Viper
- **internal**: Private packages following domain-driven design
  - Each domain (user, goal, plan, etc.) has its own module with:
    - `app/`: Application initialization and dependency injection
    - `controller/`: HTTP request handlers
    - `dto/`: Request/response DTOs
    - `model/`: Domain models (GORM)
    - `repository/`: Data access layer
    - `router/`: Route registration
    - `service/`: Business logic

## Features

- ✅ Go 1.22+ standard library HTTP server with ServeMux
- ✅ Structured logging with zap
- ✅ Configuration management with Viper
- ✅ PostgreSQL database with GORM
- ✅ CLI interface with Cobra
- ✅ Graceful shutdown
- ✅ Health check endpoints
- ✅ CORS middleware
- ✅ Request logging middleware

## Setup

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Configure environment:**
   - Copy `config/config.yaml` and update with your settings
   - Create `.env` file with required environment variables:
     ```
     POSTGRES_CONNECTION_STRING=postgres://user:password@localhost:5432/goalkeeper_plan?sslmode=disable
     AUTH_SECRET=your-secret-key
     REDIS_HOST=localhost
     REDIS_PORT=6379
     REDIS_PASSWORD=
     ```

3. **Run database migrations:**
   ```bash
   # TODO: Add migration commands
   ```

4. **Run the server:**
   ```bash
   go run cmd/api/main.go api
   ```

## API Endpoints

### Health Checks
- `GET /ready/readiness` - Service readiness check
- `GET /ready/liveliness` - Service liveness check

### Users
- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users/{id}` - Get user by ID
- `GET /api/v1/users` - List users (with pagination)
- `PUT /api/v1/users/{id}` - Update user
- `PATCH /api/v1/users/{id}` - Partially update user
- `DELETE /api/v1/users/{id}` - Delete user

## Development

### Adding a New Domain Module

1. Create the module structure:
   ```
   internal/{domain}/
   ├── app/
   ├── controller/
   ├── dto/
   ├── model/
   ├── repository/
   ├── router/
   └── service/
   ```

2. Implement each layer following the user module as a reference.

3. Register the module in `cmd/api/router.go`:
   ```go
   {domain}.NewApplication(db, apiV1, configs, logger)
   ```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## License

[Your License Here]

