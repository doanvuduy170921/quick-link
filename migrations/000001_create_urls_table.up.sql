create table if not exists urls(
    id bigserial primary key,
    short_code varchar(10) not null unique,
    original_url text not null,
    created_at timestamptz not null default current_timestamp,
    expires_at timestamptz null
);
