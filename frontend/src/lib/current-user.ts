import 'server-only'

import { cache } from 'react'
import { apiFetch, ApiError } from '@/lib/api'
import type { User } from '@/lib/definitions'

export const getCurrentUser = cache(async (): Promise<User | null> => {
  try {
    const data = await apiFetch<{ user: User }>('/auth/me')
    return data.user
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      return null
    }

    throw err
  }
})
