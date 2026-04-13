'use server'

import { revalidatePath } from 'next/cache'
import { apiFetch, ApiError } from '@/lib/api'
import type { Category, ActionState } from '@/lib/definitions'

export async function createCategory(_prev: ActionState, formData: FormData): Promise<ActionState> {
  const householdId = (formData.get('household_id') as string | null)?.trim()
  const name = (formData.get('name') as string | null)?.trim()
  const parentId = (formData.get('parent_id') as string | null)?.trim()
  const icon = (formData.get('icon') as string | null)?.trim()
  const color = (formData.get('color') as string | null)?.trim()

  if (!householdId) return { success: false, error: 'Create a household first.' }
  if (!name) return { success: false, error: 'Category name is required.' }

  try {
    await apiFetch<{ category: Category }>('/categories', {
      method: 'POST',
      body: JSON.stringify({
        household_id: householdId,
        name,
        parent_id: parentId || undefined,
        icon: icon || undefined,
        color: color || undefined,
      }),
    })
  } catch (err) {
    if (err instanceof ApiError) {
      return { success: false, error: err.message }
    }
    return { success: false, error: 'Failed to create category.' }
  }

  revalidatePath('/categories')
  revalidatePath('/expenses')
  revalidatePath('/budgets')
  revalidatePath('/analytics')
  return { success: true, message: 'Category created.' }
}

export async function updateCategory(
  categoryId: string,
  _prev: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const householdId = (formData.get('household_id') as string | null)?.trim()
  const name = (formData.get('name') as string | null)?.trim()
  const parentId = (formData.get('parent_id') as string | null)?.trim()
  const icon = (formData.get('icon') as string | null)?.trim()
  const color = (formData.get('color') as string | null)?.trim()

  if (!householdId) return { success: false, error: 'Household is required.' }
  if (!name) return { success: false, error: 'Category name is required.' }

  try {
    await apiFetch<{ category: Category }>(`/categories/${categoryId}`, {
      method: 'PUT',
      body: JSON.stringify({
        household_id: householdId,
        name,
        parent_id: parentId || undefined,
        icon: icon || undefined,
        color: color || undefined,
      }),
    })
  } catch (err) {
    if (err instanceof ApiError) {
      return { success: false, error: err.message }
    }
    return { success: false, error: 'Failed to update category.' }
  }

  revalidatePath('/categories')
  revalidatePath('/expenses')
  revalidatePath('/budgets')
  revalidatePath('/analytics')
  return { success: true, message: 'Category updated.' }
}

export async function deleteCategory(categoryId: string, householdId: string) {
  await apiFetch(`/categories/${categoryId}?household_id=${encodeURIComponent(householdId)}`, {
    method: 'DELETE',
  })
  revalidatePath('/categories')
  revalidatePath('/expenses')
  revalidatePath('/budgets')
  revalidatePath('/analytics')
}
