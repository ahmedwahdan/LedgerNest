import { createCategory } from '@/actions/categories'
import { apiFetch } from '@/lib/api'
import { getActiveHousehold, getHouseholds } from '@/lib/household-context'
import type { Category } from '@/lib/definitions'
import { CategoryForm } from './category-form'
import { CategoryRow } from './category-row'

async function getCategories(householdId?: string) {
  const query = householdId ? `?household_id=${encodeURIComponent(householdId)}` : ''

  try {
    const data = await apiFetch<{ categories: Category[] }>(`/categories${query}`)
    return data.categories ?? []
  } catch {
    return []
  }
}

export default async function CategoriesPage() {
  const [households, activeHousehold] = await Promise.all([getHouseholds(), getActiveHousehold()])
  const categories = await getCategories(activeHousehold?.id)
  const systemCategories = categories.filter((category) => category.is_system)
  const customCategories = categories.filter((category) => !category.is_system)

  return (
    <div className="shell-grid flex flex-1 flex-col overflow-auto">
      <div className="mx-auto w-full max-w-5xl space-y-6 px-5 py-8 sm:px-8">
        <header>
          <p className="text-sm uppercase tracking-[0.28em] text-muted">Classification</p>
          <h1 className="display-font mt-1 text-4xl">Categories</h1>
          <p className="mt-2 max-w-2xl text-sm text-muted">
            Keep expense and budget classification tidy. System categories are read-only.
            Custom categories belong to your currently active household.
          </p>
        </header>

        {!activeHousehold ? (
          <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
            <p className="text-sm text-muted">
              You need a household before you can create custom categories. Create one from the
              households page first.
            </p>
          </section>
        ) : (
          <div className="grid gap-6 lg:grid-cols-[0.95fr_1.05fr]">
            <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-xs uppercase tracking-[0.28em] text-muted">Active household</p>
                  <h2 className="display-font mt-1 text-2xl">{activeHousehold.name}</h2>
                </div>
                <div className="rounded-full border border-[var(--line)] bg-white/60 px-3 py-1.5 text-xs text-muted">
                  {households.length} household{households.length === 1 ? '' : 's'}
                </div>
              </div>

              <div className="mt-6">
                <h3 className="mb-3 text-sm font-medium">Create custom category</h3>
                <CategoryForm
                  householdId={activeHousehold.id}
                  categories={customCategories}
                  action={createCategory}
                />
              </div>
            </section>

            <section className="space-y-6">
              <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
                <div className="flex items-center justify-between">
                  <h2 className="text-lg font-medium">Custom categories</h2>
                  <span className="text-sm text-muted">{customCategories.length}</span>
                </div>

                {customCategories.length === 0 ? (
                  <p className="mt-4 text-sm text-muted">
                    No custom categories yet for {activeHousehold.name}.
                  </p>
                ) : (
                  <ul className="mt-5 space-y-3">
                    {customCategories.map((category) => (
                      <CategoryRow
                        key={category.id}
                        category={category}
                        categories={customCategories}
                        householdId={activeHousehold.id}
                      />
                    ))}
                  </ul>
                )}
              </section>

              <section className="glass-panel rounded-[2rem] p-6 sm:p-8">
                <div className="flex items-center justify-between">
                  <h2 className="text-lg font-medium">System categories</h2>
                  <span className="text-sm text-muted">{systemCategories.length}</span>
                </div>

                <div className="mt-5 flex flex-wrap gap-2">
                  {systemCategories.map((category) => (
                    <span
                      key={category.id}
                      className="rounded-full border border-[var(--line)] bg-white/65 px-3 py-1.5 text-sm text-muted"
                    >
                      {category.name}
                    </span>
                  ))}
                </div>
              </section>
            </section>
          </div>
        )}
      </div>
    </div>
  )
}
