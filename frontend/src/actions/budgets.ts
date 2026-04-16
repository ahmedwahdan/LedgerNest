'use server'

import { revalidatePath } from 'next/cache'
import { apiFetch, ApiError } from '@/lib/api'
import type { Budget } from '@/lib/definitions'
import type { ActionState } from '@/lib/definitions'

export async function createBudget(_prev: ActionState, formData: FormData): Promise<ActionState> {
  const amount = formData.get('amount') as string
  const householdId = formData.get('household_id') as string | null
  const snapshotId = formData.get('snapshot_id') as string | null
  const categoryId = formData.get('category_id') as string | null

  if (!householdId) {
    return { success: false, error: 'Select a household before creating a budget.' }
  }

  try {
    await apiFetch<{ budget: Budget }>('/budgets', {
      method: 'POST',
      body: JSON.stringify({
        scope: 'household',
        household_id: householdId,
        snapshot_id: snapshotId || undefined,
        category_id: categoryId || undefined,
        amount,
      }),
    })
  } catch (err) {
    if (err instanceof ApiError) {
      return { success: false, error: err.message }
    }
    return { success: false, error: 'Failed to create budget.' }
  }

  revalidatePath('/budgets')
  revalidatePath('/dashboard')
  return { success: true }
}

export async function deleteBudget(budgetId: string, householdId: string) {
  await apiFetch(`/budgets/${budgetId}?household_id=${encodeURIComponent(householdId)}`, {
    method: 'DELETE',
  })
  revalidatePath('/budgets')
  revalidatePath('/dashboard')
}
