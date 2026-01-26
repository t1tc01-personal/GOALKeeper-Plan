/**
 * Generate OAuth authorization URL for GitHub or Google
 * @param provider - OAuth provider ('github' or 'google')
 * @returns OAuth authorization URL
 */
export function generateOAuthURL(provider: 'github' | 'google'): string {
  let baseUrl: string
  
  if (typeof window !== 'undefined') {
    // Client-side: use window.location.origin (always has protocol and port)
    baseUrl = window.location.origin
  } else {
    // Server-side: validate and fix NEXT_PUBLIC_DOMAIN
    baseUrl = process.env.NEXT_PUBLIC_DOMAIN || 'http://localhost:3000'
    
    // Ensure baseUrl has protocol
    if (!baseUrl.startsWith('http://') && !baseUrl.startsWith('https://')) {
      // If it's just a hostname, assume http:// for localhost, https:// for others
      if (baseUrl.includes('localhost') || baseUrl.includes('127.0.0.1')) {
        baseUrl = `http://${baseUrl}`
        // Add port if not present for localhost
        if (!baseUrl.includes(':')) {
          baseUrl = `${baseUrl}:3000`
        }
      } else {
        baseUrl = `https://${baseUrl}`
      }
    }
  }
  
  // Remove trailing slash
  const cleanBaseUrl = baseUrl.replace(/\/$/, '')
  const redirectUri = `${cleanBaseUrl}/auth/${provider}/callback`
  
  // Generate secure state parameter for CSRF protection
  const state = generateSecureState()
  
  // Store state in sessionStorage to validate on callback
  if (typeof window !== 'undefined') {
    sessionStorage.setItem(`oauth_state_${provider}`, state)
  }
  
  if (provider === 'github') {
    // Use NEXT_PUBLIC_ prefix (exposed via next.config.ts from GITHUB_CLIENT_ID)
    const clientId = process.env.NEXT_PUBLIC_GITHUB_CLIENT_ID
    if (!clientId) {
      throw new Error('GITHUB_CLIENT_ID is not configured. Please set it in .env file.')
    }
    
    const params = new URLSearchParams({
      client_id: clientId,
      redirect_uri: redirectUri,
      scope: 'user:email',
      state: state,
    })
    
    return `https://github.com/login/oauth/authorize?${params.toString()}`
  } else if (provider === 'google') {
    // Use NEXT_PUBLIC_ prefix (exposed via next.config.ts from GOOGLE_CLIENT_ID)
    const clientId = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID
    if (!clientId) {
      throw new Error('GOOGLE_CLIENT_ID is not configured. Please set it in .env file.')
    }
    
    const params = new URLSearchParams({
      client_id: clientId,
      redirect_uri: redirectUri,
      response_type: 'code',
      scope: 'openid profile email',
      state: state,
    })
    
    return `https://accounts.google.com/o/oauth2/v2/auth?${params.toString()}`
  } else {
    throw new Error(`Unsupported OAuth provider: ${provider}`)
  }
}

/**
 * Generate a cryptographically secure random state string
 * Uses crypto.randomUUID() if available, otherwise falls back to Math.random()
 */
function generateSecureState(): string {
  if (typeof window !== 'undefined' && window.crypto && window.crypto.randomUUID) {
    return window.crypto.randomUUID()
  }
  
  // Fallback for older browsers
  return `${Date.now()}-${Math.random().toString(36).substring(2, 15)}-${Math.random().toString(36).substring(2, 15)}`
}

/**
 * Validate OAuth state parameter from callback
 * @param provider - OAuth provider
 * @param state - State parameter from OAuth callback
 * @returns true if state is valid
 */
export function validateOAuthState(provider: 'github' | 'google', state: string): boolean {
  if (typeof window === 'undefined') return false
  
  const storedState = sessionStorage.getItem(`oauth_state_${provider}`)
  if (!storedState) return false
  
  const isValid = storedState === state
  
  // Clean up stored state after validation
  sessionStorage.removeItem(`oauth_state_${provider}`)
  
  return isValid
}

