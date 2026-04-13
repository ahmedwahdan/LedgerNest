'use server'

import { revalidatePath } from 'next/cache'
import { redirect } from 'next/navigation'
import { apiFetch, ApiError } from '@/lib/api'
import type { Household } from '@/lib/definitions'
import type { ActionState } from '@/lib/definitions'

export async function createHousehold(
  _prev: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const name = (formData.get('name') as string).trim()
  if (!name) return { success: false, error: 'Name is required.' }

  let household: Household
  try {
    const res = await apiFetch<{ household: Household }>('/households', {
      method: 'POST',
      body: JSON.stringify({ name }),
    })
    household = res.household
  } catch (err) {
    if (err instanceof ApiError) return { success: false, error: err.message }
    return { success: false, error: 'Failed to create household.' }
  }

  revalidatePath('/households')
  redirect(`/households/${household.id}`)
}

export async function updateHousehold(
  householdId: string,
  _prev: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const name = (formData.get('name') as string).trim()
  if (!name) return { success: false, error: 'Name is required.' }

  try {
    await apiFetch(`/households/${householdId}`, {
      method: 'PUT',
      body: JSON.stringify({ name }),
    })
  } catch (err) {
    if (err instanceof ApiError) return { success: false, error: err.message }
    return { success: false, error: 'Failed to update household.' }
  }

  revalidatePath(`/households/${householdId}`)
  return { success: true }
}

export async function deleteHousehold(householdId: string) {
  await apiFetch(`/households/${householdId}`, { method: 'DELETE' })
  revalidatePath('/households')
  redirect('/households')
}

export async function leaveHousehold(householdId: string) {
  await apiFetch(`/households/${householdId}/leave`, { method: 'POST' })
  revalidatePath('/households')
  redirect('/households')
}

export async function updateMemberRole(
  householdId: string,
  userId: string,
  role: string,
) {
  await apiFetch(`/households/${householdId}/members/${userId}/role`, {
    method: 'PUT',
    body: JSON.stringify({ role }),
  })
  revalidatePath(`/households/${householdId}`)
}

export async function removeMember(householdId: string, userId: string) {
  await apiFetch(`/households/${householdId}/members/${userId}`, { method: 'DELETE' })
  revalidatePath(`/households/${householdId}`)
}

export async function inviteMember(
  householdId: string,
  _prev: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const email = (formData.get('email') as string).trim()
  const role = formData.get('role') as string

  if (!email) return { success: false, error: 'Email is required.' }

  try {
    await apiFetch(`/households/${householdId}/invitations`, {
      method: 'POST',
      body: JSON.stringify({ email, role: role || 'editor' }),
    })
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 409) return { success: false, error: 'This person is already a member or has a pending invitation.' }
      return { success: false, error: err.message }
    }
    return { success: false, error: 'Failed to send invitation.' }
  }

  revalidatePath(`/households/${householdId}`)
  return { success: true, message: `Invitation sent to ${email}` }
}

export async function revokeInvitation(householdId: string, invitationId: string) {
  await apiFetch(`/households/${householdId}/invitations/${invitationId}`, { method: 'DELETE' })
  revalidatePath(`/households/${householdId}`)
}

export async function acceptInvitation(
  _prev: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const token = (formData.get('token') as string).trim()
  if (!token) return { success: false, error: 'Token is required.' }

  try {
    await apiFetch('/invitations/accept', {
      method: 'POST',
      body: JSON.stringify({ token }),
    })
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 404) return { success: false, error: 'Invitation not found or already used.' }
      if (err.status === 410) return { success: false, error: 'Invitation has expired.' }
      return { success: false, error: err.message }
    }
    return { success: false, error: 'Failed to accept invitation.' }
  }

  revalidatePath('/households')
  return { success: true, message: 'You have joined the household.' }
}

export async function setCycleConfig(
  householdId: string,
  _prev: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const startDay = parseInt(formData.get('start_day') as string, 10)
  if (!startDay || startDay < 1 || startDay > 28) {
    return { success: false, error: 'Start day must be between 1 and 28.' }
  }

  try {
    await apiFetch(`/households/${householdId}/cycle`, {
      method: 'PUT',
      body: JSON.stringify({ start_day: startDay }),
    })
  } catch (err) {
    if (err instanceof ApiError) return { success: false, error: err.message }
    return { success: false, error: 'Failed to update cycle.' }
  }

  revalidatePath(`/households/${householdId}`)
  return { success: true }
}
