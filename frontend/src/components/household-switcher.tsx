'use client'

import { useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { selectActiveHousehold } from '@/actions/households'
import type { Household } from '@/lib/definitions'

export function HouseholdSwitcher({
  households,
  activeHouseholdId,
}: {
  households: Household[]
  activeHouseholdId?: string
}) {
  const [pending, startTransition] = useTransition()
  const router = useRouter()

  if (households.length === 0) {
    return (
      <div className="rounded-full border border-[var(--line)] bg-white/60 px-3 py-1.5 text-xs text-muted">
        No household
      </div>
    )
  }

  return (
    <label className="flex items-center gap-2 rounded-full border border-[var(--line)] bg-white/60 px-3 py-1.5 text-xs text-muted">
      <span className="hidden sm:inline">Household</span>
      <select
        aria-label="Select active household"
        defaultValue={activeHouseholdId ?? households[0]?.id}
        disabled={pending}
        onChange={(event) => {
          const nextId = event.target.value
          startTransition(async () => {
            await selectActiveHousehold(nextId)
            router.refresh()
          })
        }}
        className="min-w-0 bg-transparent text-xs outline-none"
      >
        {households.map((household) => (
          <option key={household.id} value={household.id}>
            {household.name}
          </option>
        ))}
      </select>
    </label>
  )
}
