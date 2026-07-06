-- Contract: Supabase/PostgreSQL schema for Transaction Categorization (feature 001)
-- Model: SHARED FAMILY HOUSEHOLD ledger, multiple members, EQUAL permissions.
-- This is the DATA CONTRACT. Executable copy lives in supabase/migrations/.
-- Traceability tags [FR-xxx]/[SC-xxx]/[Rn] map back to spec.md / research.md.
-- INCOME = Thu, EXPENSE = Chi

-- =========================================================
-- households & membership  [R10][R11]
-- =========================================================
create table households (
    id          uuid primary key default gen_random_uuid(),
    name        text not null,
    created_by  uuid not null references auth.users(id),
    created_at  timestamptz not null default now()
);

create table household_members (
    household_id uuid not null references households(id) on delete cascade,
    user_id      uuid not null references auth.users(id) on delete cascade,
    joined_at    timestamptz not null default now(),
    primary key (household_id, user_id)        -- no role column: all members equal [R11]
);
create index idx_member_user on household_members(user_id);

-- membership check used by all RLS policies [R12]
create function is_member(p_household uuid) returns boolean as $$
    select exists (
        select 1 from household_members
        where household_id = p_household and user_id = auth.uid()
    );
$$ language sql security definer stable;

-- ---------- categories (shared in household) ----------
create table categories (
    id          uuid primary key default gen_random_uuid(),
    household_id uuid not null references households(id) on delete cascade,  -- [R10]
    name        text not null,
    type        text not null check (type in ('INCOME','EXPENSE')),         -- [FR-004]
    icon        text,
    parent_id   uuid references categories(id) on delete restrict,          -- one level [FR-010]
    is_default  boolean not null default false,                             -- [FR-001][FR-007]
    is_hidden   boolean not null default false,                             -- [FR-020]
    created_by  uuid references auth.users(id),                             -- audit [R14]
    created_at  timestamptz not null default now(),
    updated_at  timestamptz not null default now()                          -- optimistic concurrency [R13]
);
create index idx_categories_hh_type on categories(household_id, type);
create index idx_categories_parent on categories(parent_id);

-- one-level nesting + subcategory inherits parent type [FR-010][FR-011][SC-008]
create function enforce_one_level() returns trigger as $$
begin
    if new.parent_id is not null then
        if (select parent_id from categories where id = new.parent_id) is not null then
            raise exception 'Subcategory nesting limited to one level (FR-010)';
        end if;
        new.type := (select type from categories where id = new.parent_id);
    end if;
    return new;
end; $$ language plpgsql;
create trigger trg_inherit_type before insert or update on categories
    for each row execute function enforce_one_level();

-- type immutable after creation [FR-005]
create function reject_type_change() returns trigger as $$
begin
    if new.type <> old.type then
        raise exception 'Category type is immutable after creation (FR-005)';
    end if;
    return new;
end; $$ language plpgsql;
create trigger trg_type_immutable before update on categories
    for each row execute function reject_type_change();

-- touch updated_at on every write; clients send the last-seen updated_at as a
-- conditional filter so a stale write matches 0 rows (no silent overwrite) [R13]
create function touch_updated_at() returns trigger as $$
begin
    new.updated_at := now();
    return new;
end; $$ language plpgsql;
create trigger trg_touch_updated_at before update on categories
    for each row execute function touch_updated_at();

-- ---------- transactions (categorization-relevant subset; full model BR-002) ----------
create table transactions (
    id                uuid primary key default gen_random_uuid(),
    household_id      uuid not null references households(id) on delete cascade, -- [R10]
    created_by        uuid not null references auth.users(id),                   -- who entered [R14]
    amount            numeric(14,2) not null check (amount > 0),                 -- [BR-TRK-001]
    type              text not null check (type in ('INCOME','EXPENSE')),
    category_id       uuid not null references categories(id) on delete restrict,-- [FR-013] no orphans
    description       text check (char_length(description) <= 255),             -- [BR-TRK-004]
    transaction_date  timestamptz not null default now()
);
create index idx_transactions_category on transactions(category_id);
create index idx_transactions_hh on transactions(household_id);

-- transaction type MUST equal category type, AND same household [FR-014][SC-003]
create function enforce_txn_rules() returns trigger as $$
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
create trigger trg_txn_rules before insert or update on transactions
    for each row execute function enforce_txn_rules();

-- ---------- categorization_rules (shared, learns from household history) ----------
create table categorization_rules (
    id           uuid primary key default gen_random_uuid(),
    household_id uuid not null references households(id) on delete cascade,   -- [R10]
    keyword      text not null,
    category_id  uuid not null references categories(id) on delete cascade,
    match_count  integer not null default 0 check (match_count >= 0),
    created_at   timestamptz not null default now()
);

-- =========================================================
-- Safe category deletion (household-scoped) [FR-008][FR-009][FR-012][SC-007]
-- =========================================================
create function delete_category(p_category uuid, p_action text, p_target uuid default null)
returns void as $$
declare v_type text; v_hh uuid;
begin
    select type, household_id into v_type, v_hh from categories where id = p_category;
    if not is_member(v_hh) then raise exception 'Not a member of this household'; end if;  -- [R12]
    if p_action = 'REASSIGN' then
        if (select type from categories where id = p_target) <> v_type then
            raise exception 'Reassign target must be same type (FR-009)';
        end if;
        update transactions set category_id = p_target
            where category_id in (select id from categories
                                  where id = p_category or parent_id = p_category);
    elsif p_action = 'DELETE' then
        delete from transactions
            where category_id in (select id from categories
                                  where id = p_category or parent_id = p_category);
    else
        raise exception 'Unknown action %', p_action;
    end if;
    delete from categories where id = p_category or parent_id = p_category;  -- parent + children [FR-012]
end; $$ language plpgsql security definer;

-- =========================================================
-- Row-Level Security: members of a household share its data; isolated across households
-- [FR-018 updated][R10][R12]   (equal permissions: same policy for read & write [R11])
-- =========================================================
alter table households            enable row level security;
alter table household_members     enable row level security;
alter table categories            enable row level security;
alter table transactions          enable row level security;
alter table categorization_rules  enable row level security;

create policy member_households on households
    using (is_member(id));
create policy member_self on household_members
    using (user_id = auth.uid() or is_member(household_id));
create policy member_categories on categories
    using (is_member(household_id)) with check (is_member(household_id));
create policy member_transactions on transactions
    using (is_member(household_id)) with check (is_member(household_id));
create policy member_rules on categorization_rules
    using (is_member(household_id)) with check (is_member(household_id));

-- =========================================================
-- Default category seed [FR-001] — runs on HOUSEHOLD creation (not per user) [R9][R10]
-- (assumption, pending business confirmation)
-- EXPENSE: Ăn uống, Di chuyển, Hóa đơn, Mua sắm, Giải trí, Sức khỏe
-- INCOME:  Lương, Thưởng   (is_default = true; shared by all members)
-- =========================================================
