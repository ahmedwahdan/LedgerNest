'use server'

import { redirect } from 'next/navigation'
import { setTokens, clearTokens, getRefreshToken } from '@/lib/session'
import { apiFetch, ApiError } from '@/lib/api'
import type { User } from '@/lib/definitions'
import type { ActionState } from '@/lib/definitions'

interface AuthResponse {
  access_token: string
  refresh_token: string
  expires_at: string
  user: User
}

export async function login(_prev: ActionState, formData: FormData): Promise<ActionState> {
  const email = formData.get('email') as string
  const password = formData.get('password') as string

  if (!email || !password) {
    return { success: false, error: 'Email and password are required.' }
  }

  try {
    const data = await apiFetch<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
      token: '', // no token for login
    })
    await setTokens(data.access_token, data.refresh_token, data.expires_at)
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401 || err.status === 400) {
        return { success: false, error: 'Invalid email or password.' }
      }
    }
    return { success: false, error: 'Something went wrong. Please try again.' }
  }

  redirect('/dashboard')
}

export async function register(_prev: ActionState, formData: FormData): Promise<ActionState> {
  const email = formData.get('email') as string
  const password = formData.get('password') as string
  const displayName = formData.get('display_name') as string

  if (!email || !password || !displayName) {
    return { success: false, error: 'All fields are required.' }
  }

  if (password.length < 8) {
    return { success: false, error: 'Password must be at least 8 characters.' }
  }

  try {
    const data = await apiFetch<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, display_name: displayName }),
      token: '',
    })
    await setTokens(data.access_token, data.refresh_token, data.expires_at)
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 409) {
        return { success: false, error: 'An account with this email already exists.' }
      }
    }
    return { success: false, error: 'Something went wrong. Please try again.' }
  }

  redirect('/dashboard')
}

export async function logout() {
  try {
    const refreshToken = await getRefreshToken()
    if (refreshToken) {
      await apiFetch('/auth/logout', {
        method: 'POST',
        body: JSON.stringify({ refresh_token: refreshToken }),
      })
    }
  } catch {
    // best-effort
  }
  await clearTokens()
  redirect('/login')
}

export async function refreshAccessToken(): Promise<string | null> {
  try {
    const refreshToken = await getRefreshToken()
    if (!refreshToken) return null

    const data = await apiFetch<AuthResponse>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
      token: '',
    })
    await setTokens(data.access_token, data.refresh_token, data.expires_at)
    return data.access_token
  } catch {
    return null
  }
}
