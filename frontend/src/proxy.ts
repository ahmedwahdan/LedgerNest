import { NextRequest, NextResponse } from 'next/server'

const protectedPrefixes = ['/dashboard', '/expenses', '/categories', '/budgets', '/households', '/settings', '/onboarding', '/analytics']
const authPaths = ['/login', '/register']
const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'
const THIRTY_DAYS_MS = 30 * 24 * 60 * 60 * 1000

export async function proxy(req: NextRequest) {
  const { pathname } = req.nextUrl
  const accessToken = req.cookies.get('access_token')?.value
  const refreshToken = req.cookies.get('refresh_token')?.value

  const isProtected = protectedPrefixes.some((p) => pathname.startsWith(p))
  const isAuthPage = authPaths.some((p) => pathname === p)

  if (isProtected && accessToken) {
    return NextResponse.next()
  }

  if (isProtected && refreshToken) {
    const refreshed = await refreshFromProxy(refreshToken)
    if (refreshed) {
      const response = NextResponse.next()
      setAuthCookies(response, refreshed)
      return response
    }
  }

  if (isProtected) {
    const url = req.nextUrl.clone()
    url.pathname = '/login'
    const response = NextResponse.redirect(url)
    clearAuthCookies(response)
    return response
  }

  if (isAuthPage && accessToken) {
    const url = req.nextUrl.clone()
    url.pathname = '/dashboard'
    return NextResponse.redirect(url)
  }

  if (isAuthPage && refreshToken) {
    const refreshed = await refreshFromProxy(refreshToken)
    if (refreshed) {
      const url = req.nextUrl.clone()
      url.pathname = '/dashboard'
      const response = NextResponse.redirect(url)
      setAuthCookies(response, refreshed)
      return response
    }
  }

  return NextResponse.next()
}

async function refreshFromProxy(refreshToken: string) {
  try {
    const res = await fetch(`${BACKEND_URL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
      cache: 'no-store',
    })

    if (!res.ok) {
      return null
    }

    return await res.json() as {
      access_token: string
      refresh_token: string
      expires_at: string
    }
  } catch {
    return null
  }
}

function setAuthCookies(
  response: NextResponse,
  session: { access_token: string; refresh_token: string; expires_at: string },
) {
  response.cookies.set('access_token', session.access_token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    expires: new Date(session.expires_at),
    path: '/',
  })

  response.cookies.set('refresh_token', session.refresh_token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    expires: new Date(Date.now() + THIRTY_DAYS_MS),
    path: '/',
  })
}

function clearAuthCookies(response: NextResponse) {
  response.cookies.delete('access_token')
  response.cookies.delete('refresh_token')
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
}
