import { Sidebar } from '@/components/sidebar'
import { MobileNav } from '@/components/mobile-nav'

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-full min-h-screen">
      {/* Desktop sidebar */}
      <aside className="hidden w-56 shrink-0 border-r border-[var(--line)] bg-[var(--surface)] lg:block">
        <Sidebar />
      </aside>

      {/* Main content — extra bottom padding on mobile for the nav bar */}
      <main className="flex flex-1 flex-col overflow-hidden pb-16 lg:pb-0">
        {children}
      </main>

      {/* Mobile bottom nav */}
      <MobileNav />
    </div>
  )
}
