import 'server-only'
import { getAccessToken } from './session'

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

  const { token: _drop, headers: _h, ...restInit } = init ?? {}

  const res = await fetch(`${BACKEND_URL}${path}`, {
    cache: 'no-store',
    ...restInit,
    headers,
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({})) as { error?: string }
    throw new ApiError(res.status, body.error ?? `HTTP ${res.status}`)
  }

  // 204 No Content
  if (res.status === 204) return undefined as T

  return res.json() as Promise<T>
}
