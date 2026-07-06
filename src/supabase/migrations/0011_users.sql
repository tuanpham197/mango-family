-- 0011 · public.users — bảng người dùng ĐỘC LẬP [Feature 002 · FR-015, FR-016]
-- users có khóa chính riêng (KHÔNG tham chiếu auth.users). Đây là nguồn định
-- danh duy nhất của app: household_members.user_id và mọi cột created_by đều
-- FK về users(id). Phiên đăng nhập (Supabase Auth) được nối với users qua
-- EMAIL — mật khẩu/credentials vẫn do dịch vụ xác thực quản lý, không nằm ở đây.

create table if not exists public.users (
    id           uuid primary key default gen_random_uuid(),  -- độc lập
    email        text not null unique,
    display_name text not null,
    created_at   timestamptz not null default now()
);

alter table public.users enable row level security;
grant select, insert, update on public.users to authenticated;

-- Backfill: mỗi tài khoản đăng nhập hiện có được cấp một hồ sơ users (id MỚI).
insert into public.users (email, display_name)
select email, coalesce(nullif(split_part(email, '@', 1), ''), 'user')
from auth.users
on conflict (email) do nothing;

-- Cầu nối phiên đăng nhập → users: tra theo email trong JWT.
create or replace function current_user_id() returns uuid as $$
    select id from public.users where email = auth.jwt()->>'email';
$$ language sql security definer stable;

-- Remap DỮ LIỆU hiện có: giá trị auth-uid cũ → users.id mới (nối qua email).
update household_members m set user_id = u.id
    from auth.users a join public.users u on u.email = a.email
    where m.user_id = a.id;
update households h set created_by = u.id
    from auth.users a join public.users u on u.email = a.email
    where h.created_by = a.id;
update categories c set created_by = u.id
    from auth.users a join public.users u on u.email = a.email
    where c.created_by = a.id;
update transactions t set created_by = u.id
    from auth.users a join public.users u on u.email = a.email
    where t.created_by = a.id;

-- Dọn bản ghi KHÔNG ánh xạ được (vd trỏ tới tài khoản auth đã bị xóa) —
-- nếu bỏ qua, bước ADD CONSTRAINT bên dưới sẽ fail với lỗi 23503.
insert into public.users (email, display_name)
values ('unknown@local', 'Không rõ')
on conflict (email) do nothing;                       -- hồ sơ giữ chỗ cho audit
update transactions t
    set created_by = (select id from public.users where email = 'unknown@local')
    where not exists (select 1 from public.users u where u.id = t.created_by);
update households h
    set created_by = (select id from public.users where email = 'unknown@local')
    where not exists (select 1 from public.users u where u.id = h.created_by);
update categories c set created_by = null
    where created_by is not null
      and not exists (select 1 from public.users u where u.id = c.created_by);
delete from household_members m                        -- tài khoản đã xóa ⇒ hết tư cách thành viên
    where not exists (select 1 from public.users u where u.id = m.user_id);

-- Repoint FK: auth.users → public.users.
alter table households        drop constraint if exists households_created_by_fkey;
alter table households        add constraint households_created_by_fkey
    foreign key (created_by) references public.users(id);
alter table household_members drop constraint if exists household_members_user_id_fkey;
alter table household_members add constraint household_members_user_id_fkey
    foreign key (user_id) references public.users(id) on delete cascade;
alter table categories        drop constraint if exists categories_created_by_fkey;
alter table categories        add constraint categories_created_by_fkey
    foreign key (created_by) references public.users(id);
alter table transactions      drop constraint if exists transactions_created_by_fkey;
alter table transactions      add constraint transactions_created_by_fkey
    foreign key (created_by) references public.users(id);

-- is_member: membership giờ tính theo users.id (qua current_user_id).
create or replace function is_member(p_household uuid) returns boolean as $$
    select exists (
        select 1 from household_members
        where household_id = p_household and user_id = current_user_id()
    );
$$ language sql security definer stable;

-- member_self (household_members): user_id không còn là auth.uid().
drop policy if exists member_self on household_members;
create policy member_self on household_members
    using (user_id = current_user_id() or is_member(household_id));

-- created_by do DB tự gán từ phiên đăng nhập (client không cần biết users.id);
-- giữ giá trị seed khi chạy bằng quyền admin (current_user_id() = null).
create or replace function set_created_by() returns trigger as $$
begin
    new.created_by := coalesce(current_user_id(), new.created_by);
    return new;
end; $$ language plpgsql;
drop trigger if exists trg_created_by_households on households;
create trigger trg_created_by_households before insert on households
    for each row execute function set_created_by();
drop trigger if exists trg_created_by_categories on categories;
create trigger trg_created_by_categories before insert on categories
    for each row execute function set_created_by();
drop trigger if exists trg_created_by_transactions on transactions;
create trigger trg_created_by_transactions before insert on transactions
    for each row execute function set_created_by();

-- Hiển thị hồ sơ: chính mình (theo email đăng nhập) + người cùng hộ.
create or replace function shares_household(p_user uuid) returns boolean as $$
    select exists (
        select 1 from household_members a
        join household_members b on a.household_id = b.household_id
        where a.user_id = current_user_id() and b.user_id = p_user
    );
$$ language sql security definer stable;

drop policy if exists member_users_select on public.users;
create policy member_users_select on public.users
    for select using (email = auth.jwt()->>'email' or shares_household(id));
drop policy if exists self_users_insert on public.users;
create policy self_users_insert on public.users
    for insert with check (email = auth.jwt()->>'email');  -- tự tạo hồ sơ lần đầu
drop policy if exists self_users_update on public.users;
create policy self_users_update on public.users
    for update using (email = auth.jwt()->>'email')
    with check (email = auth.jwt()->>'email');
