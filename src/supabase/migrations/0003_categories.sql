-- T007 · Categories (shared in household) + integrity triggers
-- [FR-004][FR-005][FR-010][FR-011][FR-017][SC-008]

create table if not exists categories (
    id           uuid primary key default gen_random_uuid(),
    household_id uuid not null references households(id) on delete cascade,
    name         text not null,
    type         text not null check (type in ('INCOME','EXPENSE')),   -- [FR-004]
    icon         text,
    parent_id    uuid references categories(id) on delete restrict,    -- one level [FR-010]
    is_default   boolean not null default false,                       -- [FR-001][FR-007]
    is_hidden    boolean not null default false,                       -- [FR-020]
    created_by   uuid references auth.users(id),                       -- audit [R14]
    created_at   timestamptz not null default now()
);
create index if not exists idx_categories_hh_type on categories(household_id, type);
create index if not exists idx_categories_parent on categories(parent_id);

alter table categories enable row level security;

-- One-level nesting; subcategory inherits parent type [FR-010][FR-011][SC-008]
create or replace function enforce_one_level() returns trigger as $$
begin
    if new.parent_id is not null then
        if (select parent_id from categories where id = new.parent_id) is not null then
            raise exception 'Subcategory nesting limited to one level (FR-010)';
        end if;
        new.type := (select type from categories where id = new.parent_id);
    end if;
    return new;
end; $$ language plpgsql;
drop trigger if exists trg_inherit_type on categories;
create trigger trg_inherit_type before insert or update on categories
    for each row execute function enforce_one_level();

-- Type is immutable after creation [FR-005]
create or replace function reject_type_change() returns trigger as $$
begin
    if new.type <> old.type then
        raise exception 'Category type is immutable after creation (FR-005)';
    end if;
    return new;
end; $$ language plpgsql;
drop trigger if exists trg_type_immutable on categories;
create trigger trg_type_immutable before update on categories
    for each row execute function reject_type_change();
