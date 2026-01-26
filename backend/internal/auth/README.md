# Authentication API Documentation

This module provides JWT-based authentication with support for password-based login and OAuth (GitHub, Google).

## API Endpoints

### Public Endpoints

#### 1. Sign Up
```http
POST /api/v1/auth/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

#### 2. Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:** Same as signup

#### 3. Logout
```http
POST /api/v1/auth/logout
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### 4. Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:** Same as signup (new tokens)

#### 5. Get OAuth URL
```http
GET /api/v1/auth/oauth/:provider/url
```

**Providers:** `github`, `google`

**Response:**
```json
{
  "url": "https://github.com/login/oauth/authorize?..."
}
```

#### 6. OAuth Callback (GET - Backend Redirect)
```http
GET /api/v1/auth/oauth/:provider/callback?code=xxx&state=xxx
```

**Flow:**
1. OAuth provider redirects to this endpoint with authorization code
2. Backend exchanges code for tokens
3. Tokens are stored in Redis with a temporary token ID
4. Backend redirects to frontend with token ID: `{FRONTEND_URL}/auth/oauth/callback/{provider}?token_id=xxx`

**Note:** This endpoint is called by OAuth providers, not directly by frontend.

#### 7. OAuth Callback (POST - Frontend API)
```http
POST /api/v1/auth/oauth/:provider/callback
Content-Type: application/json

{
  "code": "oauth_code_from_provider",
  "state": "optional_state"
}
```

**Response:** Same as signup

**Note:** Alternative flow where frontend receives code and calls this endpoint directly.

#### 8. Get OAuth Tokens by Token ID
```http
GET /api/v1/auth/oauth/tokens/:token_id
```

**Description:** Retrieve tokens using token ID from OAuth callback redirect. Tokens are deleted after retrieval (one-time use).

**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": "uuid",
    "email": "user@example.com",
    "name": "User Name"
  }
}
```

**Error Response (404):**
```json
{
  "success": false,
  "error": {
    "type": "not_found",
    "code": "TOKEN_INVALID",
    "message": "Token not found or expired"
  }
}
```

## Protected Routes

To protect a route, use the `AuthMiddleware`:

```go
import (
    "goalkeeper-plan/internal/auth/jwt"
    "goalkeeper-plan/internal/auth/middleware"
)

// In your router setup
jwtService := jwt.NewJWTService(...)
protected := router.Group("/api/v1/protected")
protected.Use(middleware.AuthMiddleware(jwtService, logger))
{
    protected.GET("/profile", getProfileHandler)
}
```

## Accessing User Info in Handlers

After authentication, user info is available in the context:

```go
func getProfileHandler(ctx *gin.Context) {
    userID := ctx.GetString("user_id")
    email := ctx.GetString("email")
    // Use userID and email
}
```

## OAuth Flow (Backend Redirect)

### Flow Diagram

```
1. Frontend → GET /api/v1/auth/oauth/github/url
   └─> Backend returns OAuth authorization URL

2. User → Redirect to GitHub/Google
   └─> User authorizes application

3. OAuth Provider → GET /api/v1/auth/oauth/github/callback?code=xxx
   └─> Backend receives authorization code

4. Backend:
   ├─> Exchanges code for user info
   ├─> Creates/updates user
   ├─> Generates JWT tokens
   ├─> Stores tokens in Redis with token_id
   └─> Redirects to frontend: {FRONTEND_URL}/auth/oauth/callback/github?token_id=xxx

5. Frontend → GET /api/v1/auth/oauth/tokens/{token_id}
   └─> Backend returns tokens (one-time use, deleted after retrieval)

6. Frontend stores tokens and user is authenticated
```

### Frontend Implementation

```typescript
// 1. Get OAuth URL
const response = await fetch('/api/v1/auth/oauth/github/url');
const { data } = await response.json();
window.location.href = data.url;

// 2. After redirect back from OAuth provider
// Frontend receives: /auth/oauth/callback/github?token_id=xxx
const urlParams = new URLSearchParams(window.location.search);
const tokenId = urlParams.get('token_id');

if (tokenId) {
  // 3. Get tokens
  const tokenResponse = await fetch(`/api/v1/auth/oauth/tokens/${tokenId}`);
  const { data: tokenData } = await tokenResponse.json();
  
  // 4. Store tokens
  localStorage.setItem('access_token', tokenData.access_token);
  localStorage.setItem('refresh_token', tokenData.refresh_token);
  
  // 5. Redirect to dashboard
  window.location.href = '/dashboard';
}
```

## Environment Variables

Add these to your `.env` file:

```env
AUTH_SECRET=your-secret-key-min-32-chars
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
FRONTEND_URL=http://localhost:3000
BACKEND_URL=http://localhost:8000
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

## OAuth Setup

### GitHub OAuth App
1. Go to GitHub Settings > Developer settings > OAuth Apps
2. Create a new OAuth App
3. Set Authorization callback URL to: `{BACKEND_URL}/api/v1/auth/oauth/github/callback`
   - Example: `http://localhost:8000/api/v1/auth/oauth/github/callback`
4. Copy Client ID and Client Secret

### Google OAuth App
1. Go to Google Cloud Console > APIs & Services > Credentials
2. Create OAuth 2.0 Client ID
3. Set Authorized redirect URIs to: `{BACKEND_URL}/api/v1/auth/oauth/google/callback`
   - Example: `http://localhost:8000/api/v1/auth/oauth/google/callback`
4. Copy Client ID and Client Secret

## Token Expiration

- Access Token: 24 hours (configurable via `auth.jwtExpiration` in config.yaml)
- Refresh Token: 7 days (configurable via `auth.refreshExpiration` in config.yaml)

