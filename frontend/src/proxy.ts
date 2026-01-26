import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { verifyInMiddleware, getUserRole } from './shared/lib/middleware'

export async function proxy(req: NextRequest) {
  const pathname = req.nextUrl.pathname
  const homeUrl = new URL('/', req.url)
  const loginUrl = new URL('/login', req.url)
  const isAuth = await verifyInMiddleware(req)
  const roles = isAuth ? await getUserRole(req) : []

  // Allow OAuth callback routes and signup
  const isAuthRoute = pathname === '/login' || 
                      pathname === '/signup' || 
                      pathname.startsWith('/auth/') ||
                      pathname.startsWith('/api/')

  if (!isAuth && !isAuthRoute) {
    return NextResponse.redirect(loginUrl)
  }

  if (isAuth && pathname === '/login') {
    return NextResponse.redirect(homeUrl)
  }

  if (pathname.startsWith('/dashboard') && !roles.includes('admin') && !roles.includes('editor')) {
    return NextResponse.redirect(homeUrl)
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!_next|api|favicon.ico|.*\\..*).*)'],
}

