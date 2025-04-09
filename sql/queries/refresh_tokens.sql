-- name: CreateRefreshToken :one
insert into refresh_tokens (token, user_id, expires_at, created_at, updated_at)
values ($1,
        $2,
        $3,
        now(),
        now())
returning *;