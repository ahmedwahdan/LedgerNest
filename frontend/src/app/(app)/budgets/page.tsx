import { apiFetch } from '@/lib/api'
import { getActiveHousehold } from '@/lib/household-context'
import Link from 'next/link'
import type { Budget, BudgetHealth, Category } from '@/lib/definitions'
import { DeleteBudgetButton } from './delete-budget-button'
import { AddBudgetPanel } from './add-budget-panel'

async function getData() {
  const activeHousehold = await getActiveHousehold()
  if (!activeHousehold) {
    return {
      activeHousehold: null,
      health: null,
      budgets: [],
      categories: [],
    }
  }

  const [healthRes, budgetsRes, categoriesRes] = await Promise.allSettled([
    apiFetch<BudgetHealth>(
      `/budgets/health?household_id=${encodeURIComponent(activeHousehold.id)}`,
    ),
    apiFetch<{ budgets: Budget[] }>(
      `/budgets?household_id=${encodeURIComponent(activeHousehold.id)}`,
    ),
    apiFetch<{ categories: Category[] }>(
      `/categories?household_id=${encodeURIComponent(activeHousehold.id)}`,
    ),
  ])

  return {
    activeHousehold,
    health: healthRes.status === 'fulfilled' ? healthRes.value : null,
    budgets: budgetsRes.status === 'fulfilled' ? budgetsRes.value.budgets : [],
    categories: categoriesRes.status === 'fulfilled' ? categoriesRes.value.categories : [],
  }
}

export default async function BudgetsPage() {
  const { activeHousehold, health, budgets, categories } = await getData()

  return (
    <div className="shell-grid flex flex-1 flex-col overflow-auto">
      <div className="mx-auto w-full max-w-3xl space-y-6 px-5 py-8 sm:px-8">
        <header>
          <p className="text-sm uppercase tracking-[0.28em] text-muted">Household</p>
          <h1 className="display-font mt-1 text-4xl">Budgets</h1>
          <p className="mt-2 text-sm text-muted">
            {activeHousehold
              ? `Tracking the current cycle for ${activeHousehold.name}.`
              : 'Create or select a household before setting budget caps.'}
          </p>
        </header>

        {!activeHousehold && <NoHouseholdState />}

        {/* Health summary */}
        {activeHousehold && health && (
          <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-muted">{health.snapshot.label}</p>
                <h2 className="display-font mt-0.5 text-2xl">This cycle</h2>
              </div>
              <span className="text-sm text-muted">
                {health.snapshot.cycle_start} – {health.snapshot.cycle_end}
              </span>
            </div>

            {health.overall ? (
              <div className="mt-5">
                <div className="mb-1.5 flex justify-between text-sm">
                  <span className="text-muted">Overall cap</span>
                  <span className="font-medium">
                    {health.overall.spent} / {health.overall.amount}
                  </span>
                </div>
                <div className="h-2.5 overflow-hidden rounded-full bg-[rgba(19,33,27,0.08)]">
                  <div
                    className={`h-2.5 rounded-full ${
                      health.overall.pct_used >= 90
                        ? 'bg-red-500'
                        : health.overall.pct_used >= 75
                          ? 'bg-amber-400'
                          : 'bg-[var(--accent)]'
                    }`}
                    style={{ width: `${Math.min(health.overall.pct_used, 100)}%` }}
                  />
                </div>
                <p className="mt-1 text-xs text-muted">
                  {health.overall.remaining} remaining · {Math.round(health.overall.pct_used)}% used
                </p>
              </div>
            ) : (
              <p className="mt-3 text-sm text-muted">No overall budget cap set for this cycle.</p>
            )}

            {health.categories.length > 0 && (
              <div className="mt-5 space-y-3">
                {health.categories.map((cat) => (
                  <div key={cat.budget_id}>
                    <div className="mb-1 flex justify-between text-sm">
                      <span>{cat.category_name ?? 'Uncategorized'}</span>
                      <span className="text-muted">
                        {cat.spent} / {cat.amount}
                      </span>
                    </div>
                    <div className="h-1.5 overflow-hidden rounded-full bg-[rgba(19,33,27,0.08)]">
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
                  </div>
                ))}
              </div>
            )}
          </section>
        )}

        {/* Budget list */}
        <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-medium">Budget caps</h2>
            <AddBudgetPanel
              categories={categories}
              snapshotId={health?.snapshot.id}
              householdId={activeHousehold?.id}
            />
          </div>

          {activeHousehold && budgets.length === 0 ? (
            <p className="mt-4 text-sm text-muted">
              No budgets set yet. Add an overall cap or per-category limit above.
            </p>
          ) : activeHousehold ? (
            <ul className="mt-5 divide-y divide-[var(--line)]">
              {budgets.map((b) => (
                <li key={b.id} className="flex items-center justify-between py-3">
                  <div>
                    <p className="text-sm font-medium">
                      {b.category_id
                        ? categories.find((c) => c.id === b.category_id)?.name ?? 'Category'
                        : 'Overall cap'}
                    </p>
                    <p className="text-xs text-muted">
                      {b.scope} · {b.amount}
                      {parseFloat(b.rollover_amount) > 0 && ` + ${b.rollover_amount} rollover`}
                    </p>
                  </div>
                  <DeleteBudgetButton budgetId={b.id} householdId={activeHousehold.id} />
                </li>
              ))}
            </ul>
          ) : (
            <p className="mt-4 text-sm text-muted">
              Budget caps appear here after you choose a household.
            </p>
          )}
        </section>
      </div>
    </div>
  )
}

function NoHouseholdState() {
  return (
    <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
      <p className="text-sm uppercase tracking-[0.28em] text-muted">Setup</p>
      <h2 className="display-font mt-1 text-3xl">No active household</h2>
      <p className="mt-3 max-w-2xl text-sm text-muted">
        Budgets are household-based in LedgerNest. Create a household first, then set a cycle and
        category caps for the people sharing that plan.
      </p>
      <div className="mt-5 flex flex-wrap gap-3">
        <Link
          href="/households"
          className="rounded-full bg-[var(--accent)] px-4 py-2 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)]"
        >
          Go to households
        </Link>
      </div>
    </section>
  )
}
