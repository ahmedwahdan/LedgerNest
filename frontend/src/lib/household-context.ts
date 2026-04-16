import 'server-only'

import { cache } from 'react'
import { apiFetch } from '@/lib/api'
import type { Household } from '@/lib/definitions'
import { clearActiveHouseholdId, getActiveHouseholdId } from './household-selection'

export const getHouseholds = cache(async (): Promise<Household[]> => {
  try {
    const data = await apiFetch<{ households: Household[] }>('/households')
    return data.households ?? []
  } catch {
    return []
  }
})

export const getActiveHousehold = cache(async (): Promise<Household | null> => {
  const households = await getHouseholds()
  if (households.length === 0) {
    await clearActiveHouseholdId()
    return null
  }

  const activeHouseholdId = await getActiveHouseholdId()
  if (!activeHouseholdId) {
    return households[0]
  }

  const selected = households.find((household) => household.id === activeHouseholdId)
  if (!selected) {
    await clearActiveHouseholdId()
    return households[0]
  }

  return selected
})
