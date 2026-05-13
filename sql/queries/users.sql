-- name: CreateUser :exec
INSERT INTO users (id, email, name, password, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;
