import { NextResponse } from 'next/server'
import { getRefreshToken, setTokens, clearTokens } from '@/lib/session'

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

// POST /api/auth/refresh — called server-side to exchange a refresh token for new tokens.
// Route Handlers are allowed to write cookies; Server Components are not.
export async function POST() {
  const refreshToken = await getRefreshToken()
  if (!refreshToken) {
    return NextResponse.json({ error: 'no refresh token' }, { status: 401 })
  }

  const res = await fetch(`${BACKEND_URL}/auth/refresh`, {
    method: 'POST',
    cache: 'no-store',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  })

  if (!res.ok) {
    await clearTokens()
    return NextResponse.json({ error: 'refresh failed' }, { status: 401 })
  }

  const data = (await res.json()) as {
    access_token: string
    refresh_token: string
    expires_at: string
  }

  await setTokens(data.access_token, data.refresh_token, data.expires_at)
  return NextResponse.json({ access_token: data.access_token })
}
