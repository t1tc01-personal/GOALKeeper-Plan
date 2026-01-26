import httpClientCall from '@/shared/lib/httpClient'
import { ResponseResult } from '@/shared/types/common'
import {
  AuthResponse,
  SignUpRequest,
  LoginRequest,
  RefreshTokenRequest,
  LogoutRequest,
  OAuthTokenData,
} from '@/shared/types/auth'
import ENV_CONFIG from '@/config/env'

const API_BASE = `${ENV_CONFIG.API_BASE_URL}/api/v1/auth`

export const authApi = {
  signUp: (data: SignUpRequest): ResponseResult<AuthResponse> => {
    return httpClientCall(client =>
      client.post(`${API_BASE}/signup`, data)
    )
  },

  login: (data: LoginRequest): ResponseResult<AuthResponse> => {
    return httpClientCall(client =>
      client.post(`${API_BASE}/login`, data)
    )
  },

  logout: (data: LogoutRequest): ResponseResult<void> => {
    return httpClientCall(client =>
      client.post(`${API_BASE}/logout`, data)
    )
  },

  refreshToken: (data: RefreshTokenRequest): ResponseResult<AuthResponse> => {
    return httpClientCall(client =>
      client.post(`${API_BASE}/refresh`, data)
    )
  },

  getOAuthTokens: (tokenId: string): ResponseResult<OAuthTokenData> => {
    return httpClientCall(client =>
      client.get(`${API_BASE}/oauth/tokens/${tokenId}`)
    )
  },

  oauthTokenCallback: (provider: 'github' | 'google', accessToken: string): ResponseResult<AuthResponse> => {
    return httpClientCall(client =>
      client.post(`${API_BASE}/oauth/${provider}/token-callback`, {
        access_token: accessToken,
      })
    )
  },
}

