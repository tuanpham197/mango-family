-- T011 · RLS policies: members share their household's data; isolated across households
-- [FR-018 updated][R10][R12] — equal permissions: same policy for read & write [R11]

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
