export interface User {
  id: string
  email: string
  name: string
  is_email_verified?: boolean
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  user: User
}

export interface SignUpRequest {
  email: string
  password: string
  name: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RefreshTokenRequest {
  refresh_token: string
}

export interface LogoutRequest {
  refresh_token: string
}

export interface OAuthTokenData {
  access_token: string
  refresh_token: string
  user_id: string
  email: string
  name: string
}

