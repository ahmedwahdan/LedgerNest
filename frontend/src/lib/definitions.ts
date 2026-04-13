export interface User {
  id: string
  email: string
  display_name: string
  preferred_currency: string
  created_at: string
  updated_at: string
}

export interface Expense {
  id: string
  scope: string
  user_id?: string
  household_id?: string
  created_by: string
  updated_by?: string
  amount: string
  currency: string
  merchant: string
  category_id?: string
  payment_method: string
  date: string
  notes?: string
  is_recurring: boolean
  recurrence_interval?: string
  is_deleted?: boolean
  deleted_at?: string
  created_at: string
  updated_at: string
}

export interface Category {
  id: string
  household_id?: string
  name: string
  parent_id?: string
  icon?: string
  color?: string
  is_system: boolean
  created_at: string
}

export interface BudgetCycleConfig {
  id: string
  household_id: string
  start_day: number
  effective_from: string
  created_by: string
  created_at: string
}

export interface CycleSnapshot {
  id: string
  household_id: string
  cycle_start: string
  cycle_end: string
  label: string
  status: string
  config_id: string
  created_at: string
}

export interface Budget {
  id: string
  scope: string
  user_id?: string
  household_id?: string
  category_id?: string
  cycle_snapshot_id: string
  amount: string
  rollover_amount: string
  created_at: string
  updated_at: string
}

export interface BudgetHealthItem {
  budget_id: string
  category_id?: string
  category_name?: string
  amount: string
  rollover: string
  spent: string
  remaining: string
  pct_used: number
}

export interface BudgetHealth {
  snapshot: CycleSnapshot
  overall?: BudgetHealthItem
  categories: BudgetHealthItem[]
}

export interface Household {
  id: string
  name: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface AuditLogEntry {
  id: string
  user_id?: string
  action: string
  entity_type: string
  entity_id: string
  old_values?: Record<string, unknown>
  new_values?: Record<string, unknown>
  created_at: string
}

export interface Notification {
  id: string
  user_id: string
  type: string
  title: string
  body: string
  metadata?: Record<string, unknown>
  read_at?: string
  created_at: string
}

export type ActionState =
  | { success: true; message?: string }
  | { success: false; error: string }
  | null
