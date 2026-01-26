# Swagger API Documentation

This directory contains auto-generated Swagger documentation for the GOALKeeper-Plan API.

## Generating Swagger Docs

After adding or modifying API endpoints with Swagger annotations, regenerate the docs:

```bash
make swagger
```

Or manually:

```bash
swag init -g cmd/main.go -o docs
```

## Viewing Swagger UI

1. Start the server:
```bash
make run
# or
go run ./cmd/main.go api
```

2. Open your browser and navigate to:
```
http://localhost:8000/swagger/index.html
```

## Adding Swagger Annotations

### Basic Endpoint Annotation

```go
// EndpointName godoc
// @Summary      Short summary
// @Description  Detailed description
// @Tags         tag-name
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "Parameter description"
// @Param        request   body      dto.Request  true  "Request body"
// @Success      200       {object}  response.Response{data=dto.Response}
// @Failure      400       {object}  response.Response
// @Router       /api/v1/endpoint/{id} [post]
func (c *controller) EndpointName(ctx *gin.Context) {
    // ...
}
```

### Authentication

For endpoints that require authentication, add:

```go
// @Security     BearerAuth
```

### Response Types

- `{object}` - For single object response
- `{array}` - For array response
- `response.Response{data=Type}` - For wrapped response with data field

## Swagger Annotations Reference

- `@Summary` - Short summary of the endpoint
- `@Description` - Detailed description
- `@Tags` - Group endpoints (e.g., "auth", "users")
- `@Accept` - Content types accepted (json, xml, etc.)
- `@Produce` - Content types produced
- `@Param` - Request parameters (path, query, body, header)
- `@Success` - Success response with status code and schema
- `@Failure` - Error response with status code
- `@Router` - Route path and HTTP method
- `@Security` - Security scheme (BearerAuth)

## Example

See `internal/auth/controller/auth_controller.go` and `internal/user/controller/user_controller.go` for examples of Swagger annotations.

