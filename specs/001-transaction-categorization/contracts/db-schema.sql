-- Contract: PostgreSQL schema for Transaction Categorization (feature 001) — Go + Vue re-platform
-- Model: SHARED FAMILY HOUSEHOLD ledger, multiple members, EQUAL permissions.
-- This is the DATA CONTRACT. Executable copies live in src/db/migrations/ (goose,
-- files 00001_users.sql … 00005_categorization_rules.sql, each with -- +goose Up/Down).
-- Business invariants (type immutable, one-level nesting, type-match, delete-reassign)
-- are enforced in the Go biz layer inside DB transactions (research D3); the schema
-- keeps CHECK/FK/UNIQUE as the last line of defense. NO RLS — household scoping is
-- enforced by API middleware (research D5). INCOME = Thu, EXPENSE = Chi.
-- Traceability tags [FR-xxx]/[SC-xxx]/[Dn] map to spec.md / research.md.

create extension if not exists pgcrypto; -- gen_random_uuid()

-- ============================ users (kiêm đăng nhập — D4) ============================
create table users (
    id            uuid primary key default gen_random_uuid(),
    email         text not null unique,                      -- [FR-016/002] danh tính đăng nhập
    display_name  text not null,                             -- [UC-TRK-01 5a] fallback = email
    password_hash text not null,                             -- bcrypt; never returned via API [D4]
    created_at    timestamptz not null default now()
);

-- ============================ households & members ============================
create table households (
    id         uuid primary key default gen_random_uuid(),
    name       text not null,
    created_by uuid not null references users(id),
    created_at timestamptz not null default now()
);

create table household_members (                             -- [FR-018] shared-in-household
    household_id uuid not null references households(id),
    user_id      uuid not null references users(id),
    joined_at    timestamptz not null default now(),
    primary key (household_id, user_id)                      -- no role column [FR-021]
);

-- ============================ categories ============================
create table categories (
    id           uuid primary key default gen_random_uuid(),
    household_id uuid not null references households(id),    -- [FR-018]
    name         text not null,                              -- dup warn in biz [FR-017]
    type         text not null check (type in ('INCOME','EXPENSE')), -- [FR-004]; immutable in biz [FR-005]
    icon         text,                                       -- [FR-003]
    parent_id    uuid references categories(id),             -- one level only, checked in biz [FR-010/011]
    is_default   boolean not null default false,             -- [FR-001/007]
    is_hidden    boolean not null default false,             -- [FR-020]
    created_by   uuid references users(id),                  -- audit [FR-022]
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now()          -- optimistic timestamp [D6]
);
create index idx_categories_household on categories(household_id, type, is_hidden);

-- ============================ transactions (001: categorization slice) ============================
create table transactions (
    id               uuid primary key default gen_random_uuid(),
    household_id     uuid not null references households(id),   -- [FR-018]
    created_by       uuid not null references users(id),        -- set by biz from session [FR-022]
    amount           numeric(14,2) not null check (amount > 0), -- [BR-TRK-001]
    type             text not null check (type in ('INCOME','EXPENSE')), -- must match category type (biz) [FR-014]
    category_id      uuid not null references categories(id),   -- no orphan transactions [FR-013, SC-002]
    description      text,                                      -- <= 255 checked in biz [BR-TRK-004]
    transaction_date timestamptz not null default now()         -- [BR-TRK-005]
    -- feature 002 will add: account_id, updated_at, no-future-date rule
);
create index idx_transactions_household on transactions(household_id, transaction_date desc);
create index idx_transactions_category on transactions(category_id);

-- ============================ categorization_rules (suggest — D7) ============================
create table categorization_rules (
    id           uuid primary key default gen_random_uuid(),
    household_id uuid not null references households(id),    -- [FR-015/018] learns from household history
    keyword      text not null,                              -- normalized lowercase (biz)
    category_id  uuid not null references categories(id),
    match_count  integer not null default 0 check (match_count >= 0),
    created_at   timestamptz not null default now(),
    unique (household_id, keyword, category_id)              -- upsert target [D7]
);

-- Dev seed (Alice/Bob → household A, Carol → household B, default categories per
-- household) is a SEPARATE dev-only step: `go run ./cmd/seed` — NOT a migration (D9).
