-- name: UpdateUser :one
update users
set email    = $2,
    hashed_password = $3,
    updated_at = now()
where id = $1
returning *;