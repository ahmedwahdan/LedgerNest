'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'

const mobileItems = [
  { href: '/dashboard', label: 'Home', icon: HomeIcon },
  { href: '/expenses', label: 'Expenses', icon: ExpensesIcon },
  { href: '/categories', label: 'Categories', icon: CategoriesIcon },
  { href: '/budgets', label: 'Budgets', icon: BudgetsIcon },
  { href: '/households', label: 'Household', icon: HouseholdsIcon },
]

export function MobileNav() {
  const pathname = usePathname()

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-40 border-t border-[var(--line)] bg-[var(--surface)] backdrop-blur-sm lg:hidden">
      <ul className="flex">
        {mobileItems.map(({ href, label, icon: Icon }) => {
          const active = pathname === href || pathname.startsWith(href + '/')
          return (
            <li key={href} className="flex-1">
              <Link
                href={href}
                className={`flex flex-col items-center gap-1 py-3 text-xs transition ${
                  active ? 'text-[var(--accent)]' : 'text-muted'
                }`}
              >
                <Icon className={`h-5 w-5 ${active ? 'stroke-[var(--accent)]' : ''}`} />
                {label}
              </Link>
            </li>
          )
        })}
      </ul>
    </nav>
  )
}

function HomeIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
      <path d="M1 7.5 8 2l7 5.5V14.5H1V7.5Z" strokeLinejoin="round" />
      <path d="M5.5 14.5v-4h5v4" strokeLinejoin="round" />
    </svg>
  )
}

function ExpensesIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
      <rect x="1.5" y="2.5" width="13" height="11" rx="1.5" />
      <path d="M1.5 6h13M5 9.5h2M5 11.5h4" strokeLinecap="round" />
    </svg>
  )
}

function BudgetsIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
      <circle cx="8" cy="8" r="6.5" />
      <path d="M8 4v4l2.5 2.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function CategoriesIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
      <path d="M3 3.5h10v3H3zM3 9.5h6v3H3zM11 9.5h2v3h-2z" strokeLinejoin="round" />
    </svg>
  )
}

function HouseholdsIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
      <circle cx="8" cy="6" r="3" />
      <path d="M2 14c0-2.21 2.686-4 6-4s6 1.79 6 4" strokeLinecap="round" />
    </svg>
  )
}
