-- Contract: Supabase/PostgreSQL schema DELTA for Income & Expense Tracking (feature 002)
-- Base: feature 001 schema (households, household_members, categories, transactions,
--       categorization_rules + RLS by membership). This contract layers on top.
-- Executable copies live in src/supabase/migrations/ (0011 written; 0012/0013 at implement).
-- Traceability: [FR-xxx] = spec 002 · [Rnn] = research.md 002 · [BR-TRK/ACC-xxx] = BR-002/BR-005.

-- =========================================================
-- 0011 · users — INDEPENDENT user directory  [FR-015][FR-016][R15]  (WRITTEN)
-- =========================================================
create table users (
    id           uuid primary key default gen_random_uuid(),  -- independent: NO FK to auth
    email        text not null unique,                        -- login identity, session mapping key
    display_name text not null,
    created_at   timestamptz not null default now()
);
-- session → user bridge (SECURITY DEFINER STABLE)
create function current_user_id() returns uuid as $$
    select id from users where email = auth.jwt()->>'email';
$$ language sql security definer stable;
-- created_by auto-set on households/categories/transactions:
--   new.created_by := coalesce(current_user_id(), new.created_by)   [trigger set_created_by]
-- FK repoint to users(id): household_members.user_id, households/categories/transactions.created_by
-- is_member()/member_self policies recheck via current_user_id()
-- RLS users: SELECT self-or-same-household (shares_household); INSERT/UPDATE self (by email)

-- =========================================================
-- 0012 · accounts + default seed + balances view  [FR-006][FR-011][R16][R17]
-- =========================================================
create table accounts (
    id              uuid primary key default gen_random_uuid(),
    household_id    uuid not null references households(id) on delete cascade,
    name            text not null,                                          -- [BR-ACC-002]
    type            text not null default 'CASH'
                    check (type in ('CASH','BANK','EWALLET','CREDIT')),     -- [BR-ACC-001]
    initial_balance numeric(14,2) not null default 0,                       -- [BR-ACC-002]
    created_by      uuid references users(id),                              -- audit (trigger-set)
    created_at      timestamptz not null default now()
);
create index idx_accounts_hh on accounts(household_id);
alter table accounts enable row level security;
create policy member_accounts on accounts
    using (is_member(household_id)) with check (is_member(household_id));
create trigger trg_created_by_accounts before insert on accounts
    for each row execute function set_created_by();

-- every household always has at least one account  [R17, edge case spec]
create function seed_default_account() returns trigger as $$
begin
    insert into accounts(household_id, name, type, created_by)
    values (new.id, 'Tiền mặt', 'CASH', new.created_by);
    return new;
end; $$ language plpgsql security definer;
create trigger trg_seed_default_account after insert on households
    for each row execute function seed_default_account();
-- + one-shot backfill for existing households without any account

-- balance is DERIVED — always equals initial + signed sum of its transactions  [FR-011][SC-004][R16]
create view account_balances with (security_invoker = true) as
select a.id as account_id,
       a.household_id,
       a.initial_balance
         + coalesce(sum(case t.type when 'INCOME' then t.amount else -t.amount end), 0) as balance
from accounts a
left join transactions t on t.account_id = a.id
group by a.id, a.household_id, a.initial_balance;

-- =========================================================
-- 0013 · transactions v2 — account link, optimistic timestamp, date guard
-- =========================================================
alter table transactions
    add column account_id uuid references accounts(id) on delete restrict,  -- [FR-006][BR-TRK-006]
    add column updated_at timestamptz not null default now();               -- [FR-014][R20]
-- backfill account_id to the household's default account, then SET NOT NULL
create index idx_transactions_account on transactions(account_id);
create trigger trg_touch_updated_at_txn before update on transactions
    for each row execute function touch_updated_at();                       -- reuse 001

-- extend guard: account same household + no future dates  [FR-005][R18]
create or replace function enforce_txn_rules() returns trigger as $$
declare v_cat_type text; v_cat_hh uuid; v_acc_hh uuid;
begin
    select type, household_id into v_cat_type, v_cat_hh from categories where id = new.category_id;
    if new.type <> v_cat_type then
        raise exception 'Transaction type must match category type (FR-014/001)';
    end if;
    if new.household_id <> v_cat_hh then
        raise exception 'Transaction and category must belong to the same household';
    end if;
    select household_id into v_acc_hh from accounts where id = new.account_id;
    if new.household_id <> v_acc_hh then
        raise exception 'Transaction and account must belong to the same household (FR-006)';
    end if;
    if new.transaction_date > now() + interval '1 day' then
        raise exception 'Future-dated transactions are not supported (FR-005)';
    end if;
    return new;
end; $$ language plpgsql;

-- Optimistic writes (client-side contract, not schema):
--   UPDATE transactions SET ... WHERE id = :id AND updated_at = :last_seen  → 0 rows = conflict
--   DELETE transactions          WHERE id = :id                             → 0 rows = already gone
