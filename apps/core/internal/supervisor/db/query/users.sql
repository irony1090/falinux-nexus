-- name: CreateUser :one
INSERT INTO users (
    identification,
    password,
    nickname
) VALUES ( $1, $2, $3 )
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByIdentification :one
SELECT * FROM users WHERE identification = $1;
