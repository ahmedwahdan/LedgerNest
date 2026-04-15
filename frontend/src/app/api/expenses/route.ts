import { type NextRequest, NextResponse } from 'next/server'
import { apiFetch } from '@/lib/api'
import type { Expense } from '@/lib/definitions'

export async function GET(request: NextRequest) {
  try {
    const qs = request.nextUrl.searchParams.toString()
    const data = await apiFetch<{ expenses: Expense[] }>(`/expenses${qs ? `?${qs}` : ''}`)
    return NextResponse.json(data)
  } catch {
    return NextResponse.json({ expenses: [] })
  }
}
