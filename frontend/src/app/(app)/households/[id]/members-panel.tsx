'use client'

import { useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { updateMemberRole, removeMember } from '@/actions/households'

interface Member {
  id: string
  user_id: string
  display_name: string
  email: string
  role: string
  joined_at: string
}

const ROLES = ['owner', 'editor', 'viewer']

export function MembersPanel({
  householdId,
  members,
}: {
  householdId: string
  members: Member[]
}) {
  return (
    <section className="glass-panel rounded-[2rem] p-6">
      <h2 className="mb-4 text-sm font-medium uppercase tracking-[0.2em] text-muted">
        Members ({members.length})
      </h2>
      <ul className="divide-y divide-[var(--line)]">
        {members.map((m) => (
          <MemberRow key={m.id} householdId={householdId} member={m} />
        ))}
      </ul>
    </section>
  )
}

function MemberRow({ householdId, member }: { householdId: string; member: Member }) {
  const router = useRouter()
  const [pending, startTransition] = useTransition()

  return (
    <li className="flex items-center justify-between gap-4 py-3">
      <div>
        <p className="text-sm font-medium">{member.display_name}</p>
        <p className="mt-0.5 text-xs text-muted">{member.email}</p>
      </div>

      <div className="flex items-center gap-2">
        <select
          defaultValue={member.role}
          disabled={pending}
          onChange={(e) => {
            startTransition(async () => {
              await updateMemberRole(householdId, member.user_id, e.target.value)
              router.refresh()
            })
          }}
          className="rounded-xl border border-[var(--line)] bg-white/80 px-2.5 py-1.5 text-xs outline-none focus:border-[var(--accent)] disabled:opacity-50"
        >
          {ROLES.map((r) => (
            <option key={r} value={r}>
              {r}
            </option>
          ))}
        </select>

        <button
          disabled={pending}
          onClick={() => {
            if (!confirm(`Remove ${member.display_name} from this household?`)) return
            startTransition(async () => {
              await removeMember(householdId, member.user_id)
              router.refresh()
            })
          }}
          className="rounded-full border border-red-200 px-3 py-1.5 text-xs text-red-600 transition hover:bg-red-50 disabled:opacity-50"
        >
          Remove
        </button>
      </div>
    </li>
  )
}
