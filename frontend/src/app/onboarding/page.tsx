'use client'

import { useState } from 'react'
import Link from 'next/link'

type Step = 'household' | 'cycle' | 'budget' | 'done'

export default function OnboardingPage() {
  const [step, setStep] = useState<Step>('household')
  const [householdId, setHouseholdId] = useState<string | null>(null)
  const [snapshotId, setSnapshotId] = useState<string | null>(null)

  return (
    <main className="shell-grid flex min-h-screen flex-col items-center justify-center px-5 py-10">
      <div className="w-full max-w-lg space-y-6">
        {/* Progress */}
        <div className="flex items-center gap-2">
          {(['household', 'cycle', 'budget'] as Step[]).map((s, i) => (
            <div key={s} className="flex items-center gap-2">
              <div
                className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-medium ${
                  step === s
                    ? 'bg-[var(--accent)] text-white'
                    : step === 'done' || (['household', 'cycle', 'budget'] as Step[]).indexOf(s) < (['household', 'cycle', 'budget'] as Step[]).indexOf(step)
                      ? 'bg-[var(--accent)] text-white opacity-60'
                      : 'border border-[var(--line)] text-muted'
                }`}
              >
                {i + 1}
              </div>
              {i < 2 && <div className="h-px w-8 bg-[var(--line)]" />}
            </div>
          ))}
          <span className="ml-2 text-sm text-muted">
            {step === 'household' && 'Create household'}
            {step === 'cycle' && 'Set budget cycle'}
            {step === 'budget' && 'Set first budget'}
            {step === 'done' && 'All set!'}
          </span>
        </div>

        {step === 'household' && (
          <HouseholdStep
            onCreated={(id) => { setHouseholdId(id); setStep('cycle') }}
            onSkip={() => window.location.assign('/dashboard')}
          />
        )}
        {step === 'cycle' && householdId && (
          <CycleStep
            householdId={householdId}
            onSaved={(snapId) => { setSnapshotId(snapId); setStep('budget') }}
            onSkip={() => setStep('budget')}
          />
        )}
        {step === 'budget' && (
          <BudgetStep
            snapshotId={snapshotId ?? undefined}
            onSaved={() => setStep('done')}
            onSkip={() => setStep('done')}
          />
        )}
        {step === 'done' && <DoneStep />}
      </div>
    </main>
  )
}

function HouseholdStep({
  onCreated,
  onSkip,
}: {
  onCreated: (id: string) => void
  onSkip: () => void
}) {
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  async function handleSubmit(formData: FormData) {
    setPending(true)
    setError(null)
    try {
      // We need the household id back, so we call the API directly here
      const res = await fetch('/api/households', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: formData.get('name') }),
      })
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        setError(body.error ?? 'Failed to create household.')
        return
      }
      const data = await res.json()
      onCreated(data.household.id)
    } catch {
      setError('Failed to create household.')
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="glass-panel rounded-[2rem] p-8">
      <h1 className="display-font text-3xl">Set up your household</h1>
      <p className="mt-2 text-sm text-muted">
        Give your household a name. You can invite others later.
      </p>

      <form
        onSubmit={async (e) => {
          e.preventDefault()
          await handleSubmit(new FormData(e.currentTarget))
        }}
        className="mt-6 space-y-4"
      >
        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Household name</span>
          <input
            name="name"
            type="text"
            placeholder="e.g. Our Home"
            required
            className="w-full rounded-xl border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          />
        </label>

        {error && (
          <p className="rounded-xl bg-red-50 px-4 py-3 text-sm text-red-700">{error}</p>
        )}

        <button
          type="submit"
          disabled={pending}
          className="w-full rounded-full bg-[var(--accent)] px-6 py-3 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)] disabled:opacity-60"
        >
          {pending ? 'Creating…' : 'Create & continue'}
        </button>
      </form>

      <button
        onClick={onSkip}
        className="mt-3 w-full text-center text-sm text-muted underline underline-offset-2"
      >
        Skip for now
      </button>
    </section>
  )
}

function CycleStep({
  householdId,
  onSaved,
  onSkip,
}: {
  householdId: string
  onSaved: (snapshotId: string) => void
  onSkip: () => void
}) {
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  async function handleSubmit(formData: FormData) {
    setPending(true)
    setError(null)
    const startDay = parseInt(formData.get('start_day') as string, 10)
    if (!startDay || startDay < 1 || startDay > 28) {
      setError('Start day must be between 1 and 28.')
      setPending(false)
      return
    }
    try {
      const res = await fetch(`/api/households/${householdId}/cycle`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ start_day: startDay }),
      })
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        setError(body.error ?? 'Failed to set cycle.')
        return
      }
      const data = await res.json()
      onSaved(data.current_snapshot?.id ?? '')
    } catch {
      setError('Failed to set cycle.')
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="glass-panel rounded-[2rem] p-8">
      <h1 className="display-font text-3xl">Budget cycle</h1>
      <p className="mt-2 text-sm text-muted">
        Which day of the month does your budget cycle start? Usually aligned to payday.
      </p>

      <form
        onSubmit={async (e) => {
          e.preventDefault()
          await handleSubmit(new FormData(e.currentTarget))
        }}
        className="mt-6 space-y-4"
      >
        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Cycle start day (1–28)</span>
          <input
            name="start_day"
            type="number"
            min={1}
            max={28}
            defaultValue={1}
            required
            className="w-full rounded-xl border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          />
        </label>

        {error && (
          <p className="rounded-xl bg-red-50 px-4 py-3 text-sm text-red-700">{error}</p>
        )}

        <button
          type="submit"
          disabled={pending}
          className="w-full rounded-full bg-[var(--accent)] px-6 py-3 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)] disabled:opacity-60"
        >
          {pending ? 'Saving…' : 'Set cycle & continue'}
        </button>
      </form>

      <button
        onClick={onSkip}
        className="mt-3 w-full text-center text-sm text-muted underline underline-offset-2"
      >
        Skip for now
      </button>
    </section>
  )
}

function BudgetStep({
  snapshotId,
  onSaved,
  onSkip,
}: {
  snapshotId?: string
  onSaved: () => void
  onSkip: () => void
}) {
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  async function handleSubmit(formData: FormData) {
    setPending(true)
    setError(null)
    const amount = formData.get('amount') as string
    if (!amount || parseFloat(amount) <= 0) {
      setError('Enter a positive amount.')
      setPending(false)
      return
    }
    try {
      const res = await fetch('/api/budgets', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          scope: 'personal',
          amount,
          snapshot_id: snapshotId,
        }),
      })
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        setError(body.error ?? 'Failed to create budget.')
        return
      }
      onSaved()
    } catch {
      setError('Failed to create budget.')
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="glass-panel rounded-[2rem] p-8">
      <h1 className="display-font text-3xl">Monthly budget</h1>
      <p className="mt-2 text-sm text-muted">
        Set an overall monthly spending cap. You can refine this by category later.
      </p>

      <form
        onSubmit={async (e) => {
          e.preventDefault()
          await handleSubmit(new FormData(e.currentTarget))
        }}
        className="mt-6 space-y-4"
      >
        <label className="block">
          <span className="mb-1.5 block text-sm text-muted">Monthly cap (USD)</span>
          <input
            name="amount"
            type="text"
            inputMode="decimal"
            placeholder="e.g. 2500.00"
            required
            className="w-full rounded-xl border border-[var(--line)] bg-white/80 px-4 py-2.5 text-sm outline-none transition focus:border-[var(--accent)]"
          />
        </label>

        {error && (
          <p className="rounded-xl bg-red-50 px-4 py-3 text-sm text-red-700">{error}</p>
        )}

        <button
          type="submit"
          disabled={pending}
          className="w-full rounded-full bg-[var(--accent)] px-6 py-3 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)] disabled:opacity-60"
        >
          {pending ? 'Saving…' : 'Set budget & finish'}
        </button>
      </form>

      <button
        onClick={onSkip}
        className="mt-3 w-full text-center text-sm text-muted underline underline-offset-2"
      >
        Skip for now
      </button>
    </section>
  )
}

function DoneStep() {
  return (
    <section className="glass-panel rounded-[2rem] p-8 text-center">
      <div className="mx-auto mb-5 flex h-14 w-14 items-center justify-center rounded-full bg-[rgba(15,118,110,0.12)] text-2xl">
        ✓
      </div>
      <h1 className="display-font text-3xl">You&apos;re all set</h1>
      <p className="mt-2 text-sm text-muted">
        Your household, cycle, and first budget are configured. Head to the dashboard.
      </p>
      <Link
        href="/dashboard"
        className="mt-6 inline-block rounded-full bg-[var(--accent)] px-8 py-3 text-sm font-medium text-white transition hover:bg-[var(--accent-strong)]"
      >
        Go to dashboard
      </Link>
    </section>
  )
}
