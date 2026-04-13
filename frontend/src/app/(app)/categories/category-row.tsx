'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { deleteCategory, updateCategory } from '@/actions/categories'
import { CategoryForm } from './category-form'
import type { Category } from '@/lib/definitions'

export function CategoryRow({
  category,
  categories,
  householdId,
}: {
  category: Category
  categories: Category[]
  householdId: string
}) {
  const [editing, setEditing] = useState(false)
  const [pending, startTransition] = useTransition()
  const router = useRouter()

  return (
    <li className="rounded-[1.4rem] border border-[var(--line)] bg-white/65 p-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <p className="text-sm font-medium">{category.name}</p>
            {category.color && (
              <span
                className="h-3 w-3 rounded-full border border-black/10"
                style={{ backgroundColor: category.color }}
              />
            )}
          </div>
          <p className="mt-1 text-xs text-muted">
            {category.icon ? `${category.icon} · ` : ''}
            {category.parent_id ? `Child category` : 'Top-level category'}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => setEditing((value) => !value)}
            className="rounded-full border border-[var(--line)] px-3 py-1.5 text-xs text-muted transition hover:bg-white"
          >
            {editing ? 'Cancel' : 'Edit'}
          </button>
          <button
            onClick={() => {
              if (!confirm(`Delete ${category.name}?`)) return
              startTransition(async () => {
                await deleteCategory(category.id, householdId)
                router.refresh()
              })
            }}
            disabled={pending}
            className="rounded-full border border-red-200 px-3 py-1.5 text-xs text-red-600 transition hover:bg-red-50 disabled:opacity-60"
          >
            {pending ? 'Deleting…' : 'Delete'}
          </button>
        </div>
      </div>

      {editing && (
        <div className="mt-4 border-t border-[var(--line)] pt-4">
          <CategoryForm
            householdId={householdId}
            categories={categories}
            category={category}
            action={updateCategory.bind(null, category.id)}
            onSuccess={() => {
              setEditing(false)
              router.refresh()
            }}
          />
        </div>
      )}
    </li>
  )
}
