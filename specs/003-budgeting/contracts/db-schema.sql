-- Contract: PostgreSQL schema DELTA for Budgeting (feature 003) — Go + Vue
-- Base: feature 001 schema (users, households, household_members, categories,
--       categorization_rules) + feature 002 schema (accounts, transactions v2, account_balances view).
--       See specs/001-*/contracts/db-schema.sql and specs/002-*/contracts/db-schema.sql. This layers on top.
-- Executable copies live in src/db/migrations/ (goose): 00008_budgets.sql, 00009_budget_alerts.sql.
-- NO RLS / NO business triggers — household scoping via API middleware (001 D5); invariants in Go biz
-- inside DB transactions (001 D3); schema keeps CHECK/FK/NOT NULL + partial unique index as last defense.
-- Budgeting READS transactions/categories only (derived progress — D21); it does NOT alter them.
-- Traceability: [FR-xxx] = spec 003 · [Dnn] = research 003 · [BR-BGT-xxx] = BR-003.

-- =========================================================
-- 00008 · budgets  [FR-001..004][FR-010..013][D20][D22][D25]
-- =========================================================
create table budgets (
    id            uuid primary key default gen_random_uuid(),
    household_id  uuid not null references households(id),          -- scope enforced by API [D5/001]
    type          text not null check (type in ('CATEGORY','TOTAL')),        -- [FR-001][FR-002]
    category_id   uuid references categories(id),                   -- null when TOTAL [D20]
    limit_amount  numeric(14,2) not null check (limit_amount > 0),  -- [FR-003][BR-BGT-003]
    period_type   text not null check (period_type in ('MONTHLY','WEEKLY','ONE_TIME')), -- [FR-004]
    start_date    date,                                             -- ONE_TIME only [D20]
    end_date      date,                                             -- ONE_TIME only [D20]
    status        text not null default 'ACTIVE' check (status in ('ACTIVE','ENDED')),  -- [D20]
    created_by    uuid not null references users(id),               -- set by biz from session [FR-011]
    created_at    timestamptz not null default now(),
    updated_at    timestamptz not null default now(),               -- optimistic mark [D25]
    -- type/category coherence [FR-001][FR-002][D20]
    constraint budget_category_presence check (
        (type = 'CATEGORY' and category_id is not null) or
        (type = 'TOTAL'    and category_id is null)
    ),
    -- ONE_TIME window sanity [FR-004] (biz also validates presence of both dates)
    constraint budget_period_range check (
        end_date is null or start_date is null or end_date >= start_date
    )
);
create index idx_budgets_household on budgets(household_id);

-- Uniqueness (FR-010, D22): at most one ACTIVE budget per (household, category, period_type)
-- and per (household, period_type) for TOTAL. Partial → ENDED budgets don't block new ones.
create unique index uq_budget_category_active
    on budgets(household_id, category_id, period_type)
    where status = 'ACTIVE' and type = 'CATEGORY';
create unique index uq_budget_total_active
    on budgets(household_id, period_type)
    where status = 'ACTIVE' and type = 'TOTAL';

-- Enforced in Go biz (NOT triggers — D3/001):
--   * category (when type=CATEGORY) is EXPENSE type and same household   [FR-001]
--   * ONE_TIME requires both start_date and end_date                     [FR-004]
--   * duplicate ACTIVE budget → BUDGET_DUPLICATE with existing id        [FR-010][D22]
--   * conditional UPDATE/DELETE ... WHERE updated_at = expected          [FR-013][D25]
--   * updated_at refreshed via GORM hook on every update
--   * ONE_TIME with now() > end_date → status transitions to ENDED       [D20]
-- Progress (spent/percent) is DERIVED in biz on read — NOT stored, NO view (D21):
--   EXPENSE-only sum over current period window; CATEGORY budgets include
--   one-level child categories; TOTAL budgets sum all household EXPENSE.

-- =========================================================
-- 00009 · budget_alerts  [FR-007][FR-008][FR-009][D23]
-- =========================================================
-- One row = one fired alert level for a budget within a period. Evaluator (D24) inserts on first
-- crossing, deletes when progress drops back below the level (re-arm). period_key resets each period
-- (MONTHLY 'YYYY-MM' / WEEKLY 'IYYY-IW' / ONE_TIME 'once') so state self-clears on period rollover.
create table budget_alerts (
    id          uuid primary key default gen_random_uuid(),
    budget_id   uuid not null references budgets(id) on delete cascade,      -- [UC-BGT-06]
    period_key  text not null,                                    -- current-period key [D20]
    level       text not null check (level in ('THRESHOLD_80','OVER_100')),  -- [FR-007][FR-008]
    over_amount numeric(14,2),                                    -- spent - limit; null unless OVER_100 [SC-005]
    fired_at    timestamptz not null default now()
);
create unique index uq_budget_alert_level
    on budget_alerts(budget_id, period_key, level);              -- at most once per level per period [FR-009]
create index idx_budget_alerts_budget on budget_alerts(budget_id);

-- Evaluator state machine (Go biz — D23/D24), per ACTIVE budget on each transactions_changed / budget CRUD:
--   percent >= 80  and no THRESHOLD_80 row for period_key → insert (fire 80%)
--   percent > 100  and no OVER_100 row for period_key    → insert with over_amount = spent - limit
--   percent <  80  and THRESHOLD_80 row exists           → delete (re-arm)   [US3 #4]
--   percent <= 100 and OVER_100 row exists               → delete (re-arm)
-- Then Publish(budgets_changed) → subscriber → wshub broadcast to household (SC-006, D26).
