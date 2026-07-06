-- ============================================================================
-- setup_dev.sql — Dán TOÀN BỘ file này vào Supabase Dashboard → SQL Editor → Run
-- ⚠ CHẠY NGUYÊN FILE, KHÔNG BỎ KHỐI REFRESH — script dựng lại từ đầu (greenfield):
--   users ĐẦU TIÊN (bảng độc lập), mọi FK người dùng trỏ thẳng public.users.
-- (Migration tăng dần cho môi trường có dữ liệu thật: src/supabase/migrations/0001–0011.)
-- Gồm: REFRESH + 3 user dev + users + schema 001 + seed hộ A (Alice, Bob) /
--      hộ B (Carol) + danh mục mặc định + grants.
-- Tài khoản dev:  alice@dev.local / bob@dev.local / carol@dev.local
-- Mật khẩu chung: Password123!
-- ============================================================================

-- ----------------------------------------------------------------------------
-- A · REFRESH: xóa TOÀN BỘ dữ liệu app (bảng + hàm), GIỮ tài khoản đăng nhập.
-- ----------------------------------------------------------------------------
drop view if exists account_balances;
drop table if exists categorization_rules cascade;
drop table if exists transactions cascade;
drop table if exists categories cascade;
drop table if exists accounts cascade;
drop table if exists household_members cascade;
drop table if exists households cascade;
drop table if exists public.users cascade;
drop function if exists delete_category(uuid, text, uuid);
drop function if exists is_member(uuid);
drop function if exists shares_household(uuid);
drop function if exists current_user_id();
drop function if exists enforce_one_level() cascade;
drop function if exists reject_type_change() cascade;
drop function if exists enforce_txn_rules() cascade;
drop function if exists touch_updated_at() cascade;
drop function if exists set_created_by() cascade;
drop function if exists seed_default_account() cascade;
drop trigger if exists trg_on_auth_user_created on auth.users;
drop function if exists handle_new_user();

-- ----------------------------------------------------------------------------
-- B · Tài khoản đăng nhập dev (auth) — tạo nếu chưa có, xác nhận email.
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
    update auth.users set email_confirmed_at = coalesce(email_confirmed_at, now())
        where email = any(v_emails);
end $$;

-- ----------------------------------------------------------------------------
-- 1 · USERS ĐẦU TIÊN — bảng người dùng ĐỘC LẬP (0011) [FR-015/002]
--     Backfill hồ sơ cho MỌI tài khoản đăng nhập hiện có (kể cả ngoài dev).
-- ----------------------------------------------------------------------------
create table public.users (
    id           uuid primary key default gen_random_uuid(),  -- độc lập, KHÔNG FK auth
    email        text not null unique,
    display_name text not null,
    created_at   timestamptz not null default now()
);
alter table public.users enable row level security;

insert into public.users (email, display_name)
select email, coalesce(nullif(split_part(email, '@', 1), ''), 'user')
from auth.users
where email is not null
on conflict (email) do nothing;

-- Cầu nối phiên đăng nhập → users (đối chiếu email trong phiên).
create function current_user_id() returns uuid as $$
    select id from public.users where email = auth.jwt()->>'email';
$$ language sql security definer stable;

-- ----------------------------------------------------------------------------
-- 2 · Households + membership (FK → users) + is_member [0001/0002]
-- ----------------------------------------------------------------------------
create table households (
    id          uuid primary key default gen_random_uuid(),
    name        text not null,
    created_by  uuid not null references public.users(id),
    created_at  timestamptz not null default now()
);
create table household_members (
    household_id uuid not null references households(id) on delete cascade,
    user_id      uuid not null references public.users(id) on delete cascade,
    joined_at    timestamptz not null default now(),
    primary key (household_id, user_id)
);
create index idx_member_user on household_members(user_id);

create function is_member(p_household uuid) returns boolean as $$
    select exists (
        select 1 from household_members
        where household_id = p_household and user_id = current_user_id()
    );
$$ language sql security definer stable;

alter table households        enable row level security;
alter table household_members enable row level security;

-- ----------------------------------------------------------------------------
-- 3 · Categories + integrity triggers [0003]
-- ----------------------------------------------------------------------------
create table categories (
    id           uuid primary key default gen_random_uuid(),
    household_id uuid not null references households(id) on delete cascade,
    name         text not null,
    type         text not null check (type in ('INCOME','EXPENSE')),
    icon         text,
    parent_id    uuid references categories(id) on delete restrict,
    is_default   boolean not null default false,
    is_hidden    boolean not null default false,
    created_by   uuid references public.users(id),
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now()          -- [0010]
);
create index idx_categories_hh_type on categories(household_id, type);
create index idx_categories_parent on categories(parent_id);
alter table categories enable row level security;

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

create function reject_type_change() returns trigger as $$
begin
    if new.type <> old.type then
        raise exception 'Category type is immutable after creation (FR-005)';
    end if;
    return new;
end; $$ language plpgsql;
create trigger trg_type_immutable before update on categories
    for each row execute function reject_type_change();

create function touch_updated_at() returns trigger as $$
begin
    new.updated_at := now();
    return new;
end; $$ language plpgsql;
create trigger trg_touch_updated_at before update on categories
    for each row execute function touch_updated_at();

-- ----------------------------------------------------------------------------
-- 4 · Transactions [0004]
-- ----------------------------------------------------------------------------
create table transactions (
    id                uuid primary key default gen_random_uuid(),
    household_id      uuid not null references households(id) on delete cascade,
    created_by        uuid not null references public.users(id),
    amount            numeric(14,2) not null check (amount > 0),
    type              text not null check (type in ('INCOME','EXPENSE')),
    category_id       uuid not null references categories(id) on delete restrict,
    description       text check (char_length(description) <= 255),
    transaction_date  timestamptz not null default now()
);
create index idx_transactions_category on transactions(category_id);
create index idx_transactions_hh on transactions(household_id);
alter table transactions enable row level security;

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

-- ----------------------------------------------------------------------------
-- 5 · Categorization rules [0005]
-- ----------------------------------------------------------------------------
create table categorization_rules (
    id           uuid primary key default gen_random_uuid(),
    household_id uuid not null references households(id) on delete cascade,
    keyword      text not null,
    category_id  uuid not null references categories(id) on delete cascade,
    match_count  integer not null default 0 check (match_count >= 0),
    created_at   timestamptz not null default now()
);
create index idx_rules_hh on categorization_rules(household_id);
alter table categorization_rules enable row level security;

-- ----------------------------------------------------------------------------
-- 6 · RPC delete_category [0006]
-- ----------------------------------------------------------------------------
create function delete_category(p_category uuid, p_action text, p_target uuid default null)
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
-- 7 · RLS policies [0007 + 0011]
-- ----------------------------------------------------------------------------
create policy member_households on households
    using (is_member(id));
create policy member_self on household_members
    using (user_id = current_user_id() or is_member(household_id));
create policy member_categories on categories
    using (is_member(household_id)) with check (is_member(household_id));
create policy member_transactions on transactions
    using (is_member(household_id)) with check (is_member(household_id));
create policy member_rules on categorization_rules
    using (is_member(household_id)) with check (is_member(household_id));

-- Hồ sơ: thấy chính mình + người cùng hộ; chỉ tự sửa/tạo hồ sơ của mình.
create function shares_household(p_user uuid) returns boolean as $$
    select exists (
        select 1 from household_members a
        join household_members b on a.household_id = b.household_id
        where a.user_id = current_user_id() and b.user_id = p_user
    );
$$ language sql security definer stable;

create policy member_users_select on public.users
    for select using (email = auth.jwt()->>'email' or shares_household(id));
create policy self_users_insert on public.users
    for insert with check (email = auth.jwt()->>'email');
create policy self_users_update on public.users
    for update using (email = auth.jwt()->>'email')
    with check (email = auth.jwt()->>'email');

-- created_by do DB tự gán từ phiên; giữ giá trị seed khi chạy bằng quyền admin.
create function set_created_by() returns trigger as $$
begin
    new.created_by := coalesce(current_user_id(), new.created_by);
    return new;
end; $$ language plpgsql;
create trigger trg_created_by_households before insert on households
    for each row execute function set_created_by();
create trigger trg_created_by_categories before insert on categories
    for each row execute function set_created_by();
create trigger trg_created_by_transactions before insert on transactions
    for each row execute function set_created_by();

-- ----------------------------------------------------------------------------
-- 8 · Seed hộ: A = Alice + Bob · B = Carol (id theo public.users — KHÔNG remap)
-- ----------------------------------------------------------------------------
do $$
declare
    v_a constant uuid := '00000000-0000-0000-0000-000000000001';
    v_b constant uuid := '00000000-0000-0000-0000-000000000002';
    u_alice uuid; u_bob uuid; u_carol uuid;
begin
    select id into u_alice from public.users where email = 'alice@dev.local';
    select id into u_bob   from public.users where email = 'bob@dev.local';
    select id into u_carol from public.users where email = 'carol@dev.local';
    if u_alice is null or u_bob is null then
        raise exception 'Thiếu hồ sơ users dev — phần B/1 chưa chạy?';
    end if;

    insert into households(id, name, created_by) values (v_a, 'Gia đình A', u_alice);
    insert into household_members(household_id, user_id)
        values (v_a, u_alice), (v_a, u_bob);

    if u_carol is not null then
        insert into households(id, name, created_by) values (v_b, 'Gia đình B', u_carol);
        insert into household_members(household_id, user_id) values (v_b, u_carol);
    end if;
end $$;

-- ----------------------------------------------------------------------------
-- 9 · Seed danh mục mặc định cho hộ A và B [0009 · FR-001, R9]
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
        from (values ('Ăn uống'),('Di chuyển'),('Hóa đơn'),('Mua sắm'),('Giải trí'),('Sức khỏe')) as x(name);

        insert into categories (household_id, name, type, is_default, created_by)
        select v_hh, x.name, 'INCOME', true, v_owner
        from (values ('Lương'),('Thưởng')) as x(name);
    end loop;
end $$;

-- ----------------------------------------------------------------------------
-- 10 · Grants cho API roles (PostgREST) — RLS vẫn kiểm soát theo hàng.
-- ----------------------------------------------------------------------------
grant usage on schema public to anon, authenticated;
grant select, insert, update, delete on all tables in schema public to authenticated;
grant execute on all functions in schema public to anon, authenticated;
alter default privileges in schema public
    grant select, insert, update, delete on tables to authenticated;
alter default privileges in schema public
    grant execute on functions to anon, authenticated;

-- ----------------------------------------------------------------------------
-- 11 · Tên hiển thị đẹp cho user dev + kiểm nhanh
-- ----------------------------------------------------------------------------
update public.users
set display_name = initcap(split_part(email, '@', 1))
where email like '%@dev.local';

select 'auth users (dev)' as what, count(*) from auth.users where email like '%@dev.local'
union all
select 'public.users', count(*) from public.users
union all
select 'households', count(*) from households
union all
select 'members', count(*) from household_members
union all
select 'categories', count(*) from categories;
