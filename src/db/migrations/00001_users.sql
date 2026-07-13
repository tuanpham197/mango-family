-- +goose Up
create extension if not exists pgcrypto;

create table users (
    id            uuid primary key default gen_random_uuid(),
    email         text not null unique,
    display_name  text not null,
    password_hash text not null,
    created_at    timestamptz not null default now()
);

-- +goose Down
drop table users;
