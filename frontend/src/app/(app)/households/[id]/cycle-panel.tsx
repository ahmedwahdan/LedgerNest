'use client'

import { useState } from 'react'
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
  const [selectedDay, setSelectedDay] = useState(cycle?.config.start_day ?? 1)
  const calendarDays = Array.from({ length: 28 }, (_, index) => index + 1)

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

      <form action={action} className="space-y-4">
        <input type="hidden" name="start_day" value={selectedDay} />

        <div>
          <div className="mb-3 flex items-center justify-between gap-3">
            <div>
              <span className="block text-sm text-muted">Cycle start day</span>
              <p className="mt-1 text-sm">
                Starts on <span className="font-medium">day {selectedDay}</span> of each month
              </p>
            </div>
            <div className="rounded-full border border-[var(--line)] bg-white/70 px-3 py-1.5 text-xs text-muted">
              1–28 only
            </div>
          </div>

          <div className="rounded-[1.4rem] border border-[var(--line)] bg-white/70 p-4">
            <div className="mb-3 flex items-center justify-between">
              <p className="text-sm font-medium">Monthly cycle calendar</p>
              <p className="text-xs uppercase tracking-[0.18em] text-muted">Select a start day</p>
            </div>

            <div className="mb-2 grid grid-cols-7 gap-2 text-center text-[11px] uppercase tracking-[0.16em] text-muted">
              <span>Mon</span>
              <span>Tue</span>
              <span>Wed</span>
              <span>Thu</span>
              <span>Fri</span>
              <span>Sat</span>
              <span>Sun</span>
            </div>

            <div className="grid grid-cols-7 gap-2">
              {calendarDays.map((day) => {
                const isSelected = day === selectedDay

                return (
                  <button
                    key={day}
                    type="button"
                    onClick={() => setSelectedDay(day)}
                    className={`aspect-square rounded-[1rem] border text-sm font-medium transition ${
                      isSelected
                        ? 'border-[var(--accent)] bg-[var(--accent)] text-white shadow-sm'
                        : 'border-[var(--line)] bg-white/80 hover:border-[var(--accent)] hover:bg-[rgba(15,118,110,0.08)]'
                    }`}
                    aria-pressed={isSelected}
                  >
                    {day}
                  </button>
                )
              })}
            </div>
          </div>
        </div>

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
