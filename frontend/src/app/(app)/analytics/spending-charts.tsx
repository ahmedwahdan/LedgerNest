'use client'

import type { Expense, Category } from '@/lib/definitions'

interface Props {
  expenses: Expense[]
  categories: Category[]
  from: string
  to: string
}

interface CategoryTotal {
  name: string
  total: number
  count: number
}

interface MerchantTotal {
  name: string
  total: number
  count: number
}

export function SpendingCharts({ expenses, categories, from, to }: Props) {
  const categoryMap = Object.fromEntries(categories.map((c) => [c.id, c.name]))

  // Category breakdown
  const byCategory = new Map<string, CategoryTotal>()
  for (const e of expenses) {
    const key = e.category_id ?? '__none__'
    const name = e.category_id ? (categoryMap[e.category_id] ?? 'Unknown') : 'Uncategorized'
    const existing = byCategory.get(key) ?? { name, total: 0, count: 0 }
    existing.total += parseFloat(e.amount)
    existing.count++
    byCategory.set(key, existing)
  }
  const categoryTotals = Array.from(byCategory.values()).sort((a, b) => b.total - a.total)

  // Merchant breakdown (top 10)
  const byMerchant = new Map<string, MerchantTotal>()
  for (const e of expenses) {
    const existing = byMerchant.get(e.merchant) ?? { name: e.merchant, total: 0, count: 0 }
    existing.total += parseFloat(e.amount)
    existing.count++
    byMerchant.set(e.merchant, existing)
  }
  const merchantTotals = Array.from(byMerchant.values())
    .sort((a, b) => b.total - a.total)
    .slice(0, 10)

  // Grand total
  const grandTotal = expenses.reduce((sum, e) => sum + parseFloat(e.amount), 0)

  const maxCategoryTotal = categoryTotals[0]?.total ?? 1

  return (
    <div className="space-y-6">
      {/* Summary */}
      <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
        <div className="grid gap-4 sm:grid-cols-3">
          <div>
            <p className="text-xs text-muted">Total spent</p>
            <p className="mt-1 text-2xl font-semibold">${grandTotal.toFixed(2)}</p>
            <p className="mt-0.5 text-xs text-muted">{from} – {to}</p>
          </div>
          <div>
            <p className="text-xs text-muted">Transactions</p>
            <p className="mt-1 text-2xl font-semibold">{expenses.length}</p>
          </div>
          <div>
            <p className="text-xs text-muted">Avg per transaction</p>
            <p className="mt-1 text-2xl font-semibold">
              ${expenses.length > 0 ? (grandTotal / expenses.length).toFixed(2) : '0.00'}
            </p>
          </div>
        </div>
      </section>

      {/* Category breakdown */}
      <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
        <h2 className="mb-5 text-lg font-medium">By category</h2>
        {categoryTotals.length === 0 ? (
          <p className="text-sm text-muted">No expenses in this period.</p>
        ) : (
          <ul className="space-y-3">
            {categoryTotals.map((cat) => {
              const pct = grandTotal > 0 ? (cat.total / grandTotal) * 100 : 0
              return (
                <li key={cat.name}>
                  <div className="mb-1 flex justify-between text-sm">
                    <span className="font-medium">{cat.name}</span>
                    <span className="text-muted">
                      ${cat.total.toFixed(2)} · {Math.round(pct)}%
                    </span>
                  </div>
                  <div className="h-2 overflow-hidden rounded-full bg-[rgba(19,33,27,0.08)]">
                    <div
                      className="h-2 rounded-full bg-[var(--accent)] transition-all"
                      style={{ width: `${(cat.total / maxCategoryTotal) * 100}%` }}
                    />
                  </div>
                  <p className="mt-0.5 text-xs text-muted">{cat.count} transaction{cat.count !== 1 ? 's' : ''}</p>
                </li>
              )
            })}
          </ul>
        )}
      </section>

      {/* Top merchants */}
      {merchantTotals.length > 0 && (
        <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
          <h2 className="mb-5 text-lg font-medium">Top merchants</h2>
          <ul className="divide-y divide-[var(--line)]">
            {merchantTotals.map((m, i) => (
              <li key={m.name} className="flex items-center gap-4 py-3">
                <span className="w-5 text-sm text-muted">{i + 1}</span>
                <div className="flex-1">
                  <p className="text-sm font-medium">{m.name}</p>
                  <p className="text-xs text-muted">{m.count} visit{m.count !== 1 ? 's' : ''}</p>
                </div>
                <p className="text-sm font-semibold">${m.total.toFixed(2)}</p>
              </li>
            ))}
          </ul>
        </section>
      )}

      {/* CSV export */}
      <section className="glass-panel rounded-[2rem] p-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-sm font-medium">Export</h2>
            <p className="mt-0.5 text-xs text-muted">Download expenses for this period as CSV</p>
          </div>
          <CsvExportButton expenses={expenses} from={from} to={to} />
        </div>
      </section>
    </div>
  )
}

function CsvExportButton({ expenses, from, to }: { expenses: Expense[]; from: string; to: string }) {
  function download() {
    const header = 'date,merchant,amount,currency,payment_method,category_id,notes'
    const rows = expenses.map((e) =>
      [
        e.date,
        `"${e.merchant.replace(/"/g, '""')}"`,
        e.amount,
        e.currency,
        e.payment_method,
        e.category_id ?? '',
        `"${(e.notes ?? '').replace(/"/g, '""')}"`,
      ].join(','),
    )
    const csv = [header, ...rows].join('\n')
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `expenses-${from}-${to}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <button
      onClick={download}
      disabled={expenses.length === 0}
      className="rounded-full border border-[var(--line)] px-5 py-2.5 text-sm transition hover:bg-white/70 disabled:opacity-50"
    >
      Download CSV
    </button>
  )
}
