'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/features/auth/store/authStore'
import { authApi } from '@/features/auth/api/authApi'
import { clearAuthCookieClient } from '@/shared/lib/authCookieClient'
import { isLeft } from '@/shared/utils/either'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'sonner'

export default function Home() {
  const router = useRouter()
  const { user, isAuthenticated, clearAuth, refreshToken } = useAuthStore()
  const [isChecking, setIsChecking] = useState(true)
  const [isLoggingOut, setIsLoggingOut] = useState(false)

  useEffect(() => {
    // Wait a bit for auth store to hydrate from localStorage
    const timer = setTimeout(() => {
      setIsChecking(false)
      if (!isAuthenticated) {
        router.push('/login')
      }
    }, 100)

    return () => clearTimeout(timer)
  }, [isAuthenticated, router])

  const handleLogout = async () => {
    setIsLoggingOut(true)
    try {
      // Call backend logout API if refresh token exists
      if (refreshToken) {
        const result = await authApi.logout({ refresh_token: refreshToken })
        if (isLeft(result)) {
          console.error('Logout error:', result.left)
          // Continue with logout even if API call fails
        }
      }
    } catch (error) {
      console.error('Logout error:', error)
      // Continue with logout even if API call fails
    } finally {
      // Clear cookies
      clearAuthCookieClient()
      // Clear auth state
      clearAuth()
      toast.success('Logged out successfully')
      router.push('/login')
      setIsLoggingOut(false)
    }
  }

  if (isChecking || !isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-neutral-50">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-neutral-50 to-neutral-100">
      <div className="container mx-auto px-4 py-8">
        {/* Header */}
        <header className="mb-8">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-4xl font-bold text-neutral-900">GOALKeeper Plan</h1>
              <p className="text-neutral-600 mt-2">Your goal management and planning platform</p>
            </div>
            <Button onClick={handleLogout} variant="outline" disabled={isLoggingOut}>
              {isLoggingOut ? 'Logging out...' : 'Logout'}
            </Button>
          </div>
        </header>

        {/* Welcome Card */}
        <Card className="mb-8">
          <CardHeader>
            <CardTitle className="text-2xl">Welcome back, {user?.name || 'User'}!</CardTitle>
            <CardDescription>
              Manage your goals and track your progress
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <p className="text-sm text-neutral-600">
                <span className="font-semibold">Email:</span> {user?.email}
              </p>
              {user?.is_email_verified && (
                <p className="text-sm text-green-600">
                  ✓ Email verified
                </p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Features Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Set Goals</CardTitle>
              <CardDescription>
                Create and manage your personal and professional goals
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-neutral-600">
                Define clear objectives and break them down into actionable steps.
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Track Progress</CardTitle>
              <CardDescription>
                Monitor your progress and stay motivated
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-neutral-600">
                Visualize your achievements and see how far you've come.
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Plan Ahead</CardTitle>
              <CardDescription>
                Organize your tasks and plan for the future
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-neutral-600">
                Schedule your activities and never miss important deadlines.
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Quick Actions */}
        <Card className="mt-8">
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-4">
              <Button>Create New Goal</Button>
              <Button variant="outline">View All Goals</Button>
              <Button variant="outline">View Statistics</Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
