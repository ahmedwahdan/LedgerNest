import { NextRequest, NextResponse } from 'next/server'

const protectedPrefixes = ['/dashboard', '/expenses', '/budgets', '/households', '/settings', '/onboarding']
const authPaths = ['/login', '/register']

export function proxy(req: NextRequest) {
  const { pathname } = req.nextUrl
  const accessToken = req.cookies.get('access_token')?.value

  const isProtected = protectedPrefixes.some((p) => pathname.startsWith(p))
  const isAuthPage = authPaths.some((p) => pathname === p)

  if (isProtected && !accessToken) {
    const url = req.nextUrl.clone()
    url.pathname = '/login'
    return NextResponse.redirect(url)
  }

  if (isAuthPage && accessToken) {
    const url = req.nextUrl.clone()
    url.pathname = '/dashboard'
    return NextResponse.redirect(url)
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
}
