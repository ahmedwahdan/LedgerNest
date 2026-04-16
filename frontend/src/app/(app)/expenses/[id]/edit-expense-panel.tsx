'use client'

import { useState } from 'react'
import { ExpenseForm } from '@/components/expense-form'
import { updateExpense } from '@/actions/expenses'
import type { Expense, Category } from '@/lib/definitions'
import { useRouter } from 'next/navigation'

export function EditExpensePanel({
  expense,
  categories,
}: {
  expense: Expense
  categories: Category[]
}) {
  const [open, setOpen] = useState(false)
  const router = useRouter()

  const boundUpdate = updateExpense.bind(null, expense.id)

  return (
    <section className="glass-panel rounded-[2rem] p-6">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-medium">Edit expense</h2>
        <button
          onClick={() => setOpen(!open)}
          className="rounded-full border border-[var(--line)] px-3 py-1.5 text-sm text-muted transition hover:bg-white/70"
        >
          {open ? 'Cancel' : 'Edit'}
        </button>
      </div>

      {open && (
        <div className="mt-5">
          <ExpenseForm
            action={boundUpdate}
            expense={expense}
            categories={categories}
            onSuccess={() => {
              setOpen(false)
              router.refresh()
            }}
          />
        </div>
      )}
    </section>
  )
}
