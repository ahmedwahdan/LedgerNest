'use client'

import Link from 'next/link'
import { useActionState } from 'react'
import { login } from '@/actions/auth'
import type { ActionState } from '@/lib/definitions'

export default function LoginPage() {
  const [state, action, pending] = useActionState<ActionState, FormData>(login, null)

  return (
    <main className="shell-grid flex flex-1 items-center justify-center px-5 py-10 sm:px-8">
      <div className="grid w-full max-w-6xl gap-6 lg:grid-cols-[0.95fr_1.05fr]">
        <section className="glass-panel rounded-[2rem] p-8 sm:p-10">
          <div className="inline-flex rounded-full border border-[var(--line)] bg-white/65 px-4 py-2 text-xs uppercase tracking-[0.28em] text-muted">
            LedgerNest
          </div>
          <h1 className="display-font mt-6 text-5xl leading-none">
            Step into the household ledger.
          </h1>
          <p className="mt-5 max-w-md text-base leading-7 text-muted">
            Track personal and shared spending with clear budget signals for every cycle.
          </p>
          <div className="mt-10 grid gap-4">
            {[
              'Personal and household expense tracking.',
              'Budget health with real-time cycle progress.',
              'Full audit history on every change.',
            ].map((point) => (
              <div
                key={point}
                className="rounded-[1.2rem] border border-[var(--line)] bg-white/65 px-4 py-4 text-sm text-muted"
              >
                {point}
              </div>
            ))}
          </div>
        </section>

        <section className="glass-panel rounded-[2rem] p-8 sm:p-10">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm uppercase tracking-[0.28em] text-muted">Sign in</p>
              <h2 className="display-font mt-2 text-4xl">Welcome back</h2>
            </div>
            <Link
              href="/"
              className="rounded-full border border-[var(--line)] px-4 py-2 text-sm transition hover:bg-white/70"
            >
              Home
            </Link>
          </div>

          <form action={action} className="mt-8 space-y-5">
            <label className="block">
              <span className="mb-2 block text-sm text-muted">Email</span>
              <input
                name="email"
                type="email"
                placeholder="you@example.com"
                required
                className="w-full rounded-[1rem] border border-[var(--line)] bg-white/80 px-4 py-3 outline-none transition focus:border-[var(--accent)]"
              />
            </label>

            <label className="block">
              <span className="mb-2 block text-sm text-muted">Password</span>
              <input
                name="password"
                type="password"
                placeholder="Enter your password"
                required
                className="w-full rounded-[1rem] border border-[var(--line)] bg-white/80 px-4 py-3 outline-none transition focus:border-[var(--accent)]"
              />
            </label>

            {state && !state.success && (
              <p className="rounded-[1rem] bg-red-50 px-4 py-3 text-sm text-red-700">
                {state.error}
              </p>
            )}

            <button
              type="submit"
              disabled={pending}
              className="w-full rounded-full bg-[var(--accent)] px-6 py-3 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)] disabled:opacity-60"
            >
              {pending ? 'Signing in…' : 'Continue to dashboard'}
            </button>
          </form>

          <p className="mt-6 text-center text-sm text-muted">
            No account?{' '}
            <Link href="/register" className="text-[var(--accent)] underline underline-offset-2">
              Create one
            </Link>
          </p>
        </section>
      </div>
    </main>
  )
}
