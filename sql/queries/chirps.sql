-- name: CreateChirp :one
insert into chirps (id, body,user_id,created_at,updated_at)
    values(
    gen_random_uuid(),
    $1,
    $2,
    now(),
    now()
)
    returning *;