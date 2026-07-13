-- +goose Up
-- account_id: thêm nullable → backfill về tài khoản mặc định (sớm nhất) của hộ → siết NOT NULL.
-- Trên DB mới (chưa có giao dịch) backfill là no-op; NOT NULL áp trơn.
alter table transactions add column account_id uuid references accounts(id) on delete restrict;

-- +goose StatementBegin
update transactions t
set account_id = (
    select a.id from accounts a
    where a.household_id = t.household_id
    order by a.created_at
    limit 1
)
where t.account_id is null;
-- +goose StatementEnd

alter table transactions alter column account_id set not null;

alter table transactions add column updated_at timestamptz not null default now();

create index idx_transactions_account on transactions(account_id);

-- Số dư suy ra — luôn = initial_balance + tổng bút toán có dấu (SC-004, D14).
-- Tạo ở đây (không phải 00006) vì view tham chiếu transactions.account_id vừa thêm.
-- +goose StatementBegin
create view account_balances as
select
    a.id           as account_id,
    a.household_id as household_id,
    a.initial_balance
      + coalesce(sum(case t.type when 'INCOME' then t.amount else -t.amount end), 0)
                   as balance
from accounts a
left join transactions t on t.account_id = a.id
group by a.id, a.household_id, a.initial_balance;
-- +goose StatementEnd

-- +goose Down
drop view if exists account_balances;
drop index if exists idx_transactions_account;
alter table transactions drop column updated_at;
alter table transactions drop column account_id;
