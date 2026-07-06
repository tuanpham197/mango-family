-- T006 · Membership helper + enable RLS on all tables [R12]

create or replace function is_member(p_household uuid) returns boolean as $$
    select exists (
        select 1 from household_members
        where household_id = p_household and user_id = auth.uid()
    );
$$ language sql security definer stable;

alter table households            enable row level security;
alter table household_members     enable row level security;
-- categories / transactions / categorization_rules: RLS enabled in their own migrations.
