'use client'

import { useActionState } from 'react'
import { inviteMember } from '@/actions/households'
import type { ActionState } from '@/lib/definitions'

export function InvitePanel({ householdId }: { householdId: string }) {
  const boundInvite = inviteMember.bind(null, householdId)
  const [state, action, pending] = useActionState<ActionState, FormData>(boundInvite, null)

  return (
    <section className="glass-panel rounded-[2rem] p-6">
      <h2 className="mb-4 text-sm font-medium uppercase tracking-[0.2em] text-muted">
        Invite someone
      </h2>

      <form action={action} className="flex flex-wrap gap-3">
        <input
          name="email"
          type="email"
          placeholder="their@email.com"
          required
          className="flex-1 min-w-0 rounded-xl border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
        />
        <select
          name="role"
          defaultValue="editor"
          className="rounded-xl border border-[var(--line)] bg-white/80 px-3 py-2.5 text-sm outline-none focus:border-[var(--accent)]"
        >
          <option value="editor">Editor</option>
          <option value="viewer">Viewer</option>
          <option value="owner">Owner</option>
        </select>
        <button
          type="submit"
          disabled={pending}
          className="rounded-full bg-[var(--accent)] px-5 py-2.5 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)] disabled:opacity-60"
        >
          {pending ? 'Sending…' : 'Invite'}
        </button>
      </form>

      {state && !state.success && (
        <p className="mt-3 rounded-xl bg-red-50 px-4 py-3 text-sm text-red-700">{state.error}</p>
      )}
      {state?.success && (
        <p className="mt-3 rounded-xl bg-green-50 px-4 py-3 text-sm text-green-700">
          {state.message ?? 'Invitation sent.'}
        </p>
      )}
    </section>
  )
}
