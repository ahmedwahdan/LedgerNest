-- Enable PostgreSQL extensions used throughout the schema.

-- uuid_generate_v4() for primary keys
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- pg_trgm: trigram similarity for fuzzy merchant name matching
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- citext: case-insensitive text type (used for email columns)
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TYPE household_role AS ENUM ('owner', 'editor', 'viewer');
CREATE TYPE invitation_status AS ENUM ('pending', 'accepted', 'revoked');
CREATE TYPE budget_scope AS ENUM ('personal', 'household');
CREATE TYPE cycle_status AS ENUM ('open', 'closed');
CREATE TYPE audit_action AS ENUM ('create', 'update', 'delete', 'restore');

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email CITEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  display_name TEXT NOT NULL,
  avatar_url TEXT,
  preferred_currency CHAR(3) NOT NULL DEFAULT 'USD',
  verified_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_sessions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  refresh_token_hash TEXT NOT NULL,
  device_info TEXT,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE password_reset_tokens (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE households (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL,
  created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE household_members (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  household_id UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,
  role household_role NOT NULL,
  joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, household_id)
);

CREATE TABLE invitations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email CITEXT NOT NULL,
  household_id UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,
  role household_role NOT NULL,
  token_hash TEXT NOT NULL,
  status invitation_status NOT NULL DEFAULT 'pending',
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE categories (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  household_id UUID REFERENCES households(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
  icon TEXT,
  color TEXT,
  is_system BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE NULLS NOT DISTINCT (household_id, name)
);

CREATE TABLE budget_cycle_configs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  household_id UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,
  start_day INTEGER NOT NULL CHECK (start_day BETWEEN 1 AND 28),
  effective_from DATE NOT NULL,
  created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cycle_snapshots (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  household_id UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,
  cycle_start DATE NOT NULL,
  cycle_end DATE NOT NULL,
  label TEXT NOT NULL,
  status cycle_status NOT NULL DEFAULT 'open',
  config_id UUID NOT NULL REFERENCES budget_cycle_configs(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (cycle_end >= cycle_start),
  UNIQUE (household_id, cycle_start, cycle_end)
);

CREATE TABLE budgets (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  scope budget_scope NOT NULL,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  household_id UUID REFERENCES households(id) ON DELETE CASCADE,
  category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
  cycle_snapshot_id UUID NOT NULL REFERENCES cycle_snapshots(id) ON DELETE CASCADE,
  amount NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
  rollover_amount NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (rollover_amount >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (
    (scope = 'personal' AND user_id IS NOT NULL AND household_id IS NULL) OR
    (scope = 'household' AND household_id IS NOT NULL AND user_id IS NULL)
  )
);

CREATE TABLE expenses (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  scope budget_scope NOT NULL,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  household_id UUID REFERENCES households(id) ON DELETE CASCADE,
  created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
  deleted_by UUID REFERENCES users(id) ON DELETE SET NULL,
  amount NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
  currency CHAR(3) NOT NULL DEFAULT 'USD',
  merchant TEXT NOT NULL,
  category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
  payment_method TEXT NOT NULL,
  date DATE NOT NULL,
  notes TEXT,
  is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
  recurrence_interval TEXT,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  deleted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (
    (scope = 'personal' AND user_id IS NOT NULL AND household_id IS NULL) OR
    (scope = 'household' AND household_id IS NOT NULL AND user_id IS NULL)
  )
);

CREATE TABLE audit_log (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  action audit_action NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id UUID NOT NULL,
  old_values JSONB,
  new_values JSONB,
  ip_address INET,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE notifications (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
  read_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX user_sessions_refresh_token_hash_key ON user_sessions (refresh_token_hash);
CREATE UNIQUE INDEX password_reset_tokens_token_hash_key ON password_reset_tokens (token_hash);
CREATE UNIQUE INDEX invitations_token_hash_key ON invitations (token_hash);
CREATE UNIQUE INDEX invitations_pending_email_household_key
  ON invitations (household_id, email)
  WHERE status = 'pending';
CREATE UNIQUE INDEX budget_cycle_configs_household_effective_from_key
  ON budget_cycle_configs (household_id, effective_from);
CREATE UNIQUE INDEX budgets_scope_target_category_cycle_key
  ON budgets (scope, user_id, household_id, category_id, cycle_snapshot_id);

CREATE INDEX user_sessions_user_id_idx ON user_sessions (user_id);
CREATE INDEX password_reset_tokens_user_id_idx ON password_reset_tokens (user_id);
CREATE INDEX household_members_household_id_idx ON household_members (household_id);
CREATE INDEX invitations_household_id_idx ON invitations (household_id);
CREATE INDEX categories_household_id_idx ON categories (household_id);
CREATE INDEX categories_parent_id_idx ON categories (parent_id);
CREATE INDEX cycle_snapshots_household_id_idx ON cycle_snapshots (household_id);
CREATE INDEX budgets_cycle_snapshot_id_idx ON budgets (cycle_snapshot_id);
CREATE INDEX budgets_household_id_idx ON budgets (household_id);
CREATE INDEX budgets_user_id_idx ON budgets (user_id);
CREATE INDEX expenses_household_date_idx ON expenses (household_id, date DESC);
CREATE INDEX expenses_user_date_idx ON expenses (user_id, date DESC);
CREATE INDEX expenses_category_id_idx ON expenses (category_id);
CREATE INDEX expenses_merchant_trgm_idx ON expenses USING GIN (merchant gin_trgm_ops);
CREATE INDEX audit_log_entity_idx ON audit_log (entity_type, entity_id);
CREATE INDEX audit_log_user_id_idx ON audit_log (user_id);
CREATE INDEX notifications_user_id_idx ON notifications (user_id, created_at DESC);
