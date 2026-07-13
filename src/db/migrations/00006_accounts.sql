-- +goose Up
create table accounts (
    id              uuid primary key default gen_random_uuid(),
    household_id    uuid not null references households(id),
    name            text not null,
    type            text not null default 'CASH'
                        check (type in ('CASH','BANK','EWALLET','CREDIT')),
    initial_balance numeric(14,2) not null default 0,
    created_by      uuid references users(id),
    created_at      timestamptz not null default now()
);
create index idx_accounts_household on accounts(household_id);

-- Ghi chú: view account_balances (D14) tạo ở 00007 SAU khi transactions.account_id
-- tồn tại — view tham chiếu cột đó nên không thể tạo ở đây.

-- +goose Down
drop table accounts;
