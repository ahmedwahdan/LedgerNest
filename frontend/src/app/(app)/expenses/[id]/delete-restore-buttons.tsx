'use client'

import { deleteExpense, restoreExpense } from '@/actions/expenses'
import { useRouter } from 'next/navigation'
import { useTransition } from 'react'

export function DeleteRestoreButtons({
  expenseId,
  isDeleted,
}: {
  expenseId: string
  isDeleted: boolean
}) {
  const router = useRouter()
  const [pending, startTransition] = useTransition()

  if (isDeleted) {
    return (
      <form
        action={async () => {
          await restoreExpense(expenseId)
          router.refresh()
        }}
      >
        <button
          type="submit"
          disabled={pending}
          className="rounded-full border border-[var(--accent)] px-5 py-2.5 text-sm text-[var(--accent)] transition hover:bg-[rgba(15,118,110,0.08)] disabled:opacity-60"
        >
          Restore expense
        </button>
      </form>
    )
  }

  return (
    <form
      action={async () => {
        if (!confirm('Delete this expense? It can be restored later.')) return
        startTransition(async () => {
          await deleteExpense(expenseId)
        })
      }}
    >
      <button
        type="submit"
        disabled={pending}
        className="rounded-full border border-red-200 px-5 py-2.5 text-sm text-red-600 transition hover:bg-red-50 disabled:opacity-60"
      >
        {pending ? 'Deleting…' : 'Delete expense'}
      </button>
    </form>
  )
}
