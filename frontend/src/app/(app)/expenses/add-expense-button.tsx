'use client'

import { useState } from 'react'
import { ExpenseForm } from '@/components/expense-form'
import { createExpense } from '@/actions/expenses'
import type { Category } from '@/lib/definitions'

export function AddExpenseButton({ categories }: { categories: Category[] }) {
  const [open, setOpen] = useState(false)

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="rounded-full bg-[var(--accent)] px-5 py-2.5 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)]"
      >
        + Add expense
      </button>

      {open && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4 backdrop-blur-sm"
          onClick={(e) => e.target === e.currentTarget && setOpen(false)}
        >
          <div className="glass-panel w-full max-w-md rounded-[2rem] p-7">
            <div className="mb-5 flex items-center justify-between">
              <h2 className="display-font text-2xl">New expense</h2>
              <button
                onClick={() => setOpen(false)}
                className="rounded-full border border-[var(--line)] px-3 py-1.5 text-sm text-muted transition hover:bg-white/70"
              >
                Cancel
              </button>
            </div>
            <ExpenseForm
              action={createExpense}
              categories={categories}
              onSuccess={() => setOpen(false)}
            />
          </div>
        </div>
      )}
    </>
  )
}
