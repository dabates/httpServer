-- name: GetUserFromRefreshToken :one
select * from refresh_tokens
left join users u on u.id = refresh_tokens.user_id
where refresh_tokens.token = $1;