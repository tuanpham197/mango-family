-- +goose Up
-- 003 · budget_alerts [FR-007][FR-008][FR-009][D23] — máy trạng thái cảnh báo, khóa theo
-- (budget, period_key, level). period_key reset mỗi kỳ nên state tự sạch khi sang kỳ.
create table budget_alerts (
    id          uuid primary key default gen_random_uuid(),
    budget_id   uuid not null references budgets(id) on delete cascade,
    period_key  text not null,
    level       text not null check (level in ('THRESHOLD_80','OVER_100')),
    over_amount numeric(14,2),
    fired_at    timestamptz not null default now()
);
create unique index uq_budget_alert_level
    on budget_alerts(budget_id, period_key, level);
create index idx_budget_alerts_budget on budget_alerts(budget_id);

-- +goose Down
drop table budget_alerts;
