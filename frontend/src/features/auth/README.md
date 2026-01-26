# Authentication Feature

This module provides complete authentication functionality including login, signup, and OAuth integration (GitHub, Google).

## Features

- ✅ Email/Password authentication
- ✅ OAuth authentication (GitHub, Google)
- ✅ JWT token management
- ✅ Token refresh
- ✅ Protected routes
- ✅ Persistent authentication state (Zustand + localStorage)

## File Structure

```
src/features/auth/
├── api/
│   └── authApi.ts          # API service for auth endpoints
├── store/
│   └── authStore.ts        # Zustand store for auth state
└── README.md

src/app/(auth)/
├── login/
│   └── page.tsx            # Login page
├── signup/
│   └── page.tsx            # Signup page
└── oauth/
    └── callback/
        └── [provider]/
            └── page.tsx    # OAuth callback handler
```

## Usage

### Login

Navigate to `/login` or use the login page component:

```tsx
import { useAuthStore } from '@/features/auth/store/authStore'
import { authApi } from '@/features/auth/api/authApi'

const { setAuth } = useAuthStore()

const result = await authApi.login({ email, password })
if (!isLeft(result)) {
  setAuth(result.right.data)
}
```

### Signup

Navigate to `/signup` or use the signup page component:

```tsx
const result = await authApi.signUp({ name, email, password })
if (!isLeft(result)) {
  setAuth(result.right.data)
}
```

### OAuth Login

```tsx
// Get OAuth URL
const result = await authApi.getOAuthURL('github')
if (!isLeft(result)) {
  window.location.href = result.right.data.url
}

// After redirect, tokens are automatically retrieved in callback page
```

### Using Auth State

```tsx
import { useAuth } from '@/shared/contexts/AuthProvider'

function MyComponent() {
  const { user, isAuthenticated, handleLogout } = useAuth()
  
  if (!isAuthenticated) {
    return <div>Please login</div>
  }
  
  return (
    <div>
      <p>Welcome, {user?.name}!</p>
      <button onClick={handleLogout}>Logout</button>
    </div>
  )
}
```

### Using Auth Store Directly

```tsx
import { useAuthStore } from '@/features/auth/store/authStore'

function MyComponent() {
  const { user, accessToken, isAuthenticated, clearAuth } = useAuthStore()
  
  // Access user info
  console.log(user?.email)
  
  // Clear auth (logout)
  clearAuth()
}
```

## API Endpoints

All endpoints are prefixed with `/api/v1/auth`:

- `POST /signup` - Create new account
- `POST /login` - Login with email/password
- `POST /logout` - Logout
- `POST /refresh` - Refresh access token
- `GET /oauth/:provider/url` - Get OAuth authorization URL
- `GET /oauth/:provider/callback` - OAuth callback (handled by backend)
- `GET /oauth/tokens/:token_id` - Get tokens by token ID (OAuth flow)

## Environment Variables

Make sure to set these in your `.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8000
```

## OAuth Flow

1. User clicks "Login with GitHub/Google"
2. Frontend calls `GET /oauth/:provider/url`
3. User is redirected to OAuth provider
4. OAuth provider redirects to backend: `GET /oauth/:provider/callback?code=xxx`
5. Backend exchanges code, stores tokens in Redis, redirects to frontend: `/auth/oauth/callback/:provider?token_id=xxx`
6. Frontend callback page calls `GET /oauth/tokens/:token_id`
7. Tokens are stored in Zustand store and cookies
8. User is authenticated ✅

## Token Management

- **Access Token**: Stored in Zustand store and cookie, expires in 24 hours
- **Refresh Token**: Stored in Zustand store and cookie, expires in 7 days
- **Auto Refresh**: httpClient automatically refreshes token on 401 errors
- **Logout**: Clears tokens from store and cookies

## Protected Routes

To protect a route, check authentication:

```tsx
'use client'
import { useAuth } from '@/shared/contexts/AuthProvider'
import { useRouter } from 'next/navigation'
import { useEffect } from 'react'

export default function ProtectedPage() {
  const { isAuthenticated } = useAuth()
  const router = useRouter()
  
  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login')
    }
  }, [isAuthenticated, router])
  
  if (!isAuthenticated) {
    return null
  }
  
  return <div>Protected content</div>
}
```

