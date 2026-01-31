import axios, { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import ENV_CONFIG from '@/config/env'
import { ResponseResult } from '../types/common'
import { makeLeft, makeRight } from '../utils/either'
import ErrorException, { DataResponse } from './errorException'
import { clearAuthCookieClient, getAuthTokenClient } from './authCookieClient'
import { AnyObject } from '../types'

type AuthResponse = {
  access_token: string
  refresh_token: string
  user: {
    id: string
    email: string
    name: string
  }
}

const MAX_RETRIES = 2
let refreshTokenRequest: Promise<AuthResponse | null> | null = null

const httpClient: AxiosInstance = axios.create({
  baseURL: ENV_CONFIG.API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
})

httpClient.interceptors.request.use(
  config => {
    // Add access token to request header if available
    if (typeof window !== 'undefined') {
      const token = getAuthTokenClient()
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
    }
    return config
  },
  error => ({
    error,
  })
)

httpClient.interceptors.response.use(
  (response: AxiosResponse) => {
    return {
      ...response,
    }
  },
  async (error: AxiosError) => {
    const statusCode = error?.status
    const config = error.config as AxiosRequestConfig & { retryCount?: number }
    if (!config || !config.retryCount) {
      config.retryCount = 0
    }
    if (statusCode === 401 && config.retryCount < MAX_RETRIES && typeof window !== 'undefined') {
      config.retryCount += 1
      const newAuth = await refresher()
      if (newAuth) {
        refreshTokenRequest = null
        // Update auth store
        const { useAuthStore } = await import('@/features/auth/store/authStore')
        useAuthStore.getState().setAuth(newAuth)
        // Update cookie
        document.cookie = `accessToken=${newAuth.access_token}; path=/; max-age=${24 * 60 * 60}`
      }
      return httpClient(config)
    }
    if (config.retryCount === MAX_RETRIES && typeof window !== 'undefined') {
      clearAuthCookieClient()
      const { useAuthStore } = await import('@/features/auth/store/authStore')
      useAuthStore.getState().clearAuth()
      window.location.href = '/login'
    }
    return error
  }
)

async function refresher() {
  refreshTokenRequest = refreshTokenRequest ? refreshTokenRequest : handleRefreshToken()
  const response = await refreshTokenRequest
  if (response) {
    return response
  }
  return null
}

const handleRefreshToken = async (): Promise<AuthResponse | null> => {
  if (typeof window === 'undefined') return null

  const { useAuthStore } = await import('@/features/auth/store/authStore')
  const refreshToken = useAuthStore.getState().refreshToken

  if (!refreshToken) {
    clearAuthCookieClient()
    useAuthStore.getState().clearAuth()
    window.location.href = '/login'
    return null
  }

  try {
    const res = await fetch(`${ENV_CONFIG.API_BASE_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (!res.ok) {
      clearAuthCookieClient()
      useAuthStore.getState().clearAuth()
      window.location.href = '/login'
      return null
    }

    const data = await res.json()
    return data.data as AuthResponse
  } catch (error) {
    console.error('Refresh token error:', error)
    clearAuthCookieClient()
    useAuthStore.getState().clearAuth()
    window.location.href = '/login'
    return null
  }
}

async function httpClientCall<T>(fn: (httpClient: AxiosInstance) => Promise<AxiosResponse<T>>): Promise<ResponseResult<T>> {
  try {
    const response = await fn(httpClient)
    if (axios.isAxiosError(response)) {
      return makeLeft(ErrorException.fromError(response.response?.data))
    }
    return makeRight(DataResponse.fromResponse<T>(response.data as AnyObject))
  } catch (err: unknown) {
    return makeLeft(new ErrorException(err as ErrorException))
  }
}

export default httpClientCall

