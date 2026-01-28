'use client'
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { useAuthStore } from '@/features/auth/store/authStore'
import { authApi } from '@/features/auth/api/authApi'
import { clearAuthCookieClient } from '../lib/authCookieClient'
import { EMPLOYEE_ROLE, RoleEmployee } from '../constant/permission'
import { toast } from 'sonner'
import { User } from '@/shared/types/auth'

interface UserWithRoles extends User {
  roles?: string[]
}

const AuthContext = createContext<{
  isAuthenticated: boolean
  handleLogout: () => Promise<void>
  user: User | null
  role: RoleEmployee
  isLoading: boolean
}>({
  isAuthenticated: false,
  handleLogout: async () => {},
  user: null,
  role: EMPLOYEE_ROLE.NULL,
  isLoading: false,
})

function getRole(roles?: string[]): RoleEmployee {
  if (!roles) return EMPLOYEE_ROLE.EMPLOYEE
  if (roles.includes('admin')) {
    return EMPLOYEE_ROLE.ADMIN
  }
  if (roles.includes('editor')) {
    return EMPLOYEE_ROLE.EDITOR
  }
  return EMPLOYEE_ROLE.EMPLOYEE
}

function AuthProvider({ children }: { children: React.ReactNode }) {
  const { user, accessToken, isAuthenticated, clearAuth } = useAuthStore()
  const [isLoading, setIsLoading] = useState(false)

  // Sync tokens to cookies on mount and when tokens change
  useEffect(() => {
    if (typeof window === 'undefined') return
    
    if (accessToken) {
      const { accessToken: token, refreshToken } = useAuthStore.getState()
      if (token) {
        // Set access token cookie (24 hours)
        document.cookie = `accessToken=${token}; path=/; max-age=${24 * 60 * 60}; SameSite=Lax`
      }
      if (refreshToken) {
        // Set refresh token cookie (7 days)
        document.cookie = `refreshToken=${refreshToken}; path=/; max-age=${7 * 24 * 60 * 60}; SameSite=Lax; HttpOnly`
      }
    }
  }, [accessToken])

  const handleLogout = useCallback(async () => {
    setIsLoading(true)
    try {
      const { refreshToken } = useAuthStore.getState()
      if (refreshToken) {
        await authApi.logout({ refresh_token: refreshToken })
      }
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      clearAuthCookieClient()
      clearAuth()
      toast.success('Logged out successfully')
      window.location.href = '/login'
    }
  }, [clearAuth])

  const userWithRoles = user as UserWithRoles | null
  const role = useMemo(() => getRole(userWithRoles?.roles), [userWithRoles?.roles])

  return (
    <AuthContext.Provider
      value={{ isAuthenticated, handleLogout, user, role, isLoading }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export default AuthProvider

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within a AuthProvider')
  }
  return context
}

