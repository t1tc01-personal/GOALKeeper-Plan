'use client'
import { createContext, useCallback, useContext, useMemo } from 'react'
import { clearAuthCookie } from '../lib/authCookie'
import ENV_CONFIG from '@/config/env'
import { EMPLOYEE_ROLE, RoleEmployee } from '../constant/permission'

interface User {
  id: string
  email: string
  name: string
  roles: string[]
}

const AuthContext = createContext<{
  isAuthenticated: boolean
  handleLogout: () => Promise<void>
  handleLogin: () => Promise<void>
  user: User | null
  role: RoleEmployee
  isLoading: boolean
}>({
  isAuthenticated: false,
  handleLogout: async () => {},
  handleLogin: async () => {},
  user: null,
  role: EMPLOYEE_ROLE.NULL,
  isLoading: false,
})

function getRole(roles: string[]): RoleEmployee {
  if (roles.includes('admin')) {
    return EMPLOYEE_ROLE.ADMIN
  }
  if (roles.includes('editor')) {
    return EMPLOYEE_ROLE.EDITOR
  }
  return EMPLOYEE_ROLE.EMPLOYEE
}

function AuthProvider({ children, isAuthenticated }: { children: React.ReactNode; isAuthenticated: boolean }) {
  const handleLogout = useCallback(async () => {
    fetch(`${ENV_CONFIG.API_BASE_URL}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    })
      .then(async response => {
        if (response.ok) {
          await clearAuthCookie()
          window.location.reload()
        }
      })
      .catch(async error => {
        await clearAuthCookie()
        window.location.reload()
        console.error('Logout error:', error)
      })
  }, [])

  const handleLogin = useCallback(async () => {
    try {
      const response = await fetch(`${ENV_CONFIG.API_BASE_URL}/auth/login`, {
        method: 'GET',
      })
      const data = await response.json()
      if (data.data) {
        window.location.href = data.data
      } else {
        throw new Error('Login failed')
      }
    } catch (error) {
      console.error('Login error:', error)
    }
  }, [])

  const role = useMemo(() => {
    // TODO: Get user roles from API
    return EMPLOYEE_ROLE.EMPLOYEE
  }, [])

  return (
    <AuthContext.Provider
      value={{ isAuthenticated, handleLogout, handleLogin, user: null, role, isLoading: false }}
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

