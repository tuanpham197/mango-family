-- T013 · Seed default categories for the DEFAULT household [FR-001][R9]
-- (assumption — danh sách cuối cùng chờ nghiệp vụ chốt; is_default = true, dùng chung cả hộ)
-- Depends on 0008 (default household must exist).

do $$
declare
    v_hh    uuid := '00000000-0000-0000-0000-000000000001';
    v_owner uuid;
begin
    if not exists (select 1 from households where id = v_hh) then
        raise notice 'Default household chưa tồn tại — chạy 0008_seed_default_household.sql trước.';
        return;
    end if;
    select created_by into v_owner from households where id = v_hh;

    -- EXPENSE (Chi)
    insert into categories (household_id, name, type, is_default, created_by)
    select v_hh, x.name, 'EXPENSE', true, v_owner
    from (values ('Ăn uống'),('Di chuyển'),('Hóa đơn'),('Mua sắm'),('Giải trí'),('Sức khỏe')) as x(name)
    where not exists (
        select 1 from categories c
        where c.household_id = v_hh and c.type = 'EXPENSE' and c.name = x.name and c.parent_id is null
    );

    -- INCOME (Thu)
    insert into categories (household_id, name, type, is_default, created_by)
    select v_hh, x.name, 'INCOME', true, v_owner
    from (values ('Lương'),('Thưởng')) as x(name)
    where not exists (
        select 1 from categories c
        where c.household_id = v_hh and c.type = 'INCOME' and c.name = x.name and c.parent_id is null
    );
end $$;
