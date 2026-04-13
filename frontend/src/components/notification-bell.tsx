'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import type { Notification } from '@/lib/definitions'

interface NotificationsData {
  notifications: Notification[]
  unread_count: number
}

export function NotificationBell() {
  const [open, setOpen] = useState(false)
  const [data, setData] = useState<NotificationsData>({ notifications: [], unread_count: 0 })
  const [loading, setLoading] = useState(false)
  const panelRef = useRef<HTMLDivElement>(null)

  const fetchNotifications = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fetch('/api/notifications')
      if (res.ok) setData(await res.json())
    } finally {
      setLoading(false)
    }
  }, [])

  // Fetch count on mount (silently)
  useEffect(() => {
    fetch('/api/notifications')
      .then((r) => r.json())
      .then((d: NotificationsData) => setData(d))
      .catch(() => {})
  }, [])

  // Close panel on outside click
  useEffect(() => {
    if (!open) return
    function onMouseDown(e: MouseEvent) {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', onMouseDown)
    return () => document.removeEventListener('mousedown', onMouseDown)
  }, [open])

  function toggle() {
    if (!open) fetchNotifications()
    setOpen((v) => !v)
  }

  async function markRead(id: string) {
    await fetch(`/api/notifications/${id}`, { method: 'PUT' })
    setData((prev) => ({
      ...prev,
      unread_count: Math.max(0, prev.unread_count - (prev.notifications.find((n) => n.id === id && !n.read_at) ? 1 : 0)),
      notifications: prev.notifications.map((n) =>
        n.id === id ? { ...n, read_at: new Date().toISOString() } : n,
      ),
    }))
  }

  async function markAllRead() {
    await fetch('/api/notifications/read-all', { method: 'PUT' })
    setData((prev) => ({
      unread_count: 0,
      notifications: prev.notifications.map((n) => ({
        ...n,
        read_at: n.read_at ?? new Date().toISOString(),
      })),
    }))
  }

  return (
    <div className="relative" ref={panelRef}>
      <button
        onClick={toggle}
        aria-label="Notifications"
        className="relative flex h-9 w-9 items-center justify-center rounded-full text-[var(--muted)] transition hover:bg-white/70 hover:text-[var(--foreground)]"
      >
        <BellIcon className="h-5 w-5" />
        {data.unread_count > 0 && (
          <span className="absolute right-1 top-1 flex h-4 w-4 items-center justify-center rounded-full bg-[var(--accent)] text-[10px] font-semibold text-white">
            {data.unread_count > 9 ? '9+' : data.unread_count}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 top-11 z-50 w-80 overflow-hidden rounded-[1.2rem] border border-[var(--line)] bg-[var(--surface)] shadow-xl">
          <div className="flex items-center justify-between border-b border-[var(--line)] px-4 py-3">
            <span className="text-sm font-medium">Notifications</span>
            {data.unread_count > 0 && (
              <button
                onClick={markAllRead}
                className="text-xs text-[var(--accent)] hover:underline"
              >
                Mark all read
              </button>
            )}
          </div>

          <div className="max-h-96 overflow-y-auto">
            {loading ? (
              <div className="space-y-3 p-4">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="skeleton h-14 rounded-xl" />
                ))}
              </div>
            ) : data.notifications.length === 0 ? (
              <div className="flex flex-col items-center gap-2 px-4 py-10 text-center">
                <BellIcon className="h-8 w-8 text-[var(--muted)] opacity-40" />
                <p className="text-sm text-[var(--muted)]">No notifications yet</p>
              </div>
            ) : (
              <ul>
                {data.notifications.map((n) => (
                  <li
                    key={n.id}
                    className={`cursor-pointer border-b border-[var(--line)] px-4 py-3 last:border-0 transition hover:bg-white/50 ${
                      !n.read_at ? 'bg-[var(--accent)]/5' : ''
                    }`}
                    onClick={() => !n.read_at && markRead(n.id)}
                  >
                    <div className="flex items-start gap-3">
                      <div className={`mt-1 h-2 w-2 shrink-0 rounded-full ${!n.read_at ? 'bg-[var(--accent)]' : 'bg-transparent'}`} />
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium leading-snug">{n.title}</p>
                        <p className="mt-0.5 text-xs text-[var(--muted)] leading-snug line-clamp-2">{n.body}</p>
                        <p className="mt-1 text-[10px] text-[var(--muted)] opacity-60">
                          {new Date(n.created_at).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

function BellIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
      <path d="M8 1.5a4.5 4.5 0 0 0-4.5 4.5v2.5L2 10h12l-1.5-1.5V6A4.5 4.5 0 0 0 8 1.5Z" strokeLinejoin="round" />
      <path d="M6.5 10.5a1.5 1.5 0 0 0 3 0" strokeLinecap="round" />
    </svg>
  )
}
