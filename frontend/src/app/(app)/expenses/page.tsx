import Link from 'next/link'
import { apiFetch } from '@/lib/api'
import type { Expense, Category } from '@/lib/definitions'
import { AddExpenseButton } from './add-expense-button'

interface PageProps {
  searchParams: Promise<{ from?: string; to?: string; merchant?: string }>
}

async function getData(params: { from?: string; to?: string; merchant?: string }) {
  const qs = new URLSearchParams()
  if (params.from) qs.set('from', params.from)
  if (params.to) qs.set('to', params.to)
  if (params.merchant) qs.set('merchant', params.merchant)
  qs.set('limit', '50')

  const [expensesRes, categoriesRes] = await Promise.allSettled([
    apiFetch<{ expenses: Expense[] }>(`/expenses?${qs}`),
    apiFetch<{ categories: Category[] }>('/categories'),
  ])

  return {
    expenses: expensesRes.status === 'fulfilled' ? expensesRes.value.expenses : [],
    categories: categoriesRes.status === 'fulfilled' ? categoriesRes.value.categories : [],
  }
}

export default async function ExpensesPage({ searchParams }: PageProps) {
  const params = await searchParams
  const { expenses, categories } = await getData(params)

  return (
    <div className="shell-grid flex flex-1 flex-col overflow-auto">
      <div className="mx-auto w-full max-w-4xl space-y-6 px-5 py-8 sm:px-8">
        <header className="flex items-center justify-between">
          <div>
            <p className="text-sm uppercase tracking-[0.28em] text-muted">Personal</p>
            <h1 className="display-font mt-1 text-4xl">Expenses</h1>
          </div>
          <AddExpenseButton categories={categories} />
        </header>

        {/* Filters */}
        <FilterBar params={params} />

        {/* Expense list */}
        {expenses.length === 0 ? (
          <div className="glass-panel rounded-[2rem] px-6 py-12 text-center">
            <p className="text-sm text-muted">No expenses found.</p>
          </div>
        ) : (
          <section className="glass-panel rounded-[2rem] divide-y divide-[var(--line)] overflow-hidden">
            {expenses.map((expense) => (
              <Link
                key={expense.id}
                href={`/expenses/${expense.id}`}
                className="flex items-center justify-between px-5 py-4 transition hover:bg-white/50"
              >
                <div>
                  <p className="text-sm font-medium">{expense.merchant}</p>
                  <p className="mt-0.5 text-xs text-muted">
                    {expense.date}
                    {expense.payment_method && ` · ${expense.payment_method.replace('_', ' ')}`}
                  </p>
                </div>
                <p className="text-sm font-semibold">
                  {expense.currency} {expense.amount}
                </p>
              </Link>
            ))}
          </section>
        )}
      </div>
    </div>
  )
}

function FilterBar({ params }: { params: { from?: string; to?: string; merchant?: string } }) {
  return (
    <form className="flex flex-wrap gap-3">
      <input
        name="from"
        type="date"
        defaultValue={params.from ?? ''}
        placeholder="From"
        className="rounded-xl border border-[var(--line)] bg-white/80 px-3 py-2 text-sm outline-none focus:border-[var(--accent)]"
      />
      <input
        name="to"
        type="date"
        defaultValue={params.to ?? ''}
        placeholder="To"
        className="rounded-xl border border-[var(--line)] bg-white/80 px-3 py-2 text-sm outline-none focus:border-[var(--accent)]"
      />
      <input
        name="merchant"
        type="text"
        defaultValue={params.merchant ?? ''}
        placeholder="Merchant"
        className="rounded-xl border border-[var(--line)] bg-white/80 px-3 py-2 text-sm outline-none focus:border-[var(--accent)]"
      />
      <button
        type="submit"
        className="rounded-xl border border-[var(--line)] bg-white/80 px-4 py-2 text-sm transition hover:bg-white"
      >
        Filter
      </button>
      {(params.from || params.to || params.merchant) && (
        <a
          href="/expenses"
          className="rounded-xl border border-[var(--line)] bg-white/80 px-4 py-2 text-sm text-muted transition hover:bg-white"
        >
          Clear
        </a>
      )}
    </form>
  )
}
