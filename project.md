# LedgerNest — Feature Spec & Architecture (MVP)

## Current Implementation Progress

**Phase 1 backend complete.** All Phase 3 (frontend) work is next.

Implemented:
- Auth: register, login, logout, refresh, `GET /auth/me`; bearer-token middleware + refresh-token sessions
- Categories: system seed (11 top-level + 27 subcategories via migration 000002), full CRUD
- Households: create/get/update/delete, member management (roles: owner/editor/viewer), invitations (create/list/revoke/accept), leave
- Budget cycles: `PUT|GET /households/:id/cycle` (start_day config, lazy snapshot creation), `GET .../cycle/snapshots`
- Budgets: full CRUD, `GET /budgets/health` (per-category + overall cap, pct_used, remaining)
- Personal expenses: full CRUD, soft-delete + restore, pagination (limit/offset), filters (from/to/merchant/category_id), `GET /expenses/:id/history`
- Audit log: best-effort recording of create/update/delete/restore for expenses; `ListByEntity` used by history endpoint

Not implemented yet:
- Email verification, forgot/reset password
- Analytics, notifications, CSV export
- Frontend (Phase 3)

## 1. Authentication & User Management

### Login / Registration
- Email + password registration with email verification
- Forgot password / reset flow
- Session management with JWT + refresh tokens
- Remember device option

### User Profile
- Display name, avatar, preferred currency
- Notification preferences
- Personal settings only; household budget cycle is managed at the household level
- Account deletion (GDPR compliant)

---

## 2. Core Expense Entry

### Manual Entry
- Amount, date, merchant/store name
- Category (auto-suggested based on merchant history)
- Subcategory (e.g., Groceries → Dairy, Meat, Cleaning Supplies)
- Payment method (cash, card, bank transfer)
- Notes / memo field
- Recurring flag (weekly, biweekly, monthly)

### Multi-User / Household
- Shared household account (you + partner/family)
- Per-user expense attribution ("Who spent this?")
- Shared vs. personal expenses
- Each expense and budget has a clear scope: `personal` or `household`

### Sharing & Collaboration
- Invite others by email
- Role-based access:
  - **Owner**: full control, manage members, delete household
  - **Editor**: add/edit/delete expenses, manage budgets
  - **Viewer**: read-only access to dashboard and analytics
- Pending invitations list with resend/revoke
- Leave household option for non-owners

### Audit Log / Edit History
- Full history of every action: create, edit, delete, restore
- Tracks: who, what changed, old value → new value, timestamp
- Viewable per expense ("See history" on any entry)
- Global activity feed on dashboard ("Sarah added $45 Groceries — 2 hrs ago")
- Deleted expenses are soft-deleted and recoverable for 30 days

---

## 3. Budget Management

### Custom Budget Cycle
- Define your own monthly cycle start date (e.g., 1st, 15th, 25th — aligned to payday)
- All budget calculations, analytics, and reports respect this cycle
- Example: cycle 25th → your "April budget" runs March 25 – April 24
- Configurable per household (everyone shares the same cycle)
- Cycle changes apply only to future cycles; closed cycles remain immutable for auditability

### Budget Setup
- Set monthly budgets per category (e.g., Groceries: $500, Dining Out: $200)
- Set overall monthly budget cap
- Rollover option: unspent budget carries to next cycle

### Budget Health Dashboard
- Real-time budget usage bars (spent vs. remaining)
- Pace indicator: "You're spending $18/day on groceries — at this rate you'll hit $540 by month-end"
- Days remaining vs. budget remaining ratio
- Color-coded status (green / yellow / red)

---

## 4. Analytics

### Spending Analysis
- Category breakdown (pie/bar charts)
- Monthly trends (time series)
- Merchant ranking (where you spend most)
- Month-over-month comparison

### Basic Export
- CSV export of expenses for a given date range

---

## 5. Notifications & Alerts

- Budget threshold alerts (50%, 75%, 90%, 100%)
- Weekly spending digest

---

## 6. Data Quality & UX

### Data Quality
- Duplicate detection surfaced as a suggestion, not automatic deletion; confidence based on amount + merchant + date + payment method
- Merchant name normalization ("AH Amsterdam" and "Albert Heijn" → same merchant)
- Auto-categorization that learns from corrections

### UX Essentials
- Responsive web app — mobile-friendly, works well on phones without a native app
- Dark mode
- Onboarding wizard: household setup, budget cycle, first budgets

---

## 7. Technical Architecture

### Backend

```
Tech Stack:
├── Language:       Go
├── Router:         Chi (lightweight, idiomatic)
├── Database:       PostgreSQL
├── Migrations:     golang-migrate
├── SQL:            sqlc (type-safe query generation)
├── Auth:           JWT + refresh tokens (golang-jwt/jwt)
├── Validation:     go-playground/validator
├── Config:         godotenv + viper
└── API:            REST (JSON)
```

#### Project Structure
```
/cmd/api              — main entry point
/internal
  /auth               — JWT signing, token validation, session handling
  /handler            — HTTP handlers, grouped by domain
  /middleware         — auth, CORS, request logging, rate-limit
  /service            — business logic
  /repository         — DB access (sqlc-generated + hand-written)
  /model              — domain types
  /validator          — request struct validation
/db
  /migrations         — SQL migration files (up + down)
  /queries            — sqlc query definitions (.sql)
/config               — app configuration structs
```

#### Core Database Tables
```sql
-- Auth
users (
  id, email, password_hash, display_name, avatar_url,
  preferred_currency, verified_at, created_at, updated_at
)
user_sessions (
  id, user_id, refresh_token_hash, device_info, expires_at, created_at
)
password_reset_tokens (
  id, user_id, token_hash, expires_at, used_at
)

-- Households
households (
  id, name, created_by, created_at, updated_at
)
household_members (
  id, user_id, household_id, role: owner/editor/viewer, joined_at
)
invitations (
  id, email, household_id, role, token_hash,
  status: pending/accepted/revoked, expires_at, created_at
)

-- Categories
categories (
  id, household_id nullable, name, parent_id nullable,
  icon, color, is_system, created_at
)
-- system categories are seeded globally (household_id = null);
-- households can add custom categories on top

-- Budget Cycles
budget_cycle_configs (
  id, household_id, start_day, effective_from, created_by, created_at
)
-- one active config per household; old configs retained for history
cycle_snapshots (
  id, household_id, cycle_start, cycle_end, label,
  status: open/closed, config_id, created_at
)
-- one row per billing period; closed once the period ends

-- Budgets
budgets (
  id, scope: personal/household,
  user_id nullable, household_id nullable,   -- exactly one set, driven by scope
  category_id nullable,                      -- null = overall cap
  cycle_snapshot_id, amount, rollover_amount,
  created_at, updated_at
)

-- Expenses
expenses (
  id, scope: personal/household,
  user_id nullable, household_id nullable,   -- exactly one set, driven by scope
  created_by, updated_by, deleted_by,
  amount, currency, merchant, category_id,
  payment_method, date, notes,
  is_recurring, recurrence_interval nullable,
  is_deleted, deleted_at, created_at, updated_at
)

-- Audit & Notifications
audit_log (
  id, user_id, action: create/update/delete/restore,
  entity_type, entity_id, old_values jsonb, new_values jsonb,
  ip_address, created_at
)
notifications (
  id, user_id, type, title, body, metadata jsonb,
  read_at nullable, created_at
)
```

#### Core API Endpoints
```
Auth:
  POST   /auth/register                 [implemented]
  POST   /auth/login                    [implemented]
  POST   /auth/logout                   [implemented]
  POST   /auth/refresh                  [implemented]
  POST   /auth/verify-email
  POST   /auth/forgot-password
  POST   /auth/reset-password

Profile:
  GET    /auth/me                       [implemented]
  PUT    /me
  DELETE /me

Households:
  POST   /households                    [implemented]
  GET    /households/:id                [implemented]
  PUT    /households/:id                [implemented]
  DELETE /households/:id                [implemented]
  POST   /households/:id/leave          [implemented]
  GET    /households/:id/members        [implemented]
  PUT    /households/:id/members/:userId/role  [implemented]
  DELETE /households/:id/members/:userId       [implemented]
  POST   /households/:id/invitations    [implemented]
  GET    /households/:id/invitations    [implemented]
  DELETE /households/:id/invitations/:invId    [implemented]
  POST   /invitations/accept            [implemented]

Budget Cycles:
  GET    /households/:id/cycle          [implemented]
  PUT    /households/:id/cycle          [implemented]
  GET    /households/:id/cycle/snapshots [implemented]

Categories:
  GET    /categories                    [implemented] (system + household custom)
  POST   /categories                    [implemented]
  PUT    /categories/:id                [implemented]
  DELETE /categories/:id                [implemented]

Budgets:
  GET    /budgets                       [implemented] (?snapshot_id=, ?scope=)
  POST   /budgets                       [implemented]
  PUT    /budgets/:id                   [implemented]
  DELETE /budgets/:id                   [implemented]
  GET    /budgets/health                [implemented] (current cycle health summary)

Expenses:
  GET    /expenses                      [implemented] (personal scope; from/to/merchant/category_id/limit/offset)
  POST   /expenses                      [implemented] (personal scope)
  GET    /expenses/:id                  [implemented] (personal scope)
  PUT    /expenses/:id                  [implemented] (personal scope)
  DELETE /expenses/:id                  [implemented] (soft-delete)
  POST   /expenses/:id/restore          [implemented]
  GET    /expenses/:id/history          [implemented]

Activity:
  GET    /activity                       (?from=, ?to=, ?user_id=, ?entity_type=, ?page=, ?limit=)

Analytics:
  GET    /analytics/spending             (?from=, ?to=, ?scope=)
  GET    /analytics/trends              (?months=12, ?scope=)
  GET    /analytics/merchants            (?from=, ?to=)

Reports:
  GET    /reports/export                 (?from=, ?to=, ?format=csv)

Notifications:
  GET    /notifications
  PUT    /notifications/:id/read
  PUT    /notifications/read-all
```

### Web Frontend

```
Tech Stack:
├── Framework:      Next.js (React) with TypeScript
├── Styling:        Tailwind CSS + shadcn/ui
├── Server State:   TanStack Query (React Query)
├── UI State:       Zustand
├── Charts:         Recharts
├── Forms:          React Hook Form + Zod
└── Auth:           JWT/refresh tokens from backend; stored in httpOnly cookies
```

The web app is built mobile-first and responsive. It serves as the primary interface for both desktop and mobile browsers until a native app is built in a later stage.

---

## 8. Build Order (MVP)

### Phase 1 — Backend Foundation
- [x] Go project setup: module, folder structure, config, middleware, error handling
- [x] Database schema + migrations (golang-migrate)
- [ ] sqlc setup + query definitions
- [x] Auth: register, login, logout, refresh (`GET /auth/me`); email verification + password reset deferred post-MVP
- [x] Household: create, settings, invite, manage members and roles
- [x] Budget cycle config + snapshot generation (lazy creation on GET /cycle)
- [x] Seed default category tree (migration 000002: 11 top-level + 27 subcategories)
- [x] CRUD: categories, personal expenses (soft-delete + restore + history + pagination), budgets
- [x] Budget health calculations (spent vs. remaining per category and overall cap)
- [x] Audit log: best-effort recording on expense mutations; history endpoint
- [ ] API tests for all endpoints

### Phase 2 — Analytics & Notifications Backend
- [ ] Analytics: spending breakdown by category, monthly trends, merchant ranking
- [ ] Budget threshold calculation + notification generation (50%, 75%, 90%, 100%)
- [ ] Weekly digest notification job
- [ ] Notification read/unread endpoints

### Phase 3 — Web Frontend
- [x] Auth pages: login (wired to API + httpOnly cookie session), register
- [x] Route protection via proxy.ts (redirect to /login if no access_token cookie)
- [x] App shell: sidebar navigation (dashboard, expenses, budgets), logout
- [x] Dashboard: budget health bars (pct_used, remaining, color-coded), recent expenses
- [x] Expense entry form (amount, currency, merchant, date, payment method, category, notes)
- [x] Expense list with filters (from/to date, merchant search)
- [x] Expense detail: inline edit, delete (soft), restore, audit history timeline
- [x] Budget management: health view (overall cap + per-category bars), add/remove budgets
- [ ] Onboarding wizard: household setup, budget cycle config
- [ ] Household management pages (members, invitations, roles)
- [ ] Analytics pages: category breakdown, trends, merchant ranking
- [ ] Notifications panel
- [ ] CSV export
- [ ] Dark mode + mobile-responsive polish throughout

---

## 9. Later Stages (Post-MVP)

### Near-Term
- OAuth login (Google, Apple)
- Receipt scanning — Claude API (vision) + S3/MinIO storage + async queue (asynq + Redis)
- Budget borrowing — advance from next cycle with tracking and alerts
- Per-category permission overrides
- Activity feed filters (by user, action type, category, date)
- Audit log export
- Monthly summary report (PDF + email delivery)

### Medium-Term
- Consumption tracking — item-level purchase history, usage rates, low-stock predictions, smart shopping list
- "Why is my budget failing?" AI insights (Claude API)
- Spending forecast model
- Bank & card import — CSV/OFX; ABN AMRO PSD2/Open Banking integration
- Transaction review inbox (match, ignore, split, mark-as-transfer)
- Income tracking + net cash flow view
- Savings goals with timeline and what-if simulator

### Long-Term
- Native mobile app (React Native or Flutter) — camera scanning, push notifications, offline support, quick-add widget
- Split expenses among household members or friends with settle-up flow
- Subscription auto-detection and renewal reminders
- Gamification: budget streaks, household challenges, weekly wins
- Multi-currency support + travel mode
- Warranty and big-purchase tracking
- Tags and custom fields per household
- Webhooks + personal API keys for power users
- Data portability: import from Mint/YNAB, full JSON/CSV export
