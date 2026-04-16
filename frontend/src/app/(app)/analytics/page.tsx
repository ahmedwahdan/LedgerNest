import { apiFetch } from '@/lib/api'
import { getActiveHousehold } from '@/lib/household-context'
import type {
  SpendingSummary,
  SpendingByCategory,
  TopMerchant,
} from '@/lib/definitions'
import { SpendingCharts } from './spending-charts'

interface PageProps {
  searchParams: Promise<{ from?: string; to?: string }>
}

function currentMonthBounds() {
  const now = new Date()
  const from = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`
  const last = new Date(now.getFullYear(), now.getMonth() + 1, 0)
  const to = `${last.getFullYear()}-${String(last.getMonth() + 1).padStart(2, '0')}-${String(last.getDate()).padStart(2, '0')}`
  return { from, to }
}

async function getAnalytics(from: string, to: string) {
  const activeHousehold = await getActiveHousehold()
  const scope = activeHousehold ? 'household' : 'personal'
  const base: Record<string, string> = { from, to, scope }
  if (activeHousehold) base.household_id = activeHousehold.id

  const spendingQs = new URLSearchParams(base)
  const merchantsQs = new URLSearchParams({ ...base, limit: '10' })

  const [spendingRes, merchantsRes] = await Promise.allSettled([
    apiFetch<{ summary: SpendingSummary; by_category: SpendingByCategory[] }>(
      `/analytics/spending?${spendingQs}`,
    ),
    apiFetch<{ merchants: TopMerchant[] }>(`/analytics/merchants?${merchantsQs}`),
  ])

  return {
    summary:
      spendingRes.status === 'fulfilled'
        ? spendingRes.value.summary
        : { total: '0', count: 0, average: '0' },
    byCategory:
      spendingRes.status === 'fulfilled' ? spendingRes.value.by_category : [],
    merchants:
      merchantsRes.status === 'fulfilled' ? merchantsRes.value.merchants : [],
  }
}

export default async function AnalyticsPage({ searchParams }: PageProps) {
  const params = await searchParams
  const defaultBounds = currentMonthBounds()
  const from = params.from ?? defaultBounds.from
  const to = params.to ?? defaultBounds.to

  const { summary, byCategory, merchants } = await getAnalytics(from, to)

  return (
    <div className="shell-grid flex flex-1 flex-col overflow-auto">
      <div className="mx-auto w-full max-w-4xl space-y-6 px-5 py-8 sm:px-8">
        <header className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="text-sm uppercase tracking-[0.28em] text-muted">Overview</p>
            <h1 className="display-font mt-1 text-4xl">Analytics</h1>
          </div>

          <form className="flex flex-wrap gap-2">
            <input
              name="from"
              type="date"
              defaultValue={from}
              className="rounded-xl border border-[var(--line)] bg-white/80 px-3 py-2 text-sm outline-none focus:border-[var(--accent)]"
            />
            <input
              name="to"
              type="date"
              defaultValue={to}
              className="rounded-xl border border-[var(--line)] bg-white/80 px-3 py-2 text-sm outline-none focus:border-[var(--accent)]"
            />
            <button
              type="submit"
              className="rounded-xl border border-[var(--line)] bg-white/80 px-4 py-2 text-sm transition hover:bg-white"
            >
              Apply
            </button>
          </form>
        </header>

        <SpendingCharts
          summary={summary}
          byCategory={byCategory}
          merchants={merchants}
          from={from}
          to={to}
        />
      </div>
    </div>
  )
}
