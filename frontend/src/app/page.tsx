import Link from "next/link";

const signals = [
  { label: "Groceries", value: "$412", tone: "Running 8% under plan" },
  { label: "Dining", value: "$138", tone: "Stable over last 14 days" },
  { label: "Utilities", value: "$96", tone: "Next bill expected in 3 days" },
];

const timeline = [
  "Capture personal and shared spending without losing ownership context.",
  "Watch budget pace shift in real time as expenses land.",
  "Keep everyone aligned with one household view instead of scattered notes.",
];

export default function Home() {
  return (
    <main className="shell-grid flex flex-1 flex-col overflow-hidden">
      <section className="mx-auto flex w-full max-w-7xl flex-1 flex-col px-5 py-6 sm:px-8 lg:px-10">
        <header className="glass-panel flex items-center justify-between rounded-full px-4 py-3 sm:px-6">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[var(--accent)] text-sm font-semibold text-white">
              LN
            </div>
            <div>
              <p className="display-font text-xl leading-none">LedgerNest</p>
              <p className="text-xs uppercase tracking-[0.28em] text-muted">
                Household finance cockpit
              </p>
            </div>
          </div>
          <nav className="hidden items-center gap-6 text-sm text-muted md:flex">
            <span>Budgets</span>
            <span>Shared Expenses</span>
            <span>Insights</span>
          </nav>
          <div className="flex items-center gap-3">
            <Link
              href="/login"
              className="rounded-full border border-[var(--line)] px-4 py-2 text-sm transition hover:bg-white/60"
            >
              Log in
            </Link>
            <a
              href="#preview"
              className="rounded-full bg-[var(--foreground)] px-4 py-2 text-sm text-[var(--background)] transition hover:opacity-90"
            >
              See preview
            </a>
          </div>
        </header>

        <div className="grid flex-1 gap-6 py-6 lg:grid-cols-[1.1fr_0.9fr]">
          <section className="glass-panel relative overflow-hidden rounded-[2rem] p-6 sm:p-10">
            <div className="absolute right-0 top-0 h-48 w-48 rounded-full bg-[rgba(15,118,110,0.14)] blur-3xl" />
            <div className="absolute bottom-8 left-8 h-32 w-32 rounded-full bg-[rgba(209,127,47,0.16)] blur-3xl" />

            <div className="relative z-10 flex h-full flex-col justify-between gap-10">
              <div className="space-y-6">
                <div className="inline-flex rounded-full border border-[var(--line)] bg-white/60 px-4 py-2 text-xs uppercase tracking-[0.28em] text-muted">
                  Backend ready. Frontend starts now.
                </div>
                <div className="max-w-2xl space-y-4">
                  <h1 className="display-font text-5xl leading-[0.95] tracking-tight sm:text-6xl lg:text-7xl">
                    Budgeting that feels shared, not spreadsheeted.
                  </h1>
                  <p className="max-w-xl text-base leading-7 text-muted sm:text-lg">
                    LedgerNest is moving from backend foundation into a visible
                    product surface. This first shell sets the tone for a calm,
                    household-first budgeting experience with sharper signals and
                    less admin drag.
                  </p>
                </div>
                <div className="flex flex-col gap-3 sm:flex-row">
                  <Link
                    href="/login"
                    className="rounded-full bg-[var(--accent)] px-6 py-3 text-center text-sm font-medium text-white transition hover:bg-[var(--accent-strong)]"
                  >
                    Open sign-in flow
                  </Link>
                  <a
                    href="#preview"
                    className="rounded-full border border-[var(--line)] px-6 py-3 text-center text-sm font-medium transition hover:bg-white/70"
                  >
                    Explore dashboard concept
                  </a>
                </div>
              </div>

              <div className="grid gap-4 sm:grid-cols-3">
                {signals.map((signal) => (
                  <article
                    key={signal.label}
                    className="rounded-[1.4rem] border border-[var(--line)] bg-white/70 p-4"
                  >
                    <p className="text-sm text-muted">{signal.label}</p>
                    <p className="mt-3 text-2xl font-semibold">{signal.value}</p>
                    <p className="mt-2 text-sm text-muted">{signal.tone}</p>
                  </article>
                ))}
              </div>
            </div>
          </section>

          <section id="preview" className="flex flex-col gap-6">
            <article className="glass-panel rounded-[2rem] p-6 sm:p-7">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-sm uppercase tracking-[0.28em] text-muted">
                    April cycle
                  </p>
                  <h2 className="display-font mt-2 text-3xl">Household pulse</h2>
                </div>
                <div className="rounded-full bg-[rgba(15,118,110,0.12)] px-3 py-2 text-sm text-[var(--accent-strong)]">
                  78% on track
                </div>
              </div>

              <div className="mt-8 space-y-5">
                <div>
                  <div className="mb-2 flex justify-between text-sm">
                    <span className="text-muted">Monthly cap</span>
                    <span>$1,980 / $2,540</span>
                  </div>
                  <div className="h-3 rounded-full bg-[rgba(19,33,27,0.08)]">
                    <div className="h-3 w-[78%] rounded-full bg-[var(--accent)]" />
                  </div>
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="rounded-[1.4rem] bg-[var(--surface-strong)] p-4">
                    <p className="text-sm text-muted">Most active merchant</p>
                    <p className="mt-3 text-xl font-semibold">Albert Heijn</p>
                    <p className="mt-1 text-sm text-muted">6 transactions this cycle</p>
                  </div>
                  <div className="rounded-[1.4rem] bg-[var(--surface-strong)] p-4">
                    <p className="text-sm text-muted">Pace warning</p>
                    <p className="mt-3 text-xl font-semibold text-[var(--warning)]">
                      Dining trending high
                    </p>
                    <p className="mt-1 text-sm text-muted">$18/day average vs $11 planned</p>
                  </div>
                </div>
              </div>
            </article>

            <article className="glass-panel rounded-[2rem] p-6 sm:p-7">
              <p className="text-sm uppercase tracking-[0.28em] text-muted">
                Build path
              </p>
              <ol className="mt-5 space-y-4">
                {timeline.map((item, index) => (
                  <li
                    key={item}
                    className="flex gap-4 rounded-[1.25rem] border border-[var(--line)] bg-white/65 p-4"
                  >
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-[var(--foreground)] text-sm text-[var(--background)]">
                      {index + 1}
                    </div>
                    <p className="text-sm leading-6 text-muted">{item}</p>
                  </li>
                ))}
              </ol>
            </article>
          </section>
        </div>
      </section>
    </main>
  );
}
