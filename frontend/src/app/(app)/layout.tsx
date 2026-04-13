import { redirect } from 'next/navigation'
import { HouseholdSwitcher } from '@/components/household-switcher'
import { Sidebar } from '@/components/sidebar'
import { MobileNav } from '@/components/mobile-nav'
import { NotificationBell } from '@/components/notification-bell'
import { getCurrentUser } from '@/lib/current-user'
import { getActiveHousehold, getHouseholds } from '@/lib/household-context'

export default async function AppLayout({ children }: { children: React.ReactNode }) {
  const user = await getCurrentUser()

  if (!user) {
    redirect('/login')
  }

  const [households, activeHousehold] = await Promise.all([
    getHouseholds(),
    getActiveHousehold(),
  ])

  return (
    <div className="flex h-full min-h-screen">
      {/* Desktop sidebar */}
      <aside className="hidden w-56 shrink-0 border-r border-[var(--line)] bg-[var(--surface)] lg:block">
        <Sidebar user={user} />
      </aside>

      {/* Main content area */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Top header bar */}
        <header className="flex h-14 shrink-0 items-center justify-between border-b border-[var(--line)] bg-[var(--surface)] px-4 sm:px-6">
          <div>
            <p className="text-xs uppercase tracking-[0.28em] text-muted">Signed in</p>
            <p className="mt-1 text-sm font-medium">{user.display_name}</p>
          </div>
          <div className="flex items-center gap-3">
            <HouseholdSwitcher
              households={households}
              activeHouseholdId={activeHousehold?.id}
            />
            <div className="hidden rounded-full border border-[var(--line)] bg-white/60 px-3 py-1.5 text-xs text-muted sm:block">
              {user.preferred_currency}
            </div>
            <NotificationBell />
          </div>
        </header>

        {/* Page content — extra bottom padding on mobile for the nav bar */}
        <main className="flex flex-1 flex-col overflow-hidden pb-16 lg:pb-0">
          {children}
        </main>
      </div>

      {/* Mobile bottom nav */}
      <MobileNav />
    </div>
  )
}
