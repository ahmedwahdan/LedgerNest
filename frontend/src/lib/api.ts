import 'server-only'
import { clearTokens, getAccessToken, getRefreshToken, setTokens } from './session'

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export async function apiFetch<T = unknown>(
  path: string,
  init?: RequestInit & { token?: string },
): Promise<T> {
  const token = init?.token ?? (await getAccessToken())
  return apiFetchWithToken<T>(path, init, token, true)
}

async function apiFetchWithToken<T>(
  path: string,
  init: (RequestInit & { token?: string }) | undefined,
  token: string | undefined,
  allowRefreshRetry: boolean,
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }

  // merge any caller-supplied headers (excluding our managed ones)
  const callerHeaders = init?.headers as Record<string, string> | undefined
  if (callerHeaders) {
    for (const [k, v] of Object.entries(callerHeaders)) {
      headers[k] = v
    }
  }

  const restInit = init ? { ...init } : {}
  delete restInit.token
  delete restInit.headers

  const res = await fetch(`${BACKEND_URL}${path}`, {
    cache: 'no-store',
    ...restInit,
    headers,
  })

  if (res.status === 401 && allowRefreshRetry && shouldRetryWithRefresh(path)) {
    const refreshedToken = await refreshSession()
    if (refreshedToken) {
      return apiFetchWithToken<T>(path, init, refreshedToken, false)
    }
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({})) as { error?: string }
    throw new ApiError(res.status, body.error ?? `HTTP ${res.status}`)
  }

  // 204 No Content
  if (res.status === 204) return undefined as T

  return res.json() as Promise<T>
}

function shouldRetryWithRefresh(path: string) {
  return !path.startsWith('/auth/login') &&
    !path.startsWith('/auth/register') &&
    !path.startsWith('/auth/refresh') &&
    !path.startsWith('/auth/logout')
}

async function refreshSession(): Promise<string | null> {
  const refreshToken = await getRefreshToken()
  if (!refreshToken) {
    return null
  }

  try {
    const res = await fetch(`${BACKEND_URL}/auth/refresh`, {
      method: 'POST',
      cache: 'no-store',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (!res.ok) {
      await clearTokens()
      return null
    }

    const data = await res.json() as {
      access_token: string
      refresh_token: string
      expires_at: string
    }

    await setTokens(data.access_token, data.refresh_token, data.expires_at)
    return data.access_token
  } catch {
    await clearTokens()
    return null
  }
}
