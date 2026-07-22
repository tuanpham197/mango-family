-- +goose Up
-- 003 · budgets [FR-001..004][FR-010..013][D20][D22][D25] — xem contracts/db-schema.sql.
-- Ngân sách chỉ ĐỌC transactions/categories (tiến độ suy ra ở biz — D21). Không RLS/trigger.
create table budgets (
    id            uuid primary key default gen_random_uuid(),
    household_id  uuid not null references households(id),
    type          text not null check (type in ('CATEGORY','TOTAL')),
    category_id   uuid references categories(id),
    limit_amount  numeric(14,2) not null check (limit_amount > 0),
    period_type   text not null check (period_type in ('MONTHLY','WEEKLY','ONE_TIME')),
    start_date    date,
    end_date      date,
    status        text not null default 'ACTIVE' check (status in ('ACTIVE','ENDED')),
    created_by    uuid not null references users(id),
    created_at    timestamptz not null default now(),
    updated_at    timestamptz not null default now(),
    constraint budget_category_presence check (
        (type = 'CATEGORY' and category_id is not null) or
        (type = 'TOTAL'    and category_id is null)
    ),
    constraint budget_period_range check (
        end_date is null or start_date is null or end_date >= start_date
    )
);
create index idx_budgets_household on budgets(household_id);

-- Uniqueness (FR-010, D22): partial → ngân sách ENDED không cản tạo mới.
create unique index uq_budget_category_active
    on budgets(household_id, category_id, period_type)
    where status = 'ACTIVE' and type = 'CATEGORY';
create unique index uq_budget_total_active
    on budgets(household_id, period_type)
    where status = 'ACTIVE' and type = 'TOTAL';

-- +goose Down
drop table budgets;
