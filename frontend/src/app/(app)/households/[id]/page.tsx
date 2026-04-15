import Link from 'next/link'
import { notFound } from 'next/navigation'
import { apiFetch, ApiError } from '@/lib/api'
import type { Household, CycleSnapshot, BudgetCycleConfig } from '@/lib/definitions'
import { getCurrentUser } from '@/lib/current-user'
import { MembersPanel } from './members-panel'
import { InvitePanel } from './invite-panel'
import { CyclePanel } from './cycle-panel'
import { HouseholdActionsPanel } from './household-actions-panel'

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

async function getData(id: string) {
  try {
    const [householdRes, membersRes, invitationsRes, cycleRes] = await Promise.allSettled([
      apiFetch<{ household: Household }>(`/households/${id}`),
      apiFetch<{ members: HouseholdMember[] }>(`/households/${id}/members`),
      apiFetch<{ invitations: Invitation[] }>(`/households/${id}/invitations`),
      apiFetch<CycleState>(`/households/${id}/cycle`),
    ])

    if (householdRes.status === 'rejected') {
      const err = householdRes.reason
      if (err instanceof ApiError && err.status === 404) return null
      throw err
    }

    return {
      household: householdRes.value.household,
      members: membersRes.status === 'fulfilled' ? membersRes.value.members : [],
      invitations: invitationsRes.status === 'fulfilled' ? invitationsRes.value.invitations : [],
      cycle: cycleRes.status === 'fulfilled' ? cycleRes.value : null,
    }
  } catch {
    return null
  }
}

export default async function HouseholdDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const [data, currentUser] = await Promise.all([getData(id), getCurrentUser()])
  if (!data) notFound()

  const { household, members, invitations, cycle } = data

  return (
    <div className="shell-grid flex flex-1 flex-col overflow-auto">
      <div className="mx-auto w-full max-w-3xl space-y-6 px-5 py-8 sm:px-8">
        <header className="flex items-center gap-3">
          <Link
            href="/households"
            className="rounded-full border border-[var(--line)] px-3 py-1.5 text-sm text-muted transition hover:bg-white/70"
          >
            ← Back
          </Link>
          <div>
            <p className="text-sm uppercase tracking-[0.28em] text-muted">Household</p>
            <h1 className="display-font mt-0.5 text-3xl">{household.name}</h1>
          </div>
        </header>

        {/* Members */}
        <MembersPanel householdId={id} members={members} currentUserId={currentUser?.id ?? ''} />

        {/* Pending invitations */}
        {invitations.length > 0 && (
          <InvitationsPanel householdId={id} invitations={invitations} />
        )}

        {/* Invite new member */}
        <InvitePanel householdId={id} />

        {/* Budget cycle config */}
        <CyclePanel householdId={id} cycle={cycle} />

        {/* Danger zone */}
        <HouseholdActionsPanel householdId={id} householdName={household.name} />
      </div>
    </div>
  )
}

function InvitationsPanel({
  householdId,
  invitations,
}: {
  householdId: string
  invitations: Invitation[]
}) {
  return (
    <section className="glass-panel rounded-[2rem] p-6">
      <h2 className="mb-4 text-sm font-medium uppercase tracking-[0.2em] text-muted">
        Pending invitations
      </h2>
      <ul className="divide-y divide-[var(--line)]">
        {invitations.map((inv) => (
          <li key={inv.id} className="flex items-center justify-between py-3">
            <div>
              <p className="text-sm">{inv.email}</p>
              <p className="mt-0.5 text-xs text-muted">
                {inv.role} · expires {new Date(inv.expires_at).toLocaleDateString()}
              </p>
            </div>
            <form
              action={async () => {
                'use server'
                const { revokeInvitation } = await import('@/actions/households')
                await revokeInvitation(householdId, inv.id)
              }}
            >
              <button
                type="submit"
                className="rounded-full border border-red-200 px-3 py-1.5 text-xs text-red-600 transition hover:bg-red-50"
              >
                Revoke
              </button>
            </form>
          </li>
        ))}
      </ul>
    </section>
  )
}
