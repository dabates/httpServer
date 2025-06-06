// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: update_user_to_red.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const updateUserRedStatus = `-- name: UpdateUserRedStatus :one
update users
set is_chirpy_red= true,
    updated_at   = now()
where id = $1 returning id, email, created_at, updated_at, hashed_password, is_chirpy_red
`

func (q *Queries) UpdateUserRedStatus(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRowContext(ctx, updateUserRedStatus, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.HashedPassword,
		&i.IsChirpyRed,
	)
	return i, err
}
