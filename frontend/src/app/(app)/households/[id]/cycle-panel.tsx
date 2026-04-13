'use client'

import { useActionState } from 'react'
import { setCycleConfig } from '@/actions/households'
import type { BudgetCycleConfig, CycleSnapshot, ActionState } from '@/lib/definitions'

interface CycleState {
  config: BudgetCycleConfig
  current_snapshot: CycleSnapshot
}

export function CyclePanel({
  householdId,
  cycle,
}: {
  householdId: string
  cycle: CycleState | null
}) {
  const boundSet = setCycleConfig.bind(null, householdId)
  const [state, action, pending] = useActionState<ActionState, FormData>(boundSet, null)

  return (
    <section className="glass-panel rounded-[2rem] p-6">
      <h2 className="mb-1 text-sm font-medium uppercase tracking-[0.2em] text-muted">
        Budget cycle
      </h2>

      {cycle ? (
        <div className="mb-5">
          <p className="text-sm">
            Current cycle:{' '}
            <span className="font-medium">{cycle.current_snapshot.label}</span>
          </p>
          <p className="mt-0.5 text-xs text-muted">
            {cycle.current_snapshot.cycle_start} – {cycle.current_snapshot.cycle_end} · starts
            on day {cycle.config.start_day} of each month
          </p>
        </div>
      ) : (
        <p className="mb-5 text-sm text-muted">No cycle configured yet.</p>
      )}

      <form action={action} className="flex items-end gap-3">
        <label className="block flex-1">
          <span className="mb-1.5 block text-sm text-muted">Cycle start day (1–28)</span>
          <input
            name="start_day"
            type="number"
            min={1}
            max={28}
            defaultValue={cycle?.config.start_day ?? 1}
            required
            className="w-full rounded-xl border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          />
        </label>
        <button
          type="submit"
          disabled={pending}
          className="rounded-full bg-[var(--accent)] px-5 py-2.5 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)] disabled:opacity-60"
        >
          {pending ? 'Saving…' : 'Update'}
        </button>
      </form>

      {state && !state.success && (
        <p className="mt-3 rounded-xl bg-red-50 px-4 py-3 text-sm text-red-700">{state.error}</p>
      )}
      {state?.success && (
        <p className="mt-3 rounded-xl bg-green-50 px-4 py-3 text-sm text-green-700">Cycle updated.</p>
      )}
    </section>
  )
}
