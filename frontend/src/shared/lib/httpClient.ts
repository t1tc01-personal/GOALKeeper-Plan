import axios, { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import ENV_CONFIG from '@/config/env'
import { ResponseResult } from '../types/common'
import { makeLeft, makeRight } from '../utils/either'
import ErrorException, { DataResponse } from './errorException'
import { clearAuthCookie, setAccessTokenCookie } from './authCookie'
import { AnyObject } from '../types'

type TokenResponse = {
  accessToken: string
  expiresIn: number
}

const MAX_RETRIES = 2
let refreshTokenRequest: Promise<TokenResponse> | null = null

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
    if (statusCode === 401 && config.retryCount < MAX_RETRIES) {
      config.retryCount += 1
      const newToken = await refresher()
       if (newToken) {
        refreshTokenRequest = null
        await setAccessTokenCookie(newToken.accessToken, newToken.expiresIn)
       }
      return httpClient(config)
    }
    if (config.retryCount === MAX_RETRIES) {
      await clearAuthCookie()
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
}

const handleRefreshToken = async (): Promise<TokenResponse> => {
  const res = await fetch(`${ENV_CONFIG.API_BASE_URL}/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
  })
  if (!res.ok) {
    await clearAuthCookie()
    window.location.href = '/login'
  }
  return res.json()
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

