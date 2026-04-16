import 'server-only'

import { cookies } from 'next/headers'

const ACTIVE_HOUSEHOLD_COOKIE = 'active_household_id'

export async function getActiveHouseholdId(): Promise<string | undefined> {
  const store = await cookies()
  return store.get(ACTIVE_HOUSEHOLD_COOKIE)?.value
}

export async function setActiveHouseholdId(householdId: string) {
  const store = await cookies()
  store.set(ACTIVE_HOUSEHOLD_COOKIE, householdId, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    maxAge: 60 * 60 * 24 * 365,
  })
}

export async function clearActiveHouseholdId() {
  const store = await cookies()
  store.delete(ACTIVE_HOUSEHOLD_COOKIE)
}
