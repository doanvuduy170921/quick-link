-- name: CreateURL :one
insert into urls(
                    short_code,original_url,expires_at,created_at
)
values ($1,$2,$3,NOW())
returning *;


-- name: GetOriginalUrlByShortCode :one
SELECT original_url,expires_at
FROM urls
WHERE short_code = sqlc.arg(short_code)
AND (expires_at > NOW() OR expires_at IS NULL);