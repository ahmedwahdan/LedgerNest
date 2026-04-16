import { NextResponse } from 'next/server'
import { apiFetch } from '@/lib/api'

export async function PUT() {
  try {
    await apiFetch('/notifications/read-all', { method: 'PUT' })
    return NextResponse.json({ ok: true })
  } catch {
    return NextResponse.json({ ok: false }, { status: 500 })
  }
}
