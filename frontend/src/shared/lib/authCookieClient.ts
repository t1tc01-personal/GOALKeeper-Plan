'use client'

export function getAuthTokenClient(): string | null {
  if (typeof window === 'undefined') return null
  const cookies = document.cookie.split(';')
  const tokenCookie = cookies.find(cookie => cookie.trim().startsWith('infoToken='))
  return tokenCookie ? tokenCookie.split('=')[1] : null
}

export function clearAuthCookieClient() {
  if (typeof window === 'undefined') return
  document.cookie = 'infoToken=; path=/; max-age=0'
}

