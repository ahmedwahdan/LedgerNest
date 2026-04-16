'use client'

import { useActionState } from 'react'
import type { ActionState, Category } from '@/lib/definitions'

interface CategoryFormProps {
  householdId: string
  categories: Category[]
  action: (prev: ActionState, formData: FormData) => Promise<ActionState>
  category?: Category
  onSuccess?: () => void
}

export function CategoryForm({
  householdId,
  categories,
  action,
  category,
  onSuccess,
}: CategoryFormProps) {
  const [state, formAction, pending] = useActionState<ActionState, FormData>(
    async (prev, formData) => {
      const result = await action(prev, formData)
      if (result?.success && onSuccess) onSuccess()
      return result
    },
    null,
  )

  return (
    <form action={formAction} className="space-y-4">
      <input type="hidden" name="household_id" value={householdId} />

      <label className="block">
        <span className="mb-1.5 block text-sm text-muted">Name *</span>
        <input
          name="name"
          type="text"
          defaultValue={category?.name ?? ''}
          required
          className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
        />
      </label>

      <div className="grid gap-4 sm:grid-cols-2">
        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Parent category</span>
          <select
            name="parent_id"
            defaultValue={category?.parent_id ?? ''}
            className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          >
            <option value="">No parent</option>
            {categories
              .filter((item) => item.id !== category?.id)
              .map((item) => (
                <option key={item.id} value={item.id}>
                  {item.name}
                </option>
              ))}
          </select>
        </label>

        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Color</span>
          <input
            name="color"
            type="text"
            defaultValue={category?.color ?? ''}
            placeholder="#0f766e"
            className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          />
        </label>
      </div>

      <label className="block">
        <span className="mb-1.5 block text-sm text-muted">Icon</span>
        <input
          name="icon"
          type="text"
          defaultValue={category?.icon ?? ''}
          placeholder="e.g. cart, home, coffee"
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
        {pending ? 'Saving…' : category ? 'Save category' : 'Create category'}
      </button>
    </form>
  )
}
