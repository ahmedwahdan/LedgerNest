import Link from 'next/link'
import { apiFetch } from '@/lib/api'
import type { Expense, BudgetHealth } from '@/lib/definitions'

async function getDashboardData() {
  const [expensesRes, healthRes] = await Promise.allSettled([
    apiFetch<{ expenses: Expense[] }>('/expenses?limit=5'),
    apiFetch<BudgetHealth>('/budgets/health?scope=personal'),
  ])

  const expenses = expensesRes.status === 'fulfilled' ? expensesRes.value.expenses : []
  const health = healthRes.status === 'fulfilled' ? healthRes.value : null

  return { expenses, health }
}

export default async function DashboardPage() {
  const { expenses, health } = await getDashboardData()

  return (
    <div className="shell-grid flex flex-1 flex-col overflow-auto">
      <div className="mx-auto w-full max-w-5xl space-y-6 px-5 py-8 sm:px-8">
        <header>
          <p className="text-sm uppercase tracking-[0.28em] text-muted">Overview</p>
          <h1 className="display-font mt-1 text-4xl">Dashboard</h1>
        </header>

        {/* Budget health */}
        {health ? (
          <BudgetHealthCard health={health} />
        ) : (
          <NoBudgetCard />
        )}

        {/* Recent expenses */}
        <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-medium">Recent expenses</h2>
            <Link
              href="/expenses"
              className="text-sm text-[var(--accent)] underline underline-offset-2"
            >
              View all
            </Link>
          </div>

          {expenses.length === 0 ? (
            <div className="mt-6 rounded-[1.4rem] bg-[var(--surface-strong)] px-6 py-8 text-center text-sm text-muted">
              No expenses yet.{' '}
              <Link href="/expenses" className="text-[var(--accent)] underline underline-offset-2">
                Add your first one.
              </Link>
            </div>
          ) : (
            <ul className="mt-5 space-y-2">
              {expenses.map((e) => (
                <li
                  key={e.id}
                  className="flex items-center justify-between rounded-[1.2rem] border border-[var(--line)] bg-white/65 px-4 py-3"
                >
                  <div>
                    <p className="text-sm font-medium">{e.merchant}</p>
                    <p className="mt-0.5 text-xs text-muted">{e.date}</p>
                  </div>
                  <p className="text-sm font-semibold">
                    {e.currency} {e.amount}
                  </p>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </div>
  )
}

function BudgetHealthCard({ health }: { health: BudgetHealth }) {
  const { snapshot, overall, categories } = health

  return (
    <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm uppercase tracking-[0.28em] text-muted">{snapshot.label}</p>
          <h2 className="display-font mt-1 text-3xl">Budget health</h2>
        </div>
        {overall && (
          <div
            className={`rounded-full px-3 py-1.5 text-sm ${
              overall.pct_used >= 90
                ? 'bg-red-100 text-red-700'
                : overall.pct_used >= 75
                  ? 'bg-amber-100 text-amber-700'
                  : 'bg-[rgba(15,118,110,0.12)] text-[var(--accent-strong)]'
            }`}
          >
            {Math.round(overall.pct_used)}% used
          </div>
        )}
      </div>

      {overall && (
        <div className="mt-6">
          <div className="mb-2 flex justify-between text-sm">
            <span className="text-muted">Overall cap</span>
            <span>
              {overall.spent} / {overall.amount}
            </span>
          </div>
          <div className="h-2.5 overflow-hidden rounded-full bg-[rgba(19,33,27,0.08)]">
            <div
              className={`h-2.5 rounded-full transition-all ${
                overall.pct_used >= 90
                  ? 'bg-red-500'
                  : overall.pct_used >= 75
                    ? 'bg-amber-400'
                    : 'bg-[var(--accent)]'
              }`}
              style={{ width: `${Math.min(overall.pct_used, 100)}%` }}
            />
          </div>
          <p className="mt-1.5 text-xs text-muted">
            {overall.remaining} remaining
          </p>
        </div>
      )}

      {categories.length > 0 && (
        <div className="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {categories.map((cat) => (
            <div
              key={cat.budget_id}
              className="rounded-[1.2rem] border border-[var(--line)] bg-white/65 p-4"
            >
              <p className="text-sm font-medium">{cat.category_name ?? 'Uncategorized'}</p>
              <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-[rgba(19,33,27,0.08)]">
                <div
                  className={`h-1.5 rounded-full ${
                    cat.pct_used >= 90
                      ? 'bg-red-500'
                      : cat.pct_used >= 75
                        ? 'bg-amber-400'
                        : 'bg-[var(--accent)]'
                  }`}
                  style={{ width: `${Math.min(cat.pct_used, 100)}%` }}
                />
              </div>
              <div className="mt-2 flex justify-between text-xs text-muted">
                <span>{Math.round(cat.pct_used)}% used</span>
                <span>{cat.remaining} left</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {!overall && categories.length === 0 && (
        <p className="mt-4 text-sm text-muted">
          No budgets set for this cycle yet.{' '}
          <Link href="/budgets" className="text-[var(--accent)] underline underline-offset-2">
            Set up budgets
          </Link>
        </p>
      )}
    </section>
  )
}

function NoBudgetCard() {
  return (
    <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
      <p className="text-sm uppercase tracking-[0.28em] text-muted">Budget health</p>
      <p className="mt-2 text-sm text-muted">
        Create a household to set up a budget cycle and start tracking budgets.{' '}
        <Link href="/budgets" className="text-[var(--accent)] underline underline-offset-2">
          Get started
        </Link>
      </p>
    </section>
  )
}
