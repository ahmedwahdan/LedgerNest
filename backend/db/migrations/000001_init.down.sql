DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS budgets;
DROP TABLE IF EXISTS cycle_snapshots;
DROP TABLE IF EXISTS budget_cycle_configs;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS household_members;
DROP TABLE IF EXISTS households;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS audit_action;
DROP TYPE IF EXISTS cycle_status;
DROP TYPE IF EXISTS budget_scope;
DROP TYPE IF EXISTS invitation_status;
DROP TYPE IF EXISTS household_role;

DROP EXTENSION IF EXISTS "citext";
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "uuid-ossp";
