-- name: UpdateUserRedStatus :one
update users
set is_chirpy_red= true,
    updated_at   = now()
where id = $1 returning *;