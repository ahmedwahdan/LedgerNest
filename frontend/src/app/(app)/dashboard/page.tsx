import Link from 'next/link'
import { apiFetch } from '@/lib/api'
import { getActiveHousehold } from '@/lib/household-context'
import type {
  BudgetCycleConfig,
  BudgetHealth,
  CycleSnapshot,
  Expense,
  Household,
} from '@/lib/definitions'
import { getCurrentUser } from '@/lib/current-user'

interface HouseholdMember {
  id: string
  household_id: string
  user_id: string
  display_name: string
  email: string
  role: string
  joined_at: string
}

interface Invitation {
  id: string
  household_id: string
  email: string
  role: string
  status: string
  expires_at: string
  created_at: string
}

interface CycleState {
  config: BudgetCycleConfig
  current_snapshot: CycleSnapshot
}

async function getDashboardData() {
  const activeHousehold = await getActiveHousehold()

  const [expensesRes, healthRes, householdRes, membersRes, invitationsRes, cycleRes] =
    await Promise.allSettled([
    apiFetch<{ expenses: Expense[] }>('/expenses?limit=5'),
    activeHousehold
      ? apiFetch<BudgetHealth>(
          `/budgets/health?household_id=${encodeURIComponent(activeHousehold.id)}`,
        )
      : Promise.resolve(null),
    activeHousehold
      ? apiFetch<{ household: Household }>(`/households/${activeHousehold.id}`)
      : Promise.resolve(null),
    activeHousehold
      ? apiFetch<{ members: HouseholdMember[] }>(`/households/${activeHousehold.id}/members`)
      : Promise.resolve(null),
    activeHousehold
      ? apiFetch<{ invitations: Invitation[] }>(`/households/${activeHousehold.id}/invitations`)
      : Promise.resolve(null),
    activeHousehold
      ? apiFetch<CycleState>(`/households/${activeHousehold.id}/cycle`)
      : Promise.resolve(null),
    ])

  const expenses = expensesRes.status === 'fulfilled' ? expensesRes.value.expenses : []
  const health =
    healthRes.status === 'fulfilled' && healthRes.value ? healthRes.value : null
  const household =
    householdRes.status === 'fulfilled' && householdRes.value ? householdRes.value.household : null
  const members =
    membersRes.status === 'fulfilled' && membersRes.value ? membersRes.value.members : []
  const invitations =
    invitationsRes.status === 'fulfilled' && invitationsRes.value
      ? invitationsRes.value.invitations
      : []
  const cycle =
    cycleRes.status === 'fulfilled' && cycleRes.value ? cycleRes.value : null

  return { expenses, health, household, members, invitations, cycle }
}

export default async function DashboardPage() {
  const user = await getCurrentUser()
  const { expenses, health, household, members, invitations, cycle } = await getDashboardData()

  return (
    <div className="shell-grid flex flex-1 flex-col overflow-auto">
      <div className="mx-auto w-full max-w-5xl space-y-6 px-5 py-8 sm:px-8">
        <header>
          <p className="text-sm uppercase tracking-[0.28em] text-muted">Overview</p>
          <h1 className="display-font mt-1 text-4xl">
            {user ? `Welcome back, ${user.display_name}` : 'Dashboard'}
          </h1>
          <p className="mt-2 max-w-2xl text-sm text-muted">
            Track the latest budget movement, keep an eye on your recent expenses,
            and move quickly into the parts of LedgerNest that still need setup.
          </p>
        </header>

        {household ? (
          <HouseholdOverviewCard
            household={household}
            members={members}
            invitations={invitations}
            cycle={cycle}
          />
        ) : (
          <NoHouseholdCard />
        )}

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

function HouseholdOverviewCard({
  household,
  members,
  invitations,
  cycle,
}: {
  household: Household
  members: HouseholdMember[]
  invitations: Invitation[]
  cycle: CycleState | null
}) {
  const ownerCount = members.filter((member) => member.role === 'owner').length
  const editorCount = members.filter((member) => member.role === 'editor').length

  return (
    <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p className="text-sm uppercase tracking-[0.28em] text-muted">Active household</p>
          <h2 className="display-font mt-1 text-3xl">{household.name}</h2>
          <p className="mt-2 max-w-2xl text-sm text-muted">
            The dashboard is currently scoped to this household. Switch households from the header
            to compare members, invites, and budget progress.
          </p>
        </div>
        <Link
          href={`/households/${household.id}`}
          className="rounded-full border border-[var(--line)] px-4 py-2 text-sm transition hover:bg-white/70"
        >
          Open household
        </Link>
      </div>

      <div className="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Members" value={String(members.length)} detail={`${ownerCount} owners`} />
        <StatCard label="Editors" value={String(editorCount)} detail="Can contribute expenses" />
        <StatCard
          label="Pending invites"
          value={String(invitations.length)}
          detail={invitations.length > 0 ? 'Waiting on responses' : 'No pending invites'}
        />
        <StatCard
          label="Budget cycle"
          value={cycle?.current_snapshot.label ?? 'Not set'}
          detail={
            cycle
              ? `${cycle.current_snapshot.cycle_start} to ${cycle.current_snapshot.cycle_end}`
              : 'Configure your cycle'
          }
        />
      </div>
    </section>
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

function StatCard({
  label,
  value,
  detail,
}: {
  label: string
  value: string
  detail: string
}) {
  return (
    <div className="rounded-[1.2rem] border border-[var(--line)] bg-white/65 p-4">
      <p className="text-xs uppercase tracking-[0.18em] text-muted">{label}</p>
      <p className="mt-2 text-lg font-semibold">{value}</p>
      <p className="mt-1 text-xs text-muted">{detail}</p>
    </div>
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

function NoHouseholdCard() {
  return (
    <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
      <p className="text-sm uppercase tracking-[0.28em] text-muted">Household</p>
      <h2 className="display-font mt-1 text-3xl">Set up your shared workspace</h2>
      <p className="mt-3 max-w-2xl text-sm text-muted">
        Create a household to unlock shared budgets, invitations, and a meaningful dashboard. Once
        one is active, this page will show its members, cycle, and budget health.
      </p>
      <div className="mt-5">
        <Link
          href="/households"
          className="rounded-full bg-[var(--accent)] px-4 py-2 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)]"
        >
          Create a household
        </Link>
      </div>
    </section>
  )
}
