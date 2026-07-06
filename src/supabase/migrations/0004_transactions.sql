-- T008 · Transactions (categorization-relevant subset; full model in BR-002)
-- [FR-013][FR-014][SC-002][SC-003][BR-TRK-001][R14]

create table if not exists transactions (
    id                uuid primary key default gen_random_uuid(),
    household_id      uuid not null references households(id) on delete cascade,
    created_by        uuid not null references auth.users(id),                    -- who entered [R14]
    amount            numeric(14,2) not null check (amount > 0),                  -- [BR-TRK-001]
    type              text not null check (type in ('INCOME','EXPENSE')),
    category_id       uuid not null references categories(id) on delete restrict, -- [FR-013] no orphans
    description       text check (char_length(description) <= 255),              -- [BR-TRK-004]
    transaction_date  timestamptz not null default now()
);
create index if not exists idx_transactions_category on transactions(category_id);
create index if not exists idx_transactions_hh on transactions(household_id);

alter table transactions enable row level security;

-- Transaction type must equal category type, and both same household [FR-014][SC-003]
create or replace function enforce_txn_rules() returns trigger as $$
declare v_cat_type text; v_cat_hh uuid;
begin
    select type, household_id into v_cat_type, v_cat_hh from categories where id = new.category_id;
    if new.type <> v_cat_type then
        raise exception 'Transaction type must match category type (FR-014)';
    end if;
    if new.household_id <> v_cat_hh then
        raise exception 'Transaction and category must belong to the same household';
    end if;
    return new;
end; $$ language plpgsql;
drop trigger if exists trg_txn_rules on transactions;
create trigger trg_txn_rules before insert or update on transactions
    for each row execute function enforce_txn_rules();
