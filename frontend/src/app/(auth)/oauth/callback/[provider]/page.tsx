'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams, useSearchParams } from 'next/navigation'
import { authApi } from '@/features/auth/api/authApi'
import { useAuthStore } from '@/features/auth/store/authStore'
import { toast } from 'sonner'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { isLeft } from '@/shared/utils/either'
import { setAuthCookies } from '@/shared/lib/authCookieClient'

export default function OAuthCallbackPage() {
  const router = useRouter()
  const params = useParams()
  const searchParams = useSearchParams()
  const [isLoading, setIsLoading] = useState(true)
  const { setAuth } = useAuthStore()

  const provider = params.provider as string
  const tokenId = searchParams.get('token_id')
  const error = searchParams.get('error')

  useEffect(() => {
    const handleCallback = async () => {
      // Check for OAuth error
      if (error) {
        toast.error(`OAuth error: ${error}`)
        router.push('/login')
        return
      }

      // Check for token_id
      if (!tokenId) {
        toast.error('Missing token ID')
        router.push('/login')
        return
      }

      try {
        // Get tokens from backend
        const result = await authApi.getOAuthTokens(tokenId)
        
        if (isLeft(result)) {
          const error = result.left
          toast.error(error.message || 'Failed to retrieve tokens')
          router.push('/login')
          return
        }

        const tokenData = result.right.data
        if (tokenData) {
          // Normalize token data to AuthResponse format
          const authResponse = {
            access_token: tokenData.access_token,
            refresh_token: tokenData.refresh_token,
            user: {
              id: tokenData.user_id,
              email: tokenData.email,
              name: tokenData.name,
            },
          }
          
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
  }, [tokenId, error, router, setAuth])

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

