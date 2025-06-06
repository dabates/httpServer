// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: get_user_from_refresh_token.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const getUserFromRefreshToken = `-- name: GetUserFromRefreshToken :one
select token, user_id, expires_at, revoked_at, refresh_tokens.created_at, refresh_tokens.updated_at, id, email, u.created_at, u.updated_at, hashed_password, is_chirpy_red from refresh_tokens
left join users u on u.id = refresh_tokens.user_id
where refresh_tokens.token = $1
`

type GetUserFromRefreshTokenRow struct {
	Token          string
	UserID         uuid.UUID
	ExpiresAt      sql.NullTime
	RevokedAt      sql.NullTime
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ID             uuid.NullUUID
	Email          sql.NullString
	CreatedAt_2    sql.NullTime
	UpdatedAt_2    sql.NullTime
	HashedPassword sql.NullString
	IsChirpyRed    sql.NullBool
}

func (q *Queries) GetUserFromRefreshToken(ctx context.Context, token string) (GetUserFromRefreshTokenRow, error) {
	row := q.db.QueryRowContext(ctx, getUserFromRefreshToken, token)
	var i GetUserFromRefreshTokenRow
	err := row.Scan(
		&i.Token,
		&i.UserID,
		&i.ExpiresAt,
		&i.RevokedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ID,
		&i.Email,
		&i.CreatedAt_2,
		&i.UpdatedAt_2,
		&i.HashedPassword,
		&i.IsChirpyRed,
	)
	return i, err
}
