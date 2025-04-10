-- name: DeleteChirp :exec
delete
from chirps
where id = $1
  and user_id = $2;