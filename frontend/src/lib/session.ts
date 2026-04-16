import 'server-only'
import { cookies } from 'next/headers'

const ACCESS_TOKEN_COOKIE = 'access_token'
const REFRESH_TOKEN_COOKIE = 'refresh_token'

export async function getAccessToken(): Promise<string | undefined> {
  const store = await cookies()
  return store.get(ACCESS_TOKEN_COOKIE)?.value
}

export async function getRefreshToken(): Promise<string | undefined> {
  const store = await cookies()
  return store.get(REFRESH_TOKEN_COOKIE)?.value
}

export async function setTokens(accessToken: string, refreshToken: string, expiresAt: string) {
  const store = await cookies()
  const accessExpires = new Date(expiresAt)
  const refreshExpires = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000) // 30 days

  store.set(ACCESS_TOKEN_COOKIE, accessToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    expires: accessExpires,
    path: '/',
  })

  store.set(REFRESH_TOKEN_COOKIE, refreshToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    expires: refreshExpires,
    path: '/',
  })
}

export async function clearTokens() {
  const store = await cookies()
  store.delete(ACCESS_TOKEN_COOKIE)
  store.delete(REFRESH_TOKEN_COOKIE)
}
