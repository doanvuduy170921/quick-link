-- name: BatchInsertClick :copyfrom
insert into clicks(short_code,clicked_at)
values ($1,$2);
