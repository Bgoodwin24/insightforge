-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, username, email, password_hash)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateLoginAttempts :exec
UPDATE users
SET failed_login_attempts = $2,
    last_failed_attempt = $3,
    locked_until = $4
WHERE id = $1;

-- name: ResetLoginAttempts :exec
UPDATE users
SET failed_login_attempts = 0,
    last_failed_attempt = NULL,
    locked_until = NULL
WHERE id = $1;
