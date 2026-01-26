'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams, useSearchParams } from 'next/navigation'
import { authApi } from '@/features/auth/api/authApi'
import { useAuthStore } from '@/features/auth/store/authStore'
import { toast } from 'sonner'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { isLeft } from '@/shared/utils/either'
import { setAuthCookies } from '@/shared/lib/authCookieClient'
import { validateOAuthState } from '@/shared/lib/oauth'

export default function OAuthProviderCallbackPage() {
  const router = useRouter()
  const params = useParams()
  const searchParams = useSearchParams()
  const [isLoading, setIsLoading] = useState(true)
  const { setAuth } = useAuthStore()

  const provider = params.provider as 'github' | 'google'
  const code = searchParams.get('code')
  const state = searchParams.get('state')
  const error = searchParams.get('error')
  const errorDescription = searchParams.get('error_description')

  useEffect(() => {
    const handleCallback = async () => {
      // Check for OAuth error
      if (error) {
        toast.error(`OAuth error: ${errorDescription || error}`)
        router.push('/login')
        return
      }

      // Check for code
      if (!code) {
        toast.error('Missing authorization code')
        router.push('/login')
        return
      }

      // Validate state parameter for CSRF protection
      if (state && !validateOAuthState(provider, state)) {
        toast.error('Invalid OAuth state. Please try again.')
        router.push('/login')
        return
      }

      try {
        // Exchange code for access token from provider via Next.js API route
        const exchangeResponse = await fetch(`/api/auth/${provider}/exchange`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ code }),
        })

        if (!exchangeResponse.ok) {
          const errorData = await exchangeResponse.json()
          throw new Error(errorData.error || 'Failed to exchange code for token')
        }

        const exchangeData = await exchangeResponse.json()
        const accessToken = exchangeData.access_token

        if (!accessToken) {
          throw new Error('Failed to get access token from provider')
        }

        // Call backend callback with access token
        const result = await authApi.oauthTokenCallback(provider, accessToken)

        if (isLeft(result)) {
          const error = result.left
          if (error.code === 'OAUTH_EMAIL_EXISTS' || error.code === 'EMAIL_ALREADY_EXISTS') {
            toast.error('This email is already associated with a different OAuth provider')
          } else {
            toast.error(error.message || 'Failed to authenticate')
          }
          router.push('/login')
          return
        }

        const authResponse = result.right.data
        if (authResponse) {
          // Set auth state
          setAuth(authResponse)

          // Save tokens to cookies
          setAuthCookies(
            authResponse.access_token,
            authResponse.refresh_token
          )

          toast.success('Login successful!')
          router.push('/')
        }
      } catch (error) {
        toast.error('An unexpected error occurred')
        console.error('OAuth callback error:', error)
        router.push('/login')
      } finally {
        setIsLoading(false)
      }
    }

    handleCallback()
  }, [code, state, error, errorDescription, provider, router, setAuth])

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-neutral-50 p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Completing authentication...</CardTitle>
            <CardDescription>
              Please wait while we complete your {provider} login.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex justify-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return null
}

