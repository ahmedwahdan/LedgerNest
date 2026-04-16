'use client'

import type { SpendingSummary, SpendingByCategory, TopMerchant } from '@/lib/definitions'

interface Props {
  summary: SpendingSummary
  byCategory: SpendingByCategory[]
  merchants: TopMerchant[]
  from: string
  to: string
}

export function SpendingCharts({ summary, byCategory, merchants, from, to }: Props) {
  const grandTotal = parseFloat(summary.total)
  const maxCategoryTotal = byCategory.length > 0 ? parseFloat(byCategory[0].total) : 1

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
            <p className="mt-1 text-2xl font-semibold">{summary.count}</p>
          </div>
          <div>
            <p className="text-xs text-muted">Avg per transaction</p>
            <p className="mt-1 text-2xl font-semibold">
              ${parseFloat(summary.average).toFixed(2)}
            </p>
          </div>
        </div>
      </section>

      {/* Category breakdown */}
      <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
        <h2 className="mb-5 text-lg font-medium">By category</h2>
        {byCategory.length === 0 ? (
          <p className="text-sm text-muted">No expenses in this period.</p>
        ) : (
          <ul className="space-y-3">
            {byCategory.map((cat) => {
              const catTotal = parseFloat(cat.total)
              return (
                <li key={cat.category_name}>
                  <div className="mb-1 flex justify-between text-sm">
                    <span className="font-medium">{cat.category_name}</span>
                    <span className="text-muted">
                      ${catTotal.toFixed(2)} · {Math.round(cat.pct_of_total)}%
                    </span>
                  </div>
                  <div className="h-2 overflow-hidden rounded-full bg-[rgba(19,33,27,0.08)]">
                    <div
                      className="h-2 rounded-full bg-[var(--accent)] transition-all"
                      style={{ width: `${(catTotal / maxCategoryTotal) * 100}%` }}
                    />
                  </div>
                  <p className="mt-0.5 text-xs text-muted">
                    {cat.count} transaction{cat.count !== 1 ? 's' : ''}
                  </p>
                </li>
              )
            })}
          </ul>
        )}
      </section>

      {/* Top merchants */}
      {merchants.length > 0 && (
        <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
          <h2 className="mb-5 text-lg font-medium">Top merchants</h2>
          <ul className="divide-y divide-[var(--line)]">
            {merchants.map((m, i) => (
              <li key={m.merchant} className="flex items-center gap-4 py-3">
                <span className="w-5 text-sm text-muted">{i + 1}</span>
                <div className="flex-1">
                  <p className="text-sm font-medium">{m.merchant}</p>
                  <p className="text-xs text-muted">
                    {m.count} visit{m.count !== 1 ? 's' : ''}
                  </p>
                </div>
                <p className="text-sm font-semibold">${parseFloat(m.total).toFixed(2)}</p>
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
          <CsvExportButton from={from} to={to} disabled={summary.count === 0} />
        </div>
      </section>
    </div>
  )
}

function CsvExportButton({
  from,
  to,
  disabled,
}: {
  from: string
  to: string
  disabled: boolean
}) {
  async function download() {
    const qs = new URLSearchParams({ from, to, limit: '500' })
    const res = await fetch(`/api/expenses?${qs}`)
    if (!res.ok) return
    const { expenses } = await res.json()

    const header = 'date,merchant,amount,currency,payment_method,category_id,notes'
    const rows = expenses.map(
      (e: {
        date: string
        merchant: string
        amount: string
        currency: string
        payment_method: string
        category_id?: string
        notes?: string
      }) =>
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
      disabled={disabled}
      className="rounded-full border border-[var(--line)] px-5 py-2.5 text-sm transition hover:bg-white/70 disabled:opacity-50"
    >
      Download CSV
    </button>
  )
}
