-- +goose Up
create table transactions (
    id               uuid primary key default gen_random_uuid(),
    household_id     uuid not null references households(id),
    created_by       uuid not null references users(id),
    amount           numeric(14,2) not null check (amount > 0),
    type             text not null check (type in ('INCOME','EXPENSE')),
    category_id      uuid not null references categories(id),
    description      text,
    transaction_date timestamptz not null default now()
);

create index idx_transactions_household on transactions(household_id, transaction_date desc);
create index idx_transactions_category on transactions(category_id);

-- +goose Down
drop table transactions;
