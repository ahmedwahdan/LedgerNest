import Link from 'next/link'
import { notFound } from 'next/navigation'
import { apiFetch, ApiError } from '@/lib/api'
import type { Expense, Category, AuditLogEntry } from '@/lib/definitions'
import { EditExpensePanel } from './edit-expense-panel'
import { DeleteRestoreButtons } from './delete-restore-buttons'

interface PageProps {
  params: Promise<{ id: string }>
}

async function getData(id: string) {
  try {
    const [expenseRes, categoriesRes, historyRes] = await Promise.allSettled([
      apiFetch<{ expense: Expense }>(`/expenses/${id}`),
      apiFetch<{ categories: Category[] }>('/categories'),
      apiFetch<{ history: AuditLogEntry[] }>(`/expenses/${id}/history`),
    ])

    if (expenseRes.status === 'rejected') {
      const err = expenseRes.reason
      if (err instanceof ApiError && err.status === 404) return null
      throw err
    }

    return {
      expense: expenseRes.value.expense,
      categories: categoriesRes.status === 'fulfilled' ? categoriesRes.value.categories : [],
      history: historyRes.status === 'fulfilled' ? historyRes.value.history : [],
    }
  } catch {
    return null
  }
}

export default async function ExpenseDetailPage({ params }: PageProps) {
  const { id } = await params
  const data = await getData(id)

  if (!data) notFound()

  const { expense, categories, history } = data

  return (
    <div className="shell-grid flex flex-1 flex-col overflow-auto">
      <div className="mx-auto w-full max-w-2xl space-y-6 px-5 py-8 sm:px-8">
        <header className="flex items-center gap-3">
          <Link
            href="/expenses"
            className="rounded-full border border-[var(--line)] px-3 py-1.5 text-sm text-muted transition hover:bg-white/70"
          >
            ← Back
          </Link>
          <div>
            <p className="text-sm uppercase tracking-[0.28em] text-muted">Expense</p>
            <h1 className="display-font mt-0.5 text-3xl">{expense.merchant}</h1>
          </div>
        </header>

        {/* Summary */}
        <section className="glass-panel rounded-[2rem] p-6">
          <div className="grid gap-4 sm:grid-cols-3">
            <div>
              <p className="text-xs text-muted">Amount</p>
              <p className="mt-1 text-xl font-semibold">
                {expense.currency} {expense.amount}
              </p>
            </div>
            <div>
              <p className="text-xs text-muted">Date</p>
              <p className="mt-1 text-sm">{expense.date}</p>
            </div>
            <div>
              <p className="text-xs text-muted">Payment</p>
              <p className="mt-1 text-sm">{expense.payment_method.replace('_', ' ')}</p>
            </div>
          </div>
          {expense.notes && (
            <p className="mt-4 text-sm text-muted">{expense.notes}</p>
          )}
          {expense.is_deleted && (
            <p className="mt-4 rounded-xl bg-red-50 px-3 py-2 text-sm text-red-600">
              This expense has been deleted.
            </p>
          )}
        </section>

        {/* Edit form */}
        {!expense.is_deleted && (
          <EditExpensePanel expense={expense} categories={categories} />
        )}

        {/* Delete / Restore */}
        <DeleteRestoreButtons expenseId={expense.id} isDeleted={expense.is_deleted ?? false} />

        {/* Audit history */}
        {history.length > 0 && (
          <section className="glass-panel rounded-[2rem] p-6">
            <h2 className="mb-4 text-sm font-medium">History</h2>
            <ol className="space-y-3">
              {history.map((entry) => (
                <li key={entry.id} className="flex items-start gap-3 text-sm">
                  <span
                    className={`mt-0.5 rounded-full px-2 py-0.5 text-xs ${
                      entry.action === 'create'
                        ? 'bg-green-100 text-green-700'
                        : entry.action === 'delete'
                          ? 'bg-red-100 text-red-700'
                          : entry.action === 'restore'
                            ? 'bg-blue-100 text-blue-700'
                            : 'bg-amber-100 text-amber-700'
                    }`}
                  >
                    {entry.action}
                  </span>
                  <span className="text-muted">
                    {new Date(entry.created_at).toLocaleString()}
                  </span>
                </li>
              ))}
            </ol>
          </section>
        )}
      </div>
    </div>
  )
}
