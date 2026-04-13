import Link from 'next/link'
import { apiFetch } from '@/lib/api'
import type { Household } from '@/lib/definitions'
import { CreateHouseholdCard, JoinHouseholdCard } from './household-cards'

async function getHouseholds(): Promise<Household[]> {
  try {
    const data = await apiFetch<{ households: Household[] }>('/households')
    return data.households ?? []
  } catch {
    return []
  }
}

export default async function HouseholdsPage() {
  const households = await getHouseholds()

  return (
    <div className="shell-grid flex flex-1 flex-col overflow-auto">
      <div className="mx-auto w-full max-w-3xl space-y-6 px-5 py-8 sm:px-8">
        <header>
          <p className="text-sm uppercase tracking-[0.28em] text-muted">Shared spending</p>
          <h1 className="display-font mt-1 text-4xl">Households</h1>
        </header>

        {/* Existing households */}
        {households.length > 0 && (
          <section>
            <h2 className="mb-3 text-sm font-medium text-[var(--muted)]">Your households</h2>
            <ul className="space-y-3">
              {households.map((h) => (
                <li key={h.id}>
                  <Link
                    href={`/households/${h.id}`}
                    className="glass-panel flex items-center justify-between rounded-[1.5rem] px-5 py-4 transition hover:bg-white/70"
                  >
                    <div className="flex items-center gap-4">
                      <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[var(--accent)]/10 text-sm font-semibold text-[var(--accent)]">
                        {h.name.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <p className="font-medium leading-snug">{h.name}</p>
                        <p className="text-xs text-[var(--muted)]">
                          Created {new Date(h.created_at).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                    <ChevronRightIcon className="h-4 w-4 text-[var(--muted)]" />
                  </Link>
                </li>
              ))}
            </ul>
          </section>
        )}

        {/* Create / Join */}
        <div className="grid gap-4 sm:grid-cols-2">
          <CreateHouseholdCard />
          <JoinHouseholdCard />
        </div>
      </div>
    </div>
  )
}

function ChevronRightIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
      <path d="M6 3l5 5-5 5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}
