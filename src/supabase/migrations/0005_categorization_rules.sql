-- T009 · Categorization rules (shared; learns from household history) [FR-015]

create table if not exists categorization_rules (
    id           uuid primary key default gen_random_uuid(),
    household_id uuid not null references households(id) on delete cascade,
    keyword      text not null,
    category_id  uuid not null references categories(id) on delete cascade,
    match_count  integer not null default 0 check (match_count >= 0),
    created_at   timestamptz not null default now()
);
create index if not exists idx_rules_hh on categorization_rules(household_id);

alter table categorization_rules enable row level security;
