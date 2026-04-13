'use server'

import { revalidatePath } from 'next/cache'
import { redirect } from 'next/navigation'
import { apiFetch, ApiError } from '@/lib/api'
import type { Expense } from '@/lib/definitions'
import type { ActionState } from '@/lib/definitions'

export async function createExpense(_prev: ActionState, formData: FormData): Promise<ActionState> {
  const amount = formData.get('amount') as string
  const currency = formData.get('currency') as string
  const merchant = formData.get('merchant') as string
  const paymentMethod = formData.get('payment_method') as string
  const date = formData.get('date') as string
  const notes = formData.get('notes') as string | null
  const categoryId = formData.get('category_id') as string | null

  try {
    await apiFetch<{ expense: Expense }>('/expenses', {
      method: 'POST',
      body: JSON.stringify({
        amount,
        currency: currency || 'USD',
        merchant,
        payment_method: paymentMethod,
        date,
        notes: notes || undefined,
        category_id: categoryId || undefined,
      }),
    })
  } catch (err) {
    if (err instanceof ApiError) {
      return { success: false, error: err.message }
    }
    return { success: false, error: 'Failed to create expense.' }
  }

  revalidatePath('/expenses')
  revalidatePath('/dashboard')
  return { success: true }
}

export async function updateExpense(
  expenseId: string,
  _prev: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const amount = formData.get('amount') as string
  const currency = formData.get('currency') as string
  const merchant = formData.get('merchant') as string
  const paymentMethod = formData.get('payment_method') as string
  const date = formData.get('date') as string
  const notes = formData.get('notes') as string | null
  const categoryId = formData.get('category_id') as string | null

  try {
    await apiFetch<{ expense: Expense }>(`/expenses/${expenseId}`, {
      method: 'PUT',
      body: JSON.stringify({
        amount,
        currency: currency || 'USD',
        merchant,
        payment_method: paymentMethod,
        date,
        notes: notes || undefined,
        category_id: categoryId || undefined,
      }),
    })
  } catch (err) {
    if (err instanceof ApiError) {
      return { success: false, error: err.message }
    }
    return { success: false, error: 'Failed to update expense.' }
  }

  revalidatePath('/expenses')
  revalidatePath('/dashboard')
  return { success: true }
}

export async function deleteExpense(expenseId: string) {
  await apiFetch(`/expenses/${expenseId}`, { method: 'DELETE' })
  revalidatePath('/expenses')
  revalidatePath('/dashboard')
  redirect('/expenses')
}

export async function restoreExpense(expenseId: string) {
  await apiFetch(`/expenses/${expenseId}/restore`, { method: 'POST' })
  revalidatePath('/expenses')
  revalidatePath('/dashboard')
}
