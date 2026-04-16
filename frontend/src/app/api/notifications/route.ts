import { NextResponse } from 'next/server'
import { apiFetch } from '@/lib/api'
import type { Notification } from '@/lib/definitions'

export async function GET() {
  try {
    const data = await apiFetch<{ notifications: Notification[]; unread_count: number }>(
      '/notifications?limit=20',
    )
    return NextResponse.json(data)
  } catch {
    return NextResponse.json({ notifications: [], unread_count: 0 })
  }
}
