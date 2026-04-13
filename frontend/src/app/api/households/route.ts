import { NextRequest, NextResponse } from 'next/server'
import { apiFetch, ApiError } from '@/lib/api'
import type { Household } from '@/lib/definitions'
import { setActiveHouseholdId } from '@/lib/household-selection'

export async function POST(req: NextRequest) {
  try {
    const body = await req.json()
    const data = await apiFetch<{ household: Household }>('/households', {
      method: 'POST',
      body: JSON.stringify(body),
    })
    await setActiveHouseholdId(data.household.id)
    return NextResponse.json(data)
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status })
    }
    return NextResponse.json({ error: 'Internal error' }, { status: 500 })
  }
}
