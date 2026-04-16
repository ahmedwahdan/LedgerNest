-- ============================================================
-- LedgerNest — Showcase Seed Data
-- Primary demo user: Ahmed (ahmedibrahimwahdan@gmail.com)
-- Secondary household: Smith Family (alice@example.com + bob@example.com)
-- ============================================================

BEGIN;

-- ── 0. Clean up Ahmed's orphan test households ────────────────────────────────
DELETE FROM households
WHERE created_by = '6aa95384-2a95-4887-a0f5-e59d4a7c64c2'
  AND id IN (
    '40f932b4-e83a-42d6-a04e-6d7281a09943',
    '5df775a8-fb77-4ba6-bd43-1d4ba1c3ccca',
    'cf302045-4310-4541-9bb6-ba888165ba35'
  );

-- ── 1. Wahdan Home household ──────────────────────────────────────────────────
INSERT INTO households (id, name, created_by, created_at, updated_at)
VALUES ('aa000001-0000-0000-0000-000000000001', 'Wahdan Home',
        '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
        NOW() - INTERVAL '90 days', NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO household_members (id, user_id, household_id, role, joined_at)
VALUES ('aa000001-0000-0000-0000-000000000002',
        '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
        'aa000001-0000-0000-0000-000000000001', 'owner',
        NOW() - INTERVAL '90 days')
ON CONFLICT (user_id, household_id) DO NOTHING;

-- ── 2. Budget cycle config (start_day=1) ─────────────────────────────────────
INSERT INTO budget_cycle_configs (id, household_id, start_day, effective_from, created_by, created_at)
VALUES ('bc000001-0000-0000-0000-000000000001',
        'aa000001-0000-0000-0000-000000000001',
        1, '2026-02-01',
        '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', NOW() - INTERVAL '75 days')
ON CONFLICT DO NOTHING;

-- ── 3. Cycle snapshots: Feb (closed), Mar (closed), Apr (open) ────────────────
INSERT INTO cycle_snapshots (id, household_id, cycle_start, cycle_end, label, status, config_id, created_at)
VALUES
  ('c5000001-0000-0000-0000-000000000001',
   'aa000001-0000-0000-0000-000000000001',
   '2026-02-01', '2026-02-28', 'February 2026', 'closed',
   'bc000001-0000-0000-0000-000000000001', '2026-02-01'),
  ('c5000001-0000-0000-0000-000000000002',
   'aa000001-0000-0000-0000-000000000001',
   '2026-03-01', '2026-03-31', 'March 2026', 'closed',
   'bc000001-0000-0000-0000-000000000001', '2026-03-01'),
  ('c5000001-0000-0000-0000-000000000003',
   'aa000001-0000-0000-0000-000000000001',
   '2026-04-01', '2026-04-30', 'April 2026', 'open',
   'bc000001-0000-0000-0000-000000000001', '2026-04-01')
ON CONFLICT (household_id, cycle_start, cycle_end) DO NOTHING;

-- ── 4. Personal budgets for Ahmed — April snapshot ────────────────────────────
INSERT INTO budgets (id, scope, user_id, cycle_snapshot_id, category_id, amount, rollover_amount)
VALUES
  -- Overall cap
  ('bd000001-0000-0000-0000-000000000001', 'personal',
   '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   'c5000001-0000-0000-0000-000000000003', NULL, 4000.00, 0),
  -- Food & Dining
  ('bd000001-0000-0000-0000-000000000002', 'personal',
   '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   'c5000001-0000-0000-0000-000000000003',
   'c1000000-0000-0000-0000-000000000001', 700.00, 0),
  -- Transportation
  ('bd000001-0000-0000-0000-000000000003', 'personal',
   '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   'c5000001-0000-0000-0000-000000000003',
   'c1000000-0000-0000-0000-000000000002', 300.00, 0),
  -- Housing
  ('bd000001-0000-0000-0000-000000000004', 'personal',
   '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   'c5000001-0000-0000-0000-000000000003',
   'c1000000-0000-0000-0000-000000000003', 1300.00, 0),
  -- Shopping
  ('bd000001-0000-0000-0000-000000000005', 'personal',
   '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   'c5000001-0000-0000-0000-000000000003',
   'c1000000-0000-0000-0000-000000000005', 500.00, 0),
  -- Entertainment
  ('bd000001-0000-0000-0000-000000000006', 'personal',
   '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   'c5000001-0000-0000-0000-000000000003',
   'c1000000-0000-0000-0000-000000000006', 250.00, 0),
  -- Health & Fitness
  ('bd000001-0000-0000-0000-000000000007', 'personal',
   '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   'c5000001-0000-0000-0000-000000000003',
   'c1000000-0000-0000-0000-000000000004', 200.00, 0)
ON CONFLICT DO NOTHING;

-- ── 5. Ahmed's personal expenses ─────────────────────────────────────────────
-- ── FEBRUARY 2026 ─────────────────────────────────────────────────────────────
INSERT INTO expenses
  (id, scope, user_id, created_by, amount, currency, merchant, category_id, payment_method, date, notes, is_recurring)
VALUES
  -- Housing
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   1200.00, 'EUR', 'Landlord',                    'c1000000-0000-0000-0000-000000000003', 'bank_transfer', '2026-02-01', 'Monthly rent', TRUE),
  -- Food
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   62.40, 'EUR', 'Albert Heijn',                  'c2000000-0000-0000-0000-000000000001', 'card',          '2026-02-02', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   38.90, 'EUR', 'Jumbo',                         'c2000000-0000-0000-0000-000000000001', 'card',          '2026-02-06', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   29.80, 'EUR', 'Lidl',                          'c2000000-0000-0000-0000-000000000001', 'cash',          '2026-02-10', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   55.60, 'EUR', 'Albert Heijn',                  'c2000000-0000-0000-0000-000000000001', 'card',          '2026-02-14', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   47.20, 'EUR', 'Jumbo',                         'c2000000-0000-0000-0000-000000000001', 'card',          '2026-02-20', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   36.10, 'EUR', 'Albert Heijn',                  'c2000000-0000-0000-0000-000000000001', 'card',          '2026-02-25', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   58.00, 'EUR', 'Restaurant De Kas',             'c2000000-0000-0000-0000-000000000002', 'card',          '2026-02-08', 'Valentine dinner', FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   28.50, 'EUR', 'Sumo Amsterdam',                'c2000000-0000-0000-0000-000000000002', 'card',          '2026-02-15', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   22.50, 'EUR', 'Pllek',                         'c2000000-0000-0000-0000-000000000002', 'card',          '2026-02-22', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   9.80, 'EUR', 'Starbucks',                      'c2000000-0000-0000-0000-000000000003', 'card',          '2026-02-03', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   7.60, 'EUR', 'Lot Sixty One',                  'c2000000-0000-0000-0000-000000000003', 'cash',          '2026-02-11', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   11.40, 'EUR', 'Starbucks',                     'c2000000-0000-0000-0000-000000000003', 'card',          '2026-02-18', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   8.20, 'EUR', 'Coffee & Coconuts',              'c2000000-0000-0000-0000-000000000003', 'card',          '2026-02-24', NULL, FALSE),
  -- Transport
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   28.50, 'EUR', 'NS',                            'c2000000-0000-0000-0000-000000000006', 'card',          '2026-02-01', 'Monthly OV top-up', FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   3.80, 'EUR', 'GVB',                            'c2000000-0000-0000-0000-000000000006', 'card',          '2026-02-05', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   62.40, 'EUR', 'Shell',                         'c2000000-0000-0000-0000-000000000005', 'card',          '2026-02-12', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   18.00, 'EUR', 'P+R Amsterdam West',            'c2000000-0000-0000-0000-000000000007', 'cash',          '2026-02-22', NULL, FALSE),
  -- Shopping
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   79.99, 'EUR', 'Zara',                          'c1000000-0000-0000-0000-000000000005', 'card',          '2026-02-07', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   124.99, 'EUR', 'bol.com',                      'c1000000-0000-0000-0000-000000000005', 'card',          '2026-02-16', 'Books and headphones', FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   34.50, 'EUR', 'HEMA',                          'c1000000-0000-0000-0000-000000000005', 'card',          '2026-02-19', NULL, FALSE),
  -- Entertainment
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   15.99, 'EUR', 'Netflix',                       'c1000000-0000-0000-0000-000000000006', 'card',          '2026-02-01', NULL, TRUE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   9.99, 'EUR', 'Spotify',                        'c1000000-0000-0000-0000-000000000006', 'card',          '2026-02-01', NULL, TRUE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   24.50, 'EUR', 'Pathé Cinema',                  'c1000000-0000-0000-0000-000000000006', 'card',          '2026-02-09', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   34.00, 'EUR', 'Bar Spek',                      'c1000000-0000-0000-0000-000000000006', 'cash',          '2026-02-21', 'Drinks with friends', FALSE),
  -- Health
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   60.00, 'EUR', 'David Lloyd',                   'c1000000-0000-0000-0000-000000000004', 'bank_transfer', '2026-02-01', 'Gym membership', TRUE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   18.50, 'EUR', 'Etos Pharmacy',                 'c1000000-0000-0000-0000-000000000004', 'card',          '2026-02-13', NULL, FALSE),

-- ── MARCH 2026 ────────────────────────────────────────────────────────────────
  -- Housing
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   1200.00, 'EUR', 'Landlord',                    'c1000000-0000-0000-0000-000000000003', 'bank_transfer', '2026-03-01', 'Monthly rent', TRUE),
  -- Groceries
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   71.30, 'EUR', 'Albert Heijn',                  'c2000000-0000-0000-0000-000000000001', 'card',          '2026-03-02', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   43.80, 'EUR', 'Jumbo',                         'c2000000-0000-0000-0000-000000000001', 'card',          '2026-03-07', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   58.40, 'EUR', 'Albert Heijn',                  'c2000000-0000-0000-0000-000000000001', 'card',          '2026-03-13', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   34.20, 'EUR', 'Lidl',                          'c2000000-0000-0000-0000-000000000001', 'cash',          '2026-03-17', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   52.60, 'EUR', 'Albert Heijn',                  'c2000000-0000-0000-0000-000000000001', 'card',          '2026-03-23', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   41.90, 'EUR', 'Jumbo',                         'c2000000-0000-0000-0000-000000000001', 'card',          '2026-03-28', NULL, FALSE),
  -- Restaurants & Coffee
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   72.00, 'EUR', 'Restaurant Breda',              'c2000000-0000-0000-0000-000000000002', 'card',          '2026-03-08', 'Birthday dinner', FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   31.00, 'EUR', 'Pllek',                         'c2000000-0000-0000-0000-000000000002', 'card',          '2026-03-19', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   25.50, 'EUR', 'Sumo Amsterdam',                'c2000000-0000-0000-0000-000000000002', 'card',          '2026-03-26', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   10.40, 'EUR', 'Starbucks',                     'c2000000-0000-0000-0000-000000000003', 'card',          '2026-03-04', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   8.80, 'EUR', 'Coffee & Coconuts',              'c2000000-0000-0000-0000-000000000003', 'card',          '2026-03-12', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   7.20, 'EUR', 'Lot Sixty One',                  'c2000000-0000-0000-0000-000000000003', 'cash',          '2026-03-20', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   12.50, 'EUR', 'Starbucks',                     'c2000000-0000-0000-0000-000000000003', 'card',          '2026-03-27', NULL, FALSE),
  -- Transport
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   28.50, 'EUR', 'NS',                            'c2000000-0000-0000-0000-000000000006', 'card',          '2026-03-01', 'Monthly OV top-up', FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   68.90, 'EUR', 'Shell',                         'c2000000-0000-0000-0000-000000000005', 'card',          '2026-03-09', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   71.20, 'EUR', 'BP',                            'c2000000-0000-0000-0000-000000000005', 'card',          '2026-03-22', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   15.00, 'EUR', 'P+R Amsterdam West',            'c2000000-0000-0000-0000-000000000007', 'cash',          '2026-03-15', NULL, FALSE),
  -- Shopping
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   149.99, 'EUR', 'MediaMarkt',                   'c1000000-0000-0000-0000-000000000005', 'card',          '2026-03-06', 'Keyboard', FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   64.50, 'EUR', 'bol.com',                       'c1000000-0000-0000-0000-000000000005', 'card',          '2026-03-14', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   44.99, 'EUR', 'H&M',                           'c1000000-0000-0000-0000-000000000005', 'card',          '2026-03-21', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   22.00, 'EUR', 'HEMA',                          'c1000000-0000-0000-0000-000000000005', 'card',          '2026-03-29', NULL, FALSE),
  -- Entertainment
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   15.99, 'EUR', 'Netflix',                       'c1000000-0000-0000-0000-000000000006', 'card',          '2026-03-01', NULL, TRUE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   9.99, 'EUR', 'Spotify',                        'c1000000-0000-0000-0000-000000000006', 'card',          '2026-03-01', NULL, TRUE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   24.50, 'EUR', 'Pathé Cinema',                  'c1000000-0000-0000-0000-000000000006', 'card',          '2026-03-15', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   29.99, 'EUR', 'Steam',                         'c1000000-0000-0000-0000-000000000006', 'card',          '2026-03-18', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   38.00, 'EUR', 'Bar Spek',                      'c1000000-0000-0000-0000-000000000006', 'cash',          '2026-03-28', NULL, FALSE),
  -- Health
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   60.00, 'EUR', 'David Lloyd',                   'c1000000-0000-0000-0000-000000000004', 'bank_transfer', '2026-03-01', 'Gym membership', TRUE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   34.80, 'EUR', 'Etos Pharmacy',                 'c1000000-0000-0000-0000-000000000004', 'card',          '2026-03-16', 'Vitamins', FALSE),

-- ── APRIL 2026 (current, partial) ────────────────────────────────────────────
  -- Housing
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   1200.00, 'EUR', 'Landlord',                    'c1000000-0000-0000-0000-000000000003', 'bank_transfer', '2026-04-01', 'Monthly rent', TRUE),
  -- Subscriptions (Apr 1)
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   15.99, 'EUR', 'Netflix',                       'c1000000-0000-0000-0000-000000000006', 'card',          '2026-04-01', NULL, TRUE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   9.99, 'EUR', 'Spotify',                        'c1000000-0000-0000-0000-000000000006', 'card',          '2026-04-01', NULL, TRUE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   60.00, 'EUR', 'David Lloyd',                   'c1000000-0000-0000-0000-000000000004', 'bank_transfer', '2026-04-01', 'Gym membership', TRUE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   28.50, 'EUR', 'NS',                            'c2000000-0000-0000-0000-000000000006', 'card',          '2026-04-01', 'Monthly OV top-up', FALSE),
  -- Groceries Apr
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   68.40, 'EUR', 'Albert Heijn',                  'c2000000-0000-0000-0000-000000000001', 'card',          '2026-04-02', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   44.90, 'EUR', 'Jumbo',                         'c2000000-0000-0000-0000-000000000001', 'card',          '2026-04-06', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   51.20, 'EUR', 'Albert Heijn',                  'c2000000-0000-0000-0000-000000000001', 'card',          '2026-04-10', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   33.70, 'EUR', 'Lidl',                          'c2000000-0000-0000-0000-000000000001', 'cash',          '2026-04-13', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   49.30, 'EUR', 'Albert Heijn',                  'c2000000-0000-0000-0000-000000000001', 'card',          '2026-04-16', NULL, FALSE),
  -- Restaurants Apr
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   66.00, 'EUR', 'De Silveren Spiegel',           'c2000000-0000-0000-0000-000000000002', 'card',          '2026-04-04', 'Special occasion', FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   38.50, 'EUR', 'Restaurant Breda',              'c2000000-0000-0000-0000-000000000002', 'card',          '2026-04-11', NULL, FALSE),
  -- Coffee Apr
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   12.50, 'EUR', 'Starbucks',                     'c2000000-0000-0000-0000-000000000003', 'card',          '2026-04-03', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   8.80, 'EUR', 'Lot Sixty One',                  'c2000000-0000-0000-0000-000000000003', 'cash',          '2026-04-07', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   11.90, 'EUR', 'Coffee & Coconuts',             'c2000000-0000-0000-0000-000000000003', 'card',          '2026-04-14', NULL, FALSE),
  -- Transport Apr
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   65.80, 'EUR', 'Shell',                         'c2000000-0000-0000-0000-000000000005', 'card',          '2026-04-05', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   18.00, 'EUR', 'P+R Amsterdam West',            'c2000000-0000-0000-0000-000000000007', 'cash',          '2026-04-09', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   3.80, 'EUR', 'GVB',                            'c2000000-0000-0000-0000-000000000006', 'card',          '2026-04-12', NULL, FALSE),
  -- Shopping Apr
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   89.99, 'EUR', 'Zara',                          'c1000000-0000-0000-0000-000000000005', 'card',          '2026-04-03', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   134.99, 'EUR', 'bol.com',                      'c1000000-0000-0000-0000-000000000005', 'card',          '2026-04-07', 'New monitor stand', FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   24.50, 'EUR', 'HEMA',                          'c1000000-0000-0000-0000-000000000005', 'card',          '2026-04-09', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   67.49, 'EUR', 'Amazon',                        'c1000000-0000-0000-0000-000000000005', 'card',          '2026-04-12', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   71.00, 'EUR', 'MediaMarkt',                    'c1000000-0000-0000-0000-000000000005', 'card',          '2026-04-15', NULL, FALSE),
  -- Entertainment Apr (near budget limit!)
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   29.99, 'EUR', 'Steam',                         'c1000000-0000-0000-0000-000000000006', 'card',          '2026-04-05', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   24.50, 'EUR', 'Pathé Cinema',                  'c1000000-0000-0000-0000-000000000006', 'card',          '2026-04-08', NULL, FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   10.00, 'EUR', 'Patreon',                       'c1000000-0000-0000-0000-000000000006', 'card',          '2026-04-10', NULL, TRUE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   4.99, 'EUR', 'Apple TV+',                      'c1000000-0000-0000-0000-000000000006', 'card',          '2026-04-11', NULL, TRUE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   89.00, 'EUR', 'Ticketmaster',                  'c1000000-0000-0000-0000-000000000006', 'card',          '2026-04-14', 'Concert tickets', FALSE),
  (gen_random_uuid(), 'personal', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2', '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   42.00, 'EUR', 'Bar Spek',                      'c1000000-0000-0000-0000-000000000006', 'cash',          '2026-04-16', NULL, FALSE);

-- ── 6. Smith Family — Budget cycle + household expenses ───────────────────────
INSERT INTO budget_cycle_configs (id, household_id, start_day, effective_from, created_by, created_at)
VALUES ('bc000002-0000-0000-0000-000000000001',
        '710543c2-0893-46e0-b10d-02e7b736edc0',
        1, '2026-02-01',
        '5c2b89da-6d4a-4b34-85a1-13ceae06b20b', NOW() - INTERVAL '75 days')
ON CONFLICT DO NOTHING;

INSERT INTO cycle_snapshots (id, household_id, cycle_start, cycle_end, label, status, config_id, created_at)
VALUES
  ('c5000002-0000-0000-0000-000000000001',
   '710543c2-0893-46e0-b10d-02e7b736edc0',
   '2026-02-01', '2026-02-28', 'February 2026', 'closed',
   'bc000002-0000-0000-0000-000000000001', '2026-02-01'),
  ('c5000002-0000-0000-0000-000000000002',
   '710543c2-0893-46e0-b10d-02e7b736edc0',
   '2026-03-01', '2026-03-31', 'March 2026', 'closed',
   'bc000002-0000-0000-0000-000000000001', '2026-03-01'),
  ('c5000002-0000-0000-0000-000000000003',
   '710543c2-0893-46e0-b10d-02e7b736edc0',
   '2026-04-01', '2026-04-30', 'April 2026', 'open',
   'bc000002-0000-0000-0000-000000000001', '2026-04-01')
ON CONFLICT (household_id, cycle_start, cycle_end) DO NOTHING;

-- Household budgets for Smith Family (April)
INSERT INTO budgets (id, scope, household_id, cycle_snapshot_id, category_id, amount, rollover_amount)
VALUES
  ('bd000002-0000-0000-0000-000000000001', 'household',
   '710543c2-0893-46e0-b10d-02e7b736edc0',
   'c5000002-0000-0000-0000-000000000003', NULL, 5000.00, 0),
  ('bd000002-0000-0000-0000-000000000002', 'household',
   '710543c2-0893-46e0-b10d-02e7b736edc0',
   'c5000002-0000-0000-0000-000000000003',
   'c1000000-0000-0000-0000-000000000001', 900.00, 0),
  ('bd000002-0000-0000-0000-000000000003', 'household',
   '710543c2-0893-46e0-b10d-02e7b736edc0',
   'c5000002-0000-0000-0000-000000000003',
   'c1000000-0000-0000-0000-000000000003', 2200.00, 0),
  ('bd000002-0000-0000-0000-000000000004', 'household',
   '710543c2-0893-46e0-b10d-02e7b736edc0',
   'c5000002-0000-0000-0000-000000000003',
   'c1000000-0000-0000-0000-000000000006', 350.00, 0),
  ('bd000002-0000-0000-0000-000000000005', 'household',
   '710543c2-0893-46e0-b10d-02e7b736edc0',
   'c5000002-0000-0000-0000-000000000003',
   'c1000000-0000-0000-0000-000000000002', 400.00, 0)
ON CONFLICT DO NOTHING;

-- Smith Family household expenses (scope='household', household_id set)
INSERT INTO expenses
  (id, scope, household_id, created_by, amount, currency, merchant, category_id, payment_method, date, notes, is_recurring)
VALUES
  -- Housing (rent already paid)
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '5c2b89da-6d4a-4b34-85a1-13ceae06b20b',
   2100.00, 'USD', 'Landlord',                    'c1000000-0000-0000-0000-000000000003', 'bank_transfer', '2026-04-01', 'Monthly rent', TRUE),
  -- Groceries (Alice shops)
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '5c2b89da-6d4a-4b34-85a1-13ceae06b20b',
   124.30, 'USD', 'Whole Foods',                  'c2000000-0000-0000-0000-000000000001', 'card',          '2026-04-02', NULL, FALSE),
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '14e54fe6-ff33-44e0-beee-9780985bde9d',
   87.50, 'USD', 'Trader Joe''s',                 'c2000000-0000-0000-0000-000000000001', 'card',          '2026-04-05', NULL, FALSE),
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '5c2b89da-6d4a-4b34-85a1-13ceae06b20b',
   98.20, 'USD', 'Whole Foods',                   'c2000000-0000-0000-0000-000000000001', 'card',          '2026-04-09', NULL, FALSE),
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '14e54fe6-ff33-44e0-beee-9780985bde9d',
   62.80, 'USD', 'Costco',                        'c2000000-0000-0000-0000-000000000001', 'card',          '2026-04-12', NULL, FALSE),
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '5c2b89da-6d4a-4b34-85a1-13ceae06b20b',
   76.40, 'USD', 'Whole Foods',                   'c2000000-0000-0000-0000-000000000001', 'card',          '2026-04-15', NULL, FALSE),
  -- Restaurants
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '5c2b89da-6d4a-4b34-85a1-13ceae06b20b',
   82.00, 'USD', 'Nobu',                          'c2000000-0000-0000-0000-000000000002', 'card',          '2026-04-04', 'Anniversary dinner', FALSE),
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '14e54fe6-ff33-44e0-beee-9780985bde9d',
   45.50, 'USD', 'Shake Shack',                   'c2000000-0000-0000-0000-000000000002', 'cash',          '2026-04-08', NULL, FALSE),
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '5c2b89da-6d4a-4b34-85a1-13ceae06b20b',
   34.00, 'USD', 'Chipotle',                      'c2000000-0000-0000-0000-000000000004', 'card',          '2026-04-11', NULL, FALSE),
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '14e54fe6-ff33-44e0-beee-9780985bde9d',
   28.50, 'USD', 'Sweetgreen',                    'c2000000-0000-0000-0000-000000000002', 'card',          '2026-04-14', NULL, FALSE),
  -- Entertainment
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '5c2b89da-6d4a-4b34-85a1-13ceae06b20b',
   15.99, 'USD', 'Netflix',                       'c1000000-0000-0000-0000-000000000006', 'card',          '2026-04-01', NULL, TRUE),
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '5c2b89da-6d4a-4b34-85a1-13ceae06b20b',
   22.99, 'USD', 'Disney+',                       'c1000000-0000-0000-0000-000000000006', 'card',          '2026-04-01', NULL, TRUE),
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '14e54fe6-ff33-44e0-beee-9780985bde9d',
   48.00, 'USD', 'AMC Theaters',                  'c1000000-0000-0000-0000-000000000006', 'card',          '2026-04-06', NULL, FALSE),
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '5c2b89da-6d4a-4b34-85a1-13ceae06b20b',
   95.00, 'USD', 'Broadway Tickets',              'c1000000-0000-0000-0000-000000000006', 'card',          '2026-04-10', NULL, FALSE),
  -- Transportation
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '14e54fe6-ff33-44e0-beee-9780985bde9d',
   74.20, 'USD', 'Exxon',                         'c2000000-0000-0000-0000-000000000005', 'card',          '2026-04-03', NULL, FALSE),
  (gen_random_uuid(), 'household', '710543c2-0893-46e0-b10d-02e7b736edc0',
   '5c2b89da-6d4a-4b34-85a1-13ceae06b20b',
   68.80, 'USD', 'Shell',                         'c2000000-0000-0000-0000-000000000005', 'card',          '2026-04-13', NULL, FALSE);

-- ── 7. Notifications for Ahmed ────────────────────────────────────────────────
INSERT INTO notifications (id, user_id, type, title, body, metadata, created_at)
VALUES
  (gen_random_uuid(),
   '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   'budget_threshold',
   'Entertainment budget at 90%',
   'You''ve used €225 of your €250 Entertainment budget this month.',
   '{"budget_id": "bd000001-0000-0000-0000-000000000006", "threshold": 90, "pct_used": 90.1}',
   NOW() - INTERVAL '2 hours'),
  (gen_random_uuid(),
   '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   'budget_threshold',
   'Shopping budget at 75%',
   'You''ve used €387 of your €500 Shopping budget this month.',
   '{"budget_id": "bd000001-0000-0000-0000-000000000005", "threshold": 75, "pct_used": 77.4}',
   NOW() - INTERVAL '1 day'),
  (gen_random_uuid(),
   '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
   'budget_threshold',
   'Food & Dining budget at 50%',
   'You''re halfway through your €700 Food & Dining budget.',
   '{"budget_id": "bd000001-0000-0000-0000-000000000002", "threshold": 50, "pct_used": 52.7}',
   NOW() - INTERVAL '5 days'),
  (gen_random_uuid(),
   '5c2b89da-6d4a-4b34-85a1-13ceae06b20b',
   'budget_threshold',
   'Entertainment budget at 75%',
   'Smith Family has used $267 of the $350 Entertainment budget.',
   '{"budget_id": "bd000002-0000-0000-0000-000000000004", "threshold": 75, "pct_used": 76.3}',
   NOW() - INTERVAL '3 hours');

-- ── 8. Audit log entries for Ahmed's April expenses ──────────────────────────
INSERT INTO audit_log (user_id, action, entity_type, entity_id, new_values, created_at)
SELECT
  '6aa95384-2a95-4887-a0f5-e59d4a7c64c2',
  'create',
  'expense',
  e.id,
  jsonb_build_object('amount', e.amount, 'merchant', e.merchant, 'date', e.date),
  e.created_at
FROM expenses e
WHERE e.user_id = '6aa95384-2a95-4887-a0f5-e59d4a7c64c2'
  AND e.date >= '2026-04-01';

COMMIT;
