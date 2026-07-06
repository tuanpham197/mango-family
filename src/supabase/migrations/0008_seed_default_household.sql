-- T012 · Seed: DEFAULT household + add 2 users as members  ← yêu cầu tiền đề
-- Dev seed. PREREQUISITE: tạo trước 2 tài khoản qua Supabase Auth (Alice, Bob),
-- rồi chạy migration này — nó lấy 2 user đầu tiên trong auth.users và thêm vào hộ.
-- Hộ mặc định dùng UUID cố định để các seed/khởi tạo khác tham chiếu được.

do $$
declare
    v_hh    uuid := '00000000-0000-0000-0000-000000000001';  -- DEFAULT household id
    v_users uuid[];
begin
    -- Lấy tối đa 2 user hiện có (dev seed)
    select array(select id from auth.users order by created_at limit 2) into v_users;

    if v_users is null or array_length(v_users, 1) is null then
        raise notice 'No auth.users found — tạo 2 user qua Auth rồi chạy lại migration seed này.';
        return;
    end if;

    -- Tạo hộ mặc định (created_by = user đầu tiên)
    insert into households(id, name, created_by)
    values (v_hh, 'Hộ gia đình mặc định', v_users[1])
    on conflict (id) do nothing;

    -- Thêm thành viên 1
    insert into household_members(household_id, user_id)
    values (v_hh, v_users[1])
    on conflict do nothing;

    -- Thêm thành viên 2 (nếu có)
    if array_length(v_users, 1) >= 2 then
        insert into household_members(household_id, user_id)
        values (v_hh, v_users[2])
        on conflict do nothing;
    else
        raise notice 'Chỉ có 1 user — tạo thêm 1 user nữa để đủ 2 thành viên cho hộ mặc định.';
    end if;
end $$;
