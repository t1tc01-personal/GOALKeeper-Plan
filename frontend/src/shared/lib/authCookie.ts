'use server'
import { cookies } from 'next/headers'

export async function getAuthToken(): Promise<string | null> {
  const store = await cookies()
  return store.get('accessToken')?.value ?? null
}

export async function getRefreshToken(): Promise<string | null> {
  const store = await cookies()
  return store.get('refreshToken')?.value ?? null
}

export async function canAccess() {
  const token = await getAuthToken()
  if (!token) return false
  return !!token
}

export async function setAccessTokenCookie(accessToken: string, maxAge: number) {
  const store = await cookies()
  store.set('accessToken', accessToken, {
    httpOnly: false,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    maxAge: Number(maxAge),
  })
}

export async function setRefreshTokenCookie(refreshToken: string, maxAge: number) {
  const store = await cookies()
  store.set('refreshToken', refreshToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    maxAge: Number(maxAge),
  })
}

export async function setAuthCookies(accessToken: string, refreshToken: string, accessMaxAge: number, refreshMaxAge: number) {
  await setAccessTokenCookie(accessToken, accessMaxAge)
  await setRefreshTokenCookie(refreshToken, refreshMaxAge)
}

export async function clearAuthCookie() {
  const store = await cookies()
  store.set('accessToken', '', {
    path: '/',
    maxAge: 0,
  })
  store.set('refreshToken', '', {
    path: '/',
    maxAge: 0,
  })
}

