# Expense Intelligence App — Feature Spec & Architecture

## 1. Authentication & User Management

### Login / Registration
- Email + password registration with email verification
- OAuth login (Google, Apple)
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

### Receipt Scanning (AI-Powered)
- Camera capture or photo upload
- AI extracts: merchant, date, total, tax, individual line items
- Line-item categorization (each item on a receipt gets its own category)
- User can review, correct, and confirm extracted data
- Stores original image linked to the expense

### Multi-User / Household
- Shared household account (you + partner/family)
- Per-user expense attribution ("Who spent this?")
- Shared vs. personal budgets
- Each expense, budget, and reportable object has a clear scope: `personal` or `household`

### Sharing & Collaboration
- Invite others by email or link
- Role-based access:
  - **Owner**: full control, manage members, delete household
  - **Editor**: add/edit/delete expenses, manage budgets
  - **Viewer**: read-only access to dashboard, analytics, reports
- Pending invitations list with resend/revoke
- Per-category permissions (optional): explicit overrides on top of the base household role, e.g., editor for Groceries but viewer for everything else
- Leave household option for non-owners

### Audit Log / Edit History
- Full history of every action: create, edit, delete
- Tracks: who, what changed, old value → new value, timestamp
- Viewable per expense ("See history" on any entry)
- Global activity feed on dashboard ("Sarah added $45 Groceries — 2 hrs ago")
- Filter activity by user, action type, category, date range
- Deleted expenses are soft-deleted and recoverable for 30 days
- Export audit log for record-keeping


---

## 2. Budget Management

### Custom Budget Cycle
- Define your own monthly cycle start date (e.g., 1st, 15th, 25th — aligned to payday)
- All budget calculations, analytics, and reports respect this cycle
- Example: cycle 25th → your "April budget" runs March 25 – April 24
- Configurable per household (everyone shares the same cycle)
- Cycle changes apply only to future cycles; closed cycles remain immutable for auditability and reporting
- Existing reports, borrowings, refunds, and savings allocations stay linked to the cycle snapshot in effect when they were created

### Budget Setup
- Set monthly budgets per category (e.g., Groceries: $500, Dining Out: $200)
- Set overall monthly budget cap
- Rollover option (unspent budget carries to next cycle)
- Seasonal budgets (e.g., higher utility budget in winter)

### Budget Borrowing (Advance from Next Month)
- "Borrow" from next cycle's budget for a specific category or overall
- Tracks borrowed amount separately — next month starts with reduced budget
- Borrowing log with reason/note ("Car repair — emergency")
- Dashboard shows: original budget, borrowed amount, effective remaining
- Alerts: "You borrowed $150 from May — your May grocery budget is now $350"
- Borrowing limits (optional): cap how much you can borrow (e.g., max 30% of next month)
- Repayment tracking: if you underspend next month, auto-repays the borrowed amount
- Chain prevention: warning if you try to borrow from a month that already has debt

### Budget Health Dashboard
- Real-time budget usage bars (spent vs. remaining)
- Pace indicator: "You're spending $18/day on groceries — at this rate you'll hit $540 by month-end"
- Days remaining vs. budget remaining ratio
- Color-coded alerts (green / yellow / red)

### "Why Is My Budget Failing?" — AI Insights
- **Spike detection**: "You spent $85 at Costco on March 12 — that's 3x your usual visit"
- **Category drift**: "Your 'Groceries' budget is being hit by cleaning supplies ($67 this month). Consider a separate Household budget."
- **Frequency analysis**: "You're shopping 5x/week instead of your usual 2x — more trips = more impulse buys"
- **Price increase alerts**: "Milk averaged $3.20 last quarter, now averaging $3.80 — that's +$7/month"
- **Comparison**: "March groceries were 22% higher than your 3-month average. Top contributors: snacks (+$34), beverages (+$21)"
- **Actionable suggestions**: "Switching one dining-out meal per week to home cooking could save ~$120/month"


---

## 3. Consumption Tracking

### Item-Level Tracking
- Track purchase frequency of specific items (toilet paper, detergent, coffee, etc.)
- Auto-detected from receipt line items or manually logged
- Unit tracking where relevant (rolls, kg, liters, packs)

### Consumption Analytics
- **Usage rate**: "You buy toilet paper every 3.2 weeks on average"
- **Projected needs**: "You'll need toilet paper around April 18 — add to shopping list?"
- **Cost per unit over time**: "Toilet paper cost/roll: $0.52 avg — best price was $0.38 at Costco"
- **Smart shopping list**: Auto-generated list of items likely needed soon
- **Bulk buy analysis**: "Buying 24-pack at Costco saves $4.20 vs. 6-packs at Albert Heijn"

### Inventory Awareness
- Optional "I bought this" confirmation to reset countdown
- Low-stock predictions based on usage patterns


---

## 4. Analytics & Reporting

### Spending Analysis
- Category breakdown (pie, bar, treemap charts)
- Time series (daily, weekly, monthly, yearly trends)
- Merchant ranking (where you spend most)
- Day-of-week patterns ("You spend 40% more on weekends")
- Compare month-over-month, year-over-year

### Forecasting
- Predicted end-of-month spend per category
- ML-based forecast using historical patterns + seasonality
- "What-if" simulator: adjust a habit and see projected savings
- Annual projection based on current trends
- Savings goal tracking with timeline estimates

### Reports
- Monthly summary report (PDF/email)
- Tax-ready export (filter by deductible categories)
- CSV / Excel export
- Shareable household spending report


---

## 5. Notifications & Alerts

- Budget threshold alerts (50%, 75%, 90%, 100%)
- Unusual spending alerts ("$200 at a store you've never visited")
- Bill reminders (recurring expenses due soon)
- Consumption reminders ("Time to restock laundry detergent")
- Weekly spending digest
- Soft transaction review reminders for newly imported bank transactions that are still uncategorized or not converted into expense entries
- Repeated non-blocking nudges until each imported transaction is resolved: create expense, match to existing expense, ignore, or mark as transfer
- Custom alert rules


---

## 6. Additional Important Features

### Data Quality
- Duplicate detection as a suggestion, not automatic deletion; confidence uses amount + merchant + date/time + payment method + proximity to existing entries
- Merchant name normalization ("AH Amsterdam" and "Albert Heijn" → same merchant)
- Auto-categorization that learns from your corrections

### Security & Privacy
- Encryption in transit and at rest for financial data
- Biometric login (fingerprint / face) on mobile
- PIN protection
- Data export / account deletion (GDPR compliance)

### UX Essentials
- Quick-add widget (mobile home screen)
- Offline support with sync
- Dark mode
- Multi-language support
- Onboarding wizard with budget setup


---

## 9. Good-to-Have / Future Features

### Income & Cash Flow
- Track income sources (salary, freelance, side income)
- Net cash flow view: income minus expenses per cycle
- Income vs. expense trend chart
- Financial calendar: see upcoming bills, expected income, due dates on a calendar

### Target Savings Goals
- Create named goals with a target amount (e.g., "New Car — $15,000", "Power Drill — $250")
- Set priority level (high / medium / low) when multiple goals exist
- Deadline option: "I want this by December 2026"

#### Connected to Your Full Financial Picture
- **Auto-calculated surplus**: income minus (expenses + budget commitments) = available to save
- **Realistic timeline**: "Based on your average surplus of $420/month, you'll reach $15,000 in 36 months (July 2029)"
- **Impact of expenses on goals**: "Your grocery overspend this month delayed your car goal by 12 days"
- **What-if simulator**: "If you cut dining out by $100/month, you'll reach your goal 4 months sooner"
- **Goal vs. budget tension view**: see how your budgets and goals compete for the same surplus

#### Funding & Progress
- Manual contributions ("I set aside $200 this month")
- Auto-allocate: automatically assign a fixed amount or % of surplus each cycle
- Visual progress bar with milestone markers (25%, 50%, 75%)
- Celebration moments when milestones are hit
- Pause/resume a goal without losing progress

#### Multi-Goal Management
- Dashboard showing all active goals ranked by priority
- Smart allocation: distribute surplus across goals by priority or custom split
- "Quick win" suggestions: "You're $38 away from your Power Drill goal — skip one takeout this week?"
- Rebalance wizard: if income changes, recalculate all goal timelines at once

#### Notifications
- Monthly progress update per goal
- "On track" / "Falling behind" status with specific reasons
- Deadline warnings: "At current pace, you'll miss your December deadline by 2 months"
- Opportunity alerts: "You underspent by $85 this cycle — move it to a goal?"

### Split Expenses
- Split a single expense among household members or friends
- Even split, percentage split, or custom amounts
- Track who owes whom with running balances
- Settle up flow ("Mark as paid")
- Useful for shared dinners, trips, household purchases

### Subscription Tracker
- Auto-detect recurring charges (Netflix, Spotify, gym, insurance)
- Monthly/annual subscription summary with total cost
- Renewal reminders and cancellation tracking
- "Forgotten subscriptions" alert for things you haven't used
- Annual cost view: "You spend $2,340/year on subscriptions"

### Smart Shopping List
- Auto-generated from consumption predictions ("Running low on coffee")
- Manual items with budget estimates
- Suggested store based on best historical prices
- Shareable with household members in real-time
- Check off items while shopping → auto-log expense

### Bank & Card Import
- CSV/OFX import from bank statements
- Auto-match imported transactions to existing manual entries (dedup)
- ABN AMRO bank connection as a first-class integration
- Live bank feeds via PSD2/Open Banking provider or direct bank integration where available
- Categorize imported transactions in bulk
- Imported transactions land in a review inbox before becoming final expenses
- For each imported transaction, user can: create expense entry, match to an existing expense, split, ignore, or mark as transfer
- The app keeps surfacing unresolved imported transactions in dashboard, notifications, and expense-entry flows until resolved, but does not block normal app usage
- Merchant and category suggestions are prefilled from transaction description and prior history

### Refund & Return Tracking
- Mark an expense as returned/refunded (partial or full)
- Refund restores budget for that cycle
- Tracks pending refunds vs. completed
- Linked to original expense entry

### Warranty & Big Purchase Tracking
- Tag high-value purchases with warranty expiration
- Upload warranty documents / proof of purchase
- Reminder before warranty expires
- Useful for electronics, appliances, furniture

### Tags & Custom Fields
- Flexible tags in addition to categories (e.g., "vacation", "birthday", "tax-deductible")
- Filter and report by tags across categories
- Custom fields per household (e.g., "project", "client" for freelancers)

### Gamification & Motivation
- Monthly budget streaks ("4 months under budget on Dining!")
- Savings milestones with visual progress
- Household challenges ("Can we cut takeout by 20% this month?")
- Weekly "wins" summary highlighting positive trends

### Localization
- Multi-currency support with auto-conversion
- Travel mode: temporarily switch default currency
- Date/number formatting per locale
- RTL language support

### Data Portability
- Import from other apps (Mint, YNAB, Excel)
- Full data export (JSON, CSV)
- API access with personal API keys for power users
- Webhooks for custom integrations (e.g., log expense → trigger automation)


---

## 7. Technical Architecture

### Backend (Build First)

```
Tech Stack:
├── Language:       Node.js (TypeScript) or Python (FastAPI)
├── Database:       PostgreSQL (relational data) + S3/MinIO (receipt images)
├── ORM:            Prisma (Node) or SQLAlchemy (Python)
├── Auth:           JWT + refresh tokens, OAuth2 (Google/Apple login)
├── AI/ML:          Claude API (receipt OCR + insights) + basic forecasting models
├── Queue:          Redis + BullMQ (async receipt processing)
├── Cache:          Redis (dashboard data, session)
└── API:            REST or GraphQL
```

#### Key Database Tables
```
users
households
household_members (user_id, household_id, role: owner/editor/viewer)
invitations (email, household_id, role, status, token, expires_at)
permission_overrides (household_id, subject_type, subject_id, member_id, role)
categories (hierarchical — parent/child)
budgets (per category, per cycle, scope: personal/household, owner_user_id nullable, household_id nullable)
expenses (+ scope: personal/household, owner_user_id nullable, household_id nullable, created_by, updated_by, deleted_by, is_deleted)
expense_items (line items from receipts)
receipts (image URL, raw extracted data, status)
products (normalized product catalog)
purchase_history (links expenses to products for consumption tracking)
bank_connections (user_id, provider, institution_name, consent_expires_at, status, last_synced_at)
bank_accounts (bank_connection_id, external_account_id, iban_masked, account_name, currency)
bank_transactions (bank_account_id, external_transaction_id, booked_at, amount, currency, description, counterparty, status: new/reviewed/matched/ignored/transfer)
transaction_matches (bank_transaction_id, expense_id, match_type: auto/manual, confidence)
budget_cycles (household_id, start_day, effective_from, effective_to, current_cycle_start, current_cycle_end)
cycle_snapshots (budget_cycle_id, cycle_start, cycle_end, label, status: open/closed)
budget_borrowings (from_cycle_snapshot_id, to_cycle_snapshot_id, category, amount, reason, repaid_amount)
audit_log (user_id, action, entity_type, entity_id, old_values, new_values, timestamp)
income_sources (user_id, name, amount, frequency, next_date)
savings_goals (name, target_amount, deadline, priority, current_amount, status, auto_allocate_amount)
savings_contributions (goal_id, amount, source: manual/auto/surplus, date)
recurring_expenses
notifications
```

#### Core API Endpoints
```
Auth:       POST /auth/register, /auth/login, /auth/refresh, /auth/forgot-password
            POST /auth/oauth/google, /auth/oauth/apple
Expenses:   CRUD /expenses, POST /expenses/receipt-scan
            GET  /expenses/:id/history (audit trail per expense)
Budgets:    CRUD /budgets, GET /budgets/health
            POST /budgets/borrow, GET /budgets/borrowings
Cycle:      GET/PUT /settings/budget-cycle
Categories: CRUD /categories
Analytics:  GET /analytics/spending, /analytics/trends, /analytics/forecast
Consumption:GET /consumption/items, /consumption/predictions
Reports:    GET /reports/monthly, /reports/export
Household:  POST /household/create, PUT /household/settings
            POST /household/invite, DELETE /household/invite/:id
            GET  /household/members, PUT /household/members/:id/role
            DELETE /household/members/:id
Banking:    POST /banking/connections/abn-amro
            GET  /banking/accounts, /banking/transactions
            POST /banking/transactions/:id/create-expense
            POST /banking/transactions/:id/match-expense
            POST /banking/transactions/:id/ignore
            POST /banking/transactions/:id/mark-transfer
            POST /banking/sync
Activity:   GET /activity/feed (global audit feed, filterable)
            GET /activity/export
Income:     CRUD /income, GET /income/summary
Savings:    CRUD /savings/goals, POST /savings/goals/:id/contribute
            GET  /savings/goals/:id/timeline
            GET  /savings/surplus (available to save this cycle)
            POST /savings/rebalance
            GET  /savings/what-if?cut_category=dining&amount=100
```

#### Receipt Processing Pipeline
```
1. User uploads image → stored in S3
2. Job queued for async processing
3. Claude Vision API extracts merchant, date, total, line items
4. Items matched to product catalog (fuzzy matching)
5. Auto-categorized based on history
6. User reviews & confirms in app
7. Expense + line items saved to DB
```

### Web Frontend

```
Tech Stack:
├── Framework:      Next.js (React) with TypeScript
├── State:          Zustand or React Query
├── Charts:         Recharts or Chart.js
├── UI Library:     shadcn/ui + Tailwind CSS
├── Forms:          React Hook Form + Zod validation
└── Auth:           Backend-issued JWT/refresh tokens; frontend consumes backend auth APIs
```

### Mobile Frontend

```
Tech Stack (choose one):
├── Option A:       React Native (Expo) — shares logic with web
├── Option B:       Flutter — great performance, single codebase
└── Option C:       PWA — wrap the Next.js app, least effort

Mobile-Specific Features:
├── Camera integration for receipt scanning
├── Push notifications
├── Home screen quick-add widget
├── Offline-first with background sync
└── Biometric authentication
```


---

## 8. Suggested Build Order

### Phase 1 — Foundation
- [ ] Database schema + migrations
- [ ] Auth system (register, login, JWT + refresh tokens)
- [ ] Household model, membership, scoped permissions, and budget-cycle snapshots
- [ ] CRUD for expenses, categories, budgets with personal vs. household scope
- [ ] Budget cycle configuration + borrowing logic
- [ ] Basic API tests

### Phase 2 — Intelligence
- [ ] Receipt upload + Claude Vision integration
- [ ] Auto-categorization engine
- [ ] Budget tracking + health calculations
- [ ] Basic analytics endpoints

### Phase 3 — Web App
- [ ] Dashboard (budget bars, recent expenses, spending chart)
- [ ] Expense entry form + receipt upload UI
- [ ] Budget setup & monitoring screens (cycle config, borrowing UI)
- [ ] Analytics pages with charts

### Phase 4 — Consumption & Forecasting
- [ ] Product catalog + consumption tracking
- [ ] Usage rate calculations + predictions
- [ ] Spending forecast model
- [ ] "Why is my budget failing?" insights engine

### Phase 5 — Mobile App
- [ ] Camera receipt scanning
- [ ] Push notifications
- [ ] Offline support
- [ ] Quick-add widget

### Phase 6 — Polish
- [ ] Reports & exports
- [ ] Onboarding flow
- [ ] Performance optimization & security audit
