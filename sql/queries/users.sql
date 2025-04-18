-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: FindUserByName :one
SELECT * from users 
WHERE name = $1;

-- name: FindUserById :one
SELECT * from users 
WHERE id = $1;

-- name: DeleteAllUsers :exec
DELETE from users;

-- name: GetAllUsers :many
SELECT * from users;
