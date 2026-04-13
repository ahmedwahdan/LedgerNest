import { NextResponse } from 'next/server'
import { apiFetch } from '@/lib/api'

export async function PUT(_req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  try {
    await apiFetch(`/notifications/${id}/read`, { method: 'PUT' })
    return NextResponse.json({ ok: true })
  } catch {
    return NextResponse.json({ ok: false }, { status: 500 })
  }
}
