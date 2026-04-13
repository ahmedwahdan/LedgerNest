'use client'

import { useActionState } from 'react'
import type { Expense, Category, ActionState } from '@/lib/definitions'

interface ExpenseFormProps {
  action: (prev: ActionState, formData: FormData) => Promise<ActionState>
  expense?: Expense
  categories: Category[]
  onSuccess?: () => void
}

const PAYMENT_METHODS = ['card', 'cash', 'bank_transfer', 'other']

export function ExpenseForm({ action, expense, categories, onSuccess }: ExpenseFormProps) {
  const [state, formAction, pending] = useActionState<ActionState, FormData>(
    async (prev, formData) => {
      const result = await action(prev, formData)
      if (result?.success && onSuccess) onSuccess()
      return result
    },
    null,
  )

  const today = new Date().toISOString().split('T')[0]

  return (
    <form action={formAction} className="space-y-4">
      <div className="grid gap-4 sm:grid-cols-2">
        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Amount *</span>
          <input
            name="amount"
            type="text"
            inputMode="decimal"
            placeholder="0.00"
            defaultValue={expense?.amount}
            required
            className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          />
        </label>

        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Currency</span>
          <input
            name="currency"
            type="text"
            placeholder="USD"
            defaultValue={expense?.currency ?? 'USD'}
            maxLength={3}
            className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm uppercase outline-none transition focus:border-[var(--accent)]"
          />
        </label>
      </div>

      <label className="block">
        <span className="mb-1.5 block text-sm text-muted">Merchant *</span>
        <input
          name="merchant"
          type="text"
          placeholder="Where did you spend?"
          defaultValue={expense?.merchant}
          required
          className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
        />
      </label>

      <div className="grid gap-4 sm:grid-cols-2">
        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Date *</span>
          <input
            name="date"
            type="date"
            defaultValue={expense?.date ?? today}
            required
            className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          />
        </label>

        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Payment method *</span>
          <select
            name="payment_method"
            defaultValue={expense?.payment_method ?? 'card'}
            required
            className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          >
            {PAYMENT_METHODS.map((m) => (
              <option key={m} value={m}>
                {m.replace('_', ' ')}
              </option>
            ))}
          </select>
        </label>
      </div>

      {categories.length > 0 && (
        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Category</span>
          <select
            name="category_id"
            defaultValue={expense?.category_id ?? ''}
            className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          >
            <option value="">No category</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </label>
      )}

      <label className="block">
        <span className="mb-1.5 block text-sm text-muted">Notes</span>
        <textarea
          name="notes"
          placeholder="Optional notes"
          defaultValue={expense?.notes ?? ''}
          rows={2}
          className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
        />
      </label>

      {state && !state.success && (
        <p className="rounded-[0.9rem] bg-red-50 px-4 py-3 text-sm text-red-700">{state.error}</p>
      )}

      <button
        type="submit"
        disabled={pending}
        className="w-full rounded-full bg-[var(--accent)] px-6 py-3 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)] disabled:opacity-60"
      >
        {pending ? 'Saving…' : expense ? 'Save changes' : 'Add expense'}
      </button>
    </form>
  )
}
