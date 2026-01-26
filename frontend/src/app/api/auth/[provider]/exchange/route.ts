import { NextRequest, NextResponse } from 'next/server'

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ provider: string }> }
) {
  try {
    const { provider } = await params
    const body = await request.json()
    const { code } = body

    console.log('OAuth exchange request:', { provider, hasCode: !!code })

    if (!code) {
      return NextResponse.json(
        { error: 'Missing authorization code' },
        { status: 400 }
      )
    }

    if (!provider || (provider !== 'github' && provider !== 'google')) {
      console.error('Invalid provider:', provider)
      return NextResponse.json(
        { error: `Invalid provider: ${provider}. Must be 'github' or 'google'` },
        { status: 400 }
      )
    }

    let accessToken: string | null = null

    if (provider === 'github') {
      // Check environment variables
      const githubClientId = process.env.GITHUB_CLIENT_ID
      const githubClientSecret = process.env.GITHUB_CLIENT_SECRET

      if (!githubClientId || !githubClientSecret) {
        console.error('Missing GitHub OAuth credentials')
        return NextResponse.json(
          { error: 'GitHub OAuth credentials not configured' },
          { status: 500 }
        )
      }

      // Exchange code for GitHub access token
      const response = await fetch('https://github.com/login/oauth/access_token', {
        method: 'POST',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
          client_id: githubClientId,
          client_secret: githubClientSecret,
          code: code,
        }),
      })

      if (!response.ok) {
        const errorText = await response.text()
        console.error('GitHub OAuth error:', {
          status: response.status,
          statusText: response.statusText,
          body: errorText,
        })
        return NextResponse.json(
          { error: `Failed to exchange GitHub code for token: ${response.statusText}` },
          { status: response.status }
        )
      }

      const data = await response.json()
      if (data.error) {
        console.error('GitHub OAuth response error:', data)
        return NextResponse.json(
          { error: data.error_description || data.error },
          { status: 400 }
        )
      }

      accessToken = data.access_token
    } else if (provider === 'google') {
      // Check environment variables
      const googleClientId = process.env.GOOGLE_CLIENT_ID
      const googleClientSecret = process.env.GOOGLE_CLIENT_SECRET

      if (!googleClientId || !googleClientSecret) {
        console.error('Missing Google OAuth credentials')
        return NextResponse.json(
          { error: 'Google OAuth credentials not configured' },
          { status: 500 }
        )
      }

      // Exchange code for Google access token
      // Ensure redirect URI matches exactly with Google Cloud Console configuration
      // Get base URL from environment or request headers
      let baseUrl = process.env.NEXT_PUBLIC_DOMAIN
      
      // If NEXT_PUBLIC_DOMAIN is not set or doesn't have protocol, use request origin
      if (!baseUrl || !baseUrl.startsWith('http://') && !baseUrl.startsWith('https://')) {
        const origin = request.headers.get('origin') || request.headers.get('referer')
        if (origin) {
          try {
            const url = new URL(origin)
            baseUrl = url.origin
          } catch {
            // Fallback to default if URL parsing fails
            baseUrl = 'http://localhost:3000'
          }
        } else {
          baseUrl = 'http://localhost:3000'
        }
      }
      
      // Remove trailing slash if present
      const cleanBaseUrl = baseUrl.replace(/\/$/, '')
      const redirectUri = `${cleanBaseUrl}/auth/${provider}/callback`
      
      const response = await fetch('https://oauth2.googleapis.com/token', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
          client_id: googleClientId,
          client_secret: googleClientSecret,
          code: code,
          grant_type: 'authorization_code',
          redirect_uri: redirectUri,
        }),
      })

      if (!response.ok) {
        const errorText = await response.text()
        let errorData: any = {}
        try {
          errorData = JSON.parse(errorText)
        } catch {
          errorData = { raw: errorText }
        }
        
        console.error('Google OAuth error details:', {
          status: response.status,
          statusText: response.statusText,
          error: errorData.error,
          errorDescription: errorData.error_description,
          redirectUri: redirectUri,
          fullResponse: errorText,
        })
        
        // Provide more helpful error message
        const errorMessage = errorData.error_description || errorData.error || `Failed to exchange Google code for token: ${response.statusText}`
        return NextResponse.json(
          { 
            error: errorMessage,
            details: errorData.error === 'invalid_request' 
              ? 'Please check: 1) OAuth consent screen is configured correctly, 2) Test users are added (if in Testing mode), 3) Redirect URI matches exactly in Google Cloud Console'
              : undefined
          },
          { status: response.status }
        )
      }

      const data = await response.json()
      if (data.error) {
        console.error('Google OAuth response error:', data)
        return NextResponse.json(
          { error: data.error_description || data.error },
          { status: 400 }
        )
      }

      accessToken = data.access_token
    } else {
      return NextResponse.json(
        { error: 'Invalid provider' },
        { status: 400 }
      )
    }

    if (!accessToken) {
      return NextResponse.json(
        { error: 'Failed to get access token from provider' },
        { status: 500 }
      )
    }

    return NextResponse.json({ access_token: accessToken })
  } catch (error) {
    console.error('OAuth exchange error:', error)
    const errorMessage = error instanceof Error ? error.message : 'Unknown error'
    return NextResponse.json(
      { error: `Internal server error: ${errorMessage}` },
      { status: 500 }
    )
  }
}

