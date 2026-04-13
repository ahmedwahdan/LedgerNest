'use client'

import Link from 'next/link'
import { useActionState } from 'react'
import { createHousehold, acceptInvitation } from '@/actions/households'
import type { ActionState } from '@/lib/definitions'

export function CreateHouseholdCard() {
  const [state, action, pending] = useActionState<ActionState, FormData>(createHousehold, null)

  return (
    <section className="glass-panel rounded-[2rem] p-6">
      <h2 className="display-font text-2xl">Create a household</h2>
      <p className="mt-2 text-sm text-muted">
        Start a shared budget with family or a partner.
      </p>

      <form action={action} className="mt-5 space-y-4">
        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Household name</span>
          <input
            name="name"
            type="text"
            placeholder="e.g. Our Home"
            required
            className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          />
        </label>

        {state && !state.success && (
          <p className="rounded-[0.9rem] bg-red-50 px-4 py-3 text-sm text-red-700">{state.error}</p>
        )}

        <button
          type="submit"
          disabled={pending}
          className="w-full rounded-full bg-[var(--accent)] px-5 py-2.5 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)] disabled:opacity-60"
        >
          {pending ? 'Creating…' : 'Create household'}
        </button>
      </form>
    </section>
  )
}

export function JoinHouseholdCard() {
  const [state, action, pending] = useActionState<ActionState, FormData>(acceptInvitation, null)

  return (
    <section className="glass-panel rounded-[2rem] p-6">
      <h2 className="display-font text-2xl">Join a household</h2>
      <p className="mt-2 text-sm text-muted">
        Paste an invitation token you received by email.
      </p>

      <form action={action} className="mt-5 space-y-4">
        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Invitation token</span>
          <input
            name="token"
            type="text"
            placeholder="Paste token here"
            required
            className="w-full rounded-[0.9rem] border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          />
        </label>

        {state && !state.success && (
          <p className="rounded-[0.9rem] bg-red-50 px-4 py-3 text-sm text-red-700">{state.error}</p>
        )}
        {state?.success && (
          <p className="rounded-[0.9rem] bg-green-50 px-4 py-3 text-sm text-green-700">
            {state.message}{' '}
            <Link href="/households" className="underline">Refresh to see it.</Link>
          </p>
        )}

        <button
          type="submit"
          disabled={pending}
          className="w-full rounded-full border border-[var(--line)] px-5 py-2.5 text-sm font-medium transition hover:bg-white/70 disabled:opacity-60"
        >
          {pending ? 'Joining…' : 'Accept invitation'}
        </button>
      </form>
    </section>
  )
}
