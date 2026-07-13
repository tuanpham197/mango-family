-- +goose Up
create table categories (
    id           uuid primary key default gen_random_uuid(),
    household_id uuid not null references households(id),
    name         text not null,
    type         text not null check (type in ('INCOME','EXPENSE')),
    icon         text,
    parent_id    uuid references categories(id),
    is_default   boolean not null default false,
    is_hidden    boolean not null default false,
    created_by   uuid references users(id),
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now()
);

create index idx_categories_household on categories(household_id, type, is_hidden);

-- +goose Down
drop table categories;
