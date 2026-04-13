'use client'

import { useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { deleteBudget } from '@/actions/budgets'

export function DeleteBudgetButton({
  budgetId,
  householdId,
}: {
  budgetId: string
  householdId: string
}) {
  const [pending, startTransition] = useTransition()
  const router = useRouter()

  return (
    <button
      disabled={pending}
      onClick={() => {
        if (!confirm('Remove this budget cap?')) return
        startTransition(async () => {
          await deleteBudget(budgetId, householdId)
          router.refresh()
        })
      }}
      className="rounded-full border border-red-200 px-3 py-1.5 text-xs text-red-600 transition hover:bg-red-50 disabled:opacity-50"
    >
      {pending ? '…' : 'Remove'}
    </button>
  )
}
