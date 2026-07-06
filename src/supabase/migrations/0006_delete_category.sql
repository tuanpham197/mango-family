-- T010 · Safe category deletion (household-scoped) [FR-008][FR-009][FR-012][SC-007]
-- p_action: 'REASSIGN' (move txns to p_target, same type) | 'DELETE' (remove txns)
-- Handles the parent's subcategories under the same protection.

create or replace function delete_category(p_category uuid, p_action text, p_target uuid default null)
returns void as $$
declare v_type text; v_hh uuid;
begin
    select type, household_id into v_type, v_hh from categories where id = p_category;
    if v_hh is null then raise exception 'Category not found'; end if;
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
