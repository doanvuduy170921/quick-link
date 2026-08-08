-- name: CreateURL :one
insert into urls(
                    short_code,original_url,created_at
)
values ($1,$2,NOW())
returning *;