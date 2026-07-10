-- Contract: PostgreSQL schema DELTA for Income & Expense Tracking (feature 002) — Go + Vue re-platform
-- Base: feature 001 schema (users incl. credentials, households, household_members,
--       categories, transactions minimal, categorization_rules — see
--       specs/001-transaction-categorization/contracts/db-schema.sql). This contract layers on top.
-- Executable copies live in src/db/migrations/ (goose): 00006_accounts.sql, 00007_transactions_v2.sql.
-- NO RLS / NO business triggers — household scoping via API middleware (001 D5); invariants in
-- Go biz layer inside DB transactions (001 D3); schema keeps CHECK/FK/NOT NULL as last defense.
-- Traceability: [FR-xxx] = spec 002 · [Dnn] = research 001/002 · [BR-TRK/ACC-xxx] = BR-002/BR-005.

-- =========================================================
-- 00006 · accounts + account_balances view  [FR-006][FR-011][D13][D14]
-- =========================================================
create table accounts (
    id              uuid primary key default gen_random_uuid(),
    household_id    uuid not null references households(id),   -- scope enforced by API [D5/001]
    name            text not null,                             -- e.g. "Tiền mặt" [BR-ACC-002]
    type            text not null default 'CASH'
                        check (type in ('CASH','BANK','EWALLET','CREDIT')), -- [BR-ACC-001]
    initial_balance numeric(14,2) not null default 0,          -- reserved for BR-005 [D13]
    created_by      uuid references users(id),                 -- audit; set by biz from session
    created_at      timestamptz not null default now()
);
create index idx_accounts_household on accounts(household_id);

-- Default account "Tiền mặt" (CASH) is created by APP LOGIC (household SeedDefaults — D13)
-- for new households, and backfilled by cmd/seed for dev households. NOT seeded here.

-- Derived balance — always equals initial_balance + signed sum of transactions [SC-004][D14]
-- (goose: wrap in -- +goose StatementBegin / StatementEnd)
create view account_balances as
select
    a.id           as account_id,
    a.household_id as household_id,
    a.initial_balance
      + coalesce(sum(case t.type when 'INCOME' then t.amount else -t.amount end), 0)
                   as balance
from accounts a
left join transactions t on t.account_id = a.id
group by a.id, a.household_id, a.initial_balance;

-- =========================================================
-- 00007 · transactions v2 — account_id + updated_at  [FR-006][FR-014][D17]
-- =========================================================
-- Backfill strategy: add column nullable → set to household default account → set not null.
alter table transactions add column account_id uuid references accounts(id) on delete restrict;
-- update transactions t set account_id = (select id from accounts a
--   where a.household_id = t.household_id order by a.created_at limit 1);
alter table transactions alter column account_id set not null;             -- [BR-TRK-006]

alter table transactions add column updated_at timestamptz not null default now(); -- optimistic mark [D17]

create index idx_transactions_account on transactions(account_id);          -- serves balance view [D14]

-- Enforced in Go biz (NOT triggers — D3/001):
--   * account belongs to the same household as the transaction  [FR-006/007]
--   * transaction_date not in the future (+1 day timezone slack) [FR-005][D15]
--   * type matches category type; category same household        [FR-003][D18]
--   * conditional UPDATE/DELETE ... WHERE updated_at = expected  [FR-014][D17]
--   * updated_at refreshed via GORM hook on every update
