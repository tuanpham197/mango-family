-- ============================================================================
-- setup_dev.sql — Dán TOÀN BỘ file này vào Supabase Dashboard → SQL Editor → Run
-- Gộp: 3 user dev (đã xác nhận email) + schema (migrations 0001–0007, 0010)
--      + seed hộ A (Alice, Bob) / hộ B (Carol) + danh mục mặc định.
-- Idempotent: chạy lại nhiều lần không sao.
-- Tài khoản dev:  alice@dev.local / bob@dev.local / carol@dev.local
-- Mật khẩu chung: Password123!
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 0 · Dev users (Alice & Bob cùng hộ A; Carol hộ B — quickstart #13–#17)
-- ----------------------------------------------------------------------------
do $$
declare
    v_emails constant text[] :=
        array['alice@dev.local', 'bob@dev.local', 'carol@dev.local'];
    v_email text;
    v_id    uuid;
begin
    foreach v_email in array v_emails loop
        if not exists (select 1 from auth.users where email = v_email) then
            v_id := gen_random_uuid();
            insert into auth.users (
                instance_id, id, aud, role, email, encrypted_password,
                email_confirmed_at, raw_app_meta_data, raw_user_meta_data,
                created_at, updated_at,
                confirmation_token, recovery_token,
                email_change, email_change_token_new, email_change_token_current
            ) values (
                '00000000-0000-0000-0000-000000000000', v_id,
                'authenticated', 'authenticated', v_email,
                extensions.crypt('Password123!', extensions.gen_salt('bf')),
                now(), '{"provider":"email","providers":["email"]}'::jsonb,
                '{}'::jsonb, now(), now(), '', '', '', '', ''
            );
            insert into auth.identities (
                id, user_id, identity_data, provider, provider_id,
                last_sign_in_at, created_at, updated_at
            ) values (
                gen_random_uuid(), v_id,
                jsonb_build_object('sub', v_id::text, 'email', v_email,
                                   'email_verified', true),
                'email', v_id::text, now(), now(), now()
            );
        end if;
    end loop;
    -- Xác nhận email cho cả user tạo sẵn từ trước (nếu có).
    update auth.users set email_confirmed_at = coalesce(email_confirmed_at, now())
        where email = any(v_emails);
end $$;

-- ----------------------------------------------------------------------------
-- 1 · 0001_households.sql
-- ----------------------------------------------------------------------------
create table if not exists households (
    id          uuid primary key default gen_random_uuid(),
    name        text not null,
    created_by  uuid not null references auth.users(id),
    created_at  timestamptz not null default now()
);

create table if not exists household_members (
    household_id uuid not null references households(id) on delete cascade,
    user_id      uuid not null references auth.users(id) on delete cascade,
    joined_at    timestamptz not null default now(),
    primary key (household_id, user_id)
);
create index if not exists idx_member_user on household_members(user_id);

-- ----------------------------------------------------------------------------
-- 2 · 0002_rls_helpers.sql
-- ----------------------------------------------------------------------------
create or replace function is_member(p_household uuid) returns boolean as $$
    select exists (
        select 1 from household_members
        where household_id = p_household and user_id = auth.uid()
    );
$$ language sql security definer stable;

alter table households        enable row level security;
alter table household_members enable row level security;

-- ----------------------------------------------------------------------------
-- 3 · 0003_categories.sql
-- ----------------------------------------------------------------------------
create table if not exists categories (
    id           uuid primary key default gen_random_uuid(),
    household_id uuid not null references households(id) on delete cascade,
    name         text not null,
    type         text not null check (type in ('INCOME','EXPENSE')),
    icon         text,
    parent_id    uuid references categories(id) on delete restrict,
    is_default   boolean not null default false,
    is_hidden    boolean not null default false,
    created_by   uuid references auth.users(id),
    created_at   timestamptz not null default now()
);
create index if not exists idx_categories_hh_type on categories(household_id, type);
create index if not exists idx_categories_parent on categories(parent_id);

alter table categories enable row level security;

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

-- ----------------------------------------------------------------------------
-- 4 · 0004_transactions.sql
-- ----------------------------------------------------------------------------
create table if not exists transactions (
    id                uuid primary key default gen_random_uuid(),
    household_id      uuid not null references households(id) on delete cascade,
    created_by        uuid not null references auth.users(id),
    amount            numeric(14,2) not null check (amount > 0),
    type              text not null check (type in ('INCOME','EXPENSE')),
    category_id       uuid not null references categories(id) on delete restrict,
    description       text check (char_length(description) <= 255),
    transaction_date  timestamptz not null default now()
);
create index if not exists idx_transactions_category on transactions(category_id);
create index if not exists idx_transactions_hh on transactions(household_id);

alter table transactions enable row level security;

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

-- ----------------------------------------------------------------------------
-- 5 · 0005_categorization_rules.sql
-- ----------------------------------------------------------------------------
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

-- ----------------------------------------------------------------------------
-- 6 · 0006_delete_category.sql
-- ----------------------------------------------------------------------------
create or replace function delete_category(p_category uuid, p_action text, p_target uuid default null)
returns void as $$
declare v_type text; v_hh uuid;
begin
    select type, household_id into v_type, v_hh from categories where id = p_category;
    if v_hh is null then raise exception 'Category not found'; end if;
    if not is_member(v_hh) then raise exception 'Not a member of this household'; end if;

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

    delete from categories where id = p_category or parent_id = p_category;
end; $$ language plpgsql security definer;

-- ----------------------------------------------------------------------------
-- 7 · 0007_rls_policies.sql
-- ----------------------------------------------------------------------------
drop policy if exists member_households on households;
create policy member_households on households
    using (is_member(id));

drop policy if exists member_self on household_members;
create policy member_self on household_members
    using (user_id = auth.uid() or is_member(household_id));

drop policy if exists member_categories on categories;
create policy member_categories on categories
    using (is_member(household_id)) with check (is_member(household_id));

drop policy if exists member_transactions on transactions;
create policy member_transactions on transactions
    using (is_member(household_id)) with check (is_member(household_id));

drop policy if exists member_rules on categorization_rules;
create policy member_rules on categorization_rules
    using (is_member(household_id)) with check (is_member(household_id));

-- ----------------------------------------------------------------------------
-- 8 · 0010_updated_at.sql (optimistic concurrency — R13/T050)
-- ----------------------------------------------------------------------------
alter table categories
    add column if not exists updated_at timestamptz not null default now();

create or replace function touch_updated_at() returns trigger as $$
begin
    new.updated_at := now();
    return new;
end; $$ language plpgsql;

drop trigger if exists trg_touch_updated_at on categories;
create trigger trg_touch_updated_at before update on categories
    for each row execute function touch_updated_at();

-- ----------------------------------------------------------------------------
-- 9 · Seed hộ (thay 0008 bằng bản định danh theo email — không phụ thuộc
--     "2 user đầu tiên"): hộ A = Alice+Bob (quickstart "Gia đình A"),
--     hộ B = Carol ("Gia đình B" — kiểm cô lập #15).
-- ----------------------------------------------------------------------------
do $$
declare
    v_a constant uuid := '00000000-0000-0000-0000-000000000001';
    v_b constant uuid := '00000000-0000-0000-0000-000000000002';
    u_alice uuid; u_bob uuid; u_carol uuid;
begin
    select id into u_alice from auth.users where email = 'alice@dev.local';
    select id into u_bob   from auth.users where email = 'bob@dev.local';
    select id into u_carol from auth.users where email = 'carol@dev.local';
    if u_alice is null or u_bob is null then
        raise exception 'Thiếu user dev — phần 0 của file này chưa chạy?';
    end if;

    insert into households(id, name, created_by)
        values (v_a, 'Gia đình A', u_alice) on conflict (id) do nothing;
    insert into household_members(household_id, user_id)
        values (v_a, u_alice), (v_a, u_bob) on conflict do nothing;

    if u_carol is not null then
        insert into households(id, name, created_by)
            values (v_b, 'Gia đình B', u_carol) on conflict (id) do nothing;
        insert into household_members(household_id, user_id)
            values (v_b, u_carol) on conflict do nothing;
    end if;
end $$;

-- ----------------------------------------------------------------------------
-- 10 · 0009_seed_default_categories.sql (FR-001, R9) — cho cả hộ A và B
-- ----------------------------------------------------------------------------
do $$
declare
    v_hh uuid; v_owner uuid;
begin
    for v_hh in
        select id from households
        where id in ('00000000-0000-0000-0000-000000000001',
                     '00000000-0000-0000-0000-000000000002')
    loop
        select created_by into v_owner from households where id = v_hh;

        insert into categories (household_id, name, type, is_default, created_by)
        select v_hh, x.name, 'EXPENSE', true, v_owner
        from (values ('Ăn uống'),('Di chuyển'),('Hóa đơn'),('Mua sắm'),('Giải trí'),('Sức khỏe')) as x(name)
        where not exists (
            select 1 from categories c
            where c.household_id = v_hh and c.type = 'EXPENSE'
              and c.name = x.name and c.parent_id is null
        );

        insert into categories (household_id, name, type, is_default, created_by)
        select v_hh, x.name, 'INCOME', true, v_owner
        from (values ('Lương'),('Thưởng')) as x(name)
        where not exists (
            select 1 from categories c
            where c.household_id = v_hh and c.type = 'INCOME'
              and c.name = x.name and c.parent_id is null
        );
    end loop;
end $$;

-- ----------------------------------------------------------------------------
-- 11 · Grants cho API roles (PostgREST) — RLS vẫn là lớp kiểm soát theo hàng.
--      Cần vì bảng tạo qua SQL Editor có thể không nhận default privileges.
-- ----------------------------------------------------------------------------
grant usage on schema public to anon, authenticated;
grant select, insert, update, delete on all tables in schema public to authenticated;
grant execute on all functions in schema public to anon, authenticated;
alter default privileges in schema public
    grant select, insert, update, delete on tables to authenticated;
alter default privileges in schema public
    grant execute on functions to anon, authenticated;

-- Xong. Kiểm nhanh:
select 'users' as what, count(*) from auth.users where email like '%@dev.local'
union all
select 'households', count(*) from households
union all
select 'members', count(*) from household_members
union all
select 'categories', count(*) from categories;
