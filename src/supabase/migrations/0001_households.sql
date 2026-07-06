-- T005 · Households + membership (shared family ledger backbone) [R10][R11]

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
    primary key (household_id, user_id)   -- no role column: all members equal [R11]
);

create index if not exists idx_member_user on household_members(user_id);
