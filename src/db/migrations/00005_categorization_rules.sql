-- +goose Up
create table categorization_rules (
    id           uuid primary key default gen_random_uuid(),
    household_id uuid not null references households(id),
    keyword      text not null,
    category_id  uuid not null references categories(id),
    match_count  integer not null default 0 check (match_count >= 0),
    created_at   timestamptz not null default now(),
    unique (household_id, keyword, category_id)
);

-- +goose Down
drop table categorization_rules;
