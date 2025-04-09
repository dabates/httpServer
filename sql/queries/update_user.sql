-- name: UpdateUser :one
update users
set email    = $2,
    hashed_password = $3
where id = $1
returning *;