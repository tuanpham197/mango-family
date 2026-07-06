-- T050 · Optimistic concurrency for shared category edits [R13]
-- Members may edit/delete the same category concurrently; writes carry the
-- last-seen updated_at so a stale write matches 0 rows instead of silently
-- overwriting (client notifies the member and reloads).

alter table categories
    add column if not exists updated_at timestamptz not null default now();

create or replace function touch_updated_at() returns trigger as $$
begin
    new.updated_at := now();
    return new;
end; $$ language plpgsql;

drop trigger if exists trg_touch_updated_at on categories;
create trigger trg_touch_updated_at before update on categories
    for each row execute function touch_updated_at();
