'use client'

import { useState } from 'react'
import { useActionState } from 'react'
import { createBudget } from '@/actions/budgets'
import type { Category, ActionState } from '@/lib/definitions'

export function AddBudgetPanel({
  categories,
  snapshotId,
}: {
  categories: Category[]
  snapshotId?: string
}) {
  const [open, setOpen] = useState(false)
  const [state, action, pending] = useActionState<ActionState, FormData>(
    async (prev, formData) => {
      const result = await createBudget(prev, formData)
      if (result?.success) setOpen(false)
      return result
    },
    null,
  )

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="rounded-full bg-[var(--accent)] px-4 py-2 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)]"
      >
        + Add budget
      </button>

      {open && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4 backdrop-blur-sm"
          onClick={(e) => e.target === e.currentTarget && setOpen(false)}
        >
          <div className="glass-panel w-full max-w-sm rounded-[2rem] p-7">
            <div className="mb-5 flex items-center justify-between">
              <h2 className="display-font text-2xl">New budget</h2>
              <button
                onClick={() => setOpen(false)}
                className="rounded-full border border-[var(--line)] px-3 py-1.5 text-sm text-muted transition hover:bg-white/70"
              >
                Cancel
              </button>
            </div>

            <form action={action} className="space-y-4">
              {snapshotId && (
                <input type="hidden" name="snapshot_id" value={snapshotId} />
              )}

              <label className="block">
                <span className="mb-1.5 block text-sm text-muted">Amount *</span>
                <input
                  name="amount"
                  type="text"
                  inputMode="decimal"
                  placeholder="0.00"
                  required
                  className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
                />
              </label>

              <label className="block">
                <span className="mb-1.5 block text-sm text-muted">Category (optional)</span>
                <select
                  name="category_id"
                  defaultValue=""
                  className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
                >
                  <option value="">Overall cap (no category)</option>
                  {categories.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </label>

              {state && !state.success && (
                <p className="rounded-[0.9rem] bg-red-50 px-4 py-3 text-sm text-red-700">
                  {state.error}
                </p>
              )}

              <button
                type="submit"
                disabled={pending}
                className="w-full rounded-full bg-[var(--accent)] px-6 py-3 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)] disabled:opacity-60"
              >
                {pending ? 'Saving…' : 'Create budget'}
              </button>
            </form>
          </div>
        </div>
      )}
    </>
  )
}
