'use client'

export function getAuthTokenClient(): string | null {
  if (typeof window === 'undefined') return null
  const cookies = document.cookie.split(';')
  const tokenCookie = cookies.find(cookie => cookie.trim().startsWith('accessToken='))
  return tokenCookie ? tokenCookie.split('=')[1] : null
}

export function getRefreshTokenClient(): string | null {
  if (typeof window === 'undefined') return null
  const cookies = document.cookie.split(';')
  const tokenCookie = cookies.find(cookie => cookie.trim().startsWith('refreshToken='))
  return tokenCookie ? tokenCookie.split('=')[1] : null
}

/**
 * Set access token cookie
 * @param accessToken - Access token string
 * @param maxAge - Max age in seconds (default: 24 hours)
 */
export function setAccessTokenCookie(accessToken: string, maxAge: number = 24 * 60 * 60) {
  if (typeof window === 'undefined') return
  const isProduction = process.env.NODE_ENV === 'production'
  const secureFlag = isProduction ? '; Secure' : ''
  document.cookie = `accessToken=${accessToken}; path=/; max-age=${maxAge}; SameSite=Lax${secureFlag}`
}

/**
 * Set refresh token cookie
 * @param refreshToken - Refresh token string
 * @param maxAge - Max age in seconds (default: 7 days)
 */
export function setRefreshTokenCookie(refreshToken: string, maxAge: number = 7 * 24 * 60 * 60) {
  if (typeof window === 'undefined') return
  const isProduction = process.env.NODE_ENV === 'production'
  const secureFlag = isProduction ? '; Secure' : ''
  // Note: HttpOnly cannot be set from client-side JavaScript for security reasons
  // This should be handled by server-side Set-Cookie headers in production
  document.cookie = `refreshToken=${refreshToken}; path=/; max-age=${maxAge}; SameSite=Lax${secureFlag}`
}

/**
 * Set both access and refresh token cookies
 * @param accessToken - Access token string
 * @param refreshToken - Refresh token string
 * @param accessMaxAge - Max age for access token in seconds (default: 24 hours)
 * @param refreshMaxAge - Max age for refresh token in seconds (default: 7 days)
 */
export function setAuthCookies(
  accessToken: string,
  refreshToken: string,
  accessMaxAge: number = 24 * 60 * 60,
  refreshMaxAge: number = 7 * 24 * 60 * 60
) {
  setAccessTokenCookie(accessToken, accessMaxAge)
  setRefreshTokenCookie(refreshToken, refreshMaxAge)
}

export function clearAuthCookieClient() {
  if (typeof window === 'undefined') return
  document.cookie = 'accessToken=; path=/; max-age=0'
  document.cookie = 'refreshToken=; path=/; max-age=0'
}

