'use client'

import { useTransition } from 'react'
import { deleteHousehold, leaveHousehold } from '@/actions/households'

export function HouseholdActionsPanel({
  householdId,
  householdName,
}: {
  householdId: string
  householdName: string
}) {
  const [pending, startTransition] = useTransition()

  return (
    <section className="glass-panel rounded-[2rem] p-6">
      <h2 className="mb-4 text-sm font-medium uppercase tracking-[0.2em] text-muted">
        Danger zone
      </h2>
      <div className="flex flex-wrap gap-3">
        <button
          disabled={pending}
          onClick={() => {
            if (!confirm(`Leave "${householdName}"? You will lose access to shared expenses.`)) return
            startTransition(() => leaveHousehold(householdId))
          }}
          className="rounded-full border border-[var(--line)] px-5 py-2.5 text-sm text-muted transition hover:bg-white/70 disabled:opacity-50"
        >
          Leave household
        </button>
        <button
          disabled={pending}
          onClick={() => {
            if (!confirm(`Permanently delete "${householdName}"? This cannot be undone.`)) return
            startTransition(() => deleteHousehold(householdId))
          }}
          className="rounded-full border border-red-200 px-5 py-2.5 text-sm text-red-600 transition hover:bg-red-50 disabled:opacity-50"
        >
          Delete household
        </button>
      </div>
    </section>
  )
}
