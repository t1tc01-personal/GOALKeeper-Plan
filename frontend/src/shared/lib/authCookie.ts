'use server'
import { cookies } from 'next/headers'

export async function getAuthToken(): Promise<string | null> {
  const store = await cookies()
  return store.get('infoToken')?.value ?? null
}

export async function canAccess() {
  const token = await getAuthToken()
  if (!token) return false
  return !!token
}

export async function setAccessTokenCookie(accessToken: string, maxAge: number) {
  const store = await cookies()
  store.set('infoToken', accessToken, {
    httpOnly: false,
    secure: true, 
    sameSite: 'none',
    path: '/',
    maxAge: Number(maxAge),
  })
}

export async function clearAuthCookie() {
  const store = await cookies()
  store.set('infoToken', '', {
    path: '/',
    maxAge: 0,
  })
}

