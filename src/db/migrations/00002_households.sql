-- +goose Up
create table households (
    id         uuid primary key default gen_random_uuid(),
    name       text not null,
    created_by uuid not null references users(id),
    created_at timestamptz not null default now()
);

create table household_members (
    household_id uuid not null references households(id),
    user_id      uuid not null references users(id),
    joined_at    timestamptz not null default now(),
    primary key (household_id, user_id)
);

-- +goose Down
drop table household_members;
drop table households;
