import { NextRequest, NextResponse } from 'next/server'
import { apiFetch, ApiError } from '@/lib/api'
import type { Budget } from '@/lib/definitions'

export async function POST(req: NextRequest) {
  try {
    const body = await req.json()
    const data = await apiFetch<{ budget: Budget }>('/budgets', {
      method: 'POST',
      body: JSON.stringify(body),
    })
    return NextResponse.json(data)
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status })
    }
    return NextResponse.json({ error: 'Internal error' }, { status: 500 })
  }
}
