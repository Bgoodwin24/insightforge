-- name: CreatePendingUser :one
INSERT INTO pending_users (id, email, username, password_hash, token, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPendingUserByToken :one
SELECT * FROM pending_users 
WHERE token = $1 AND created_at > NOW() - INTERVAL '1 hour';

-- name: DeletePendingUserByID :exec
DELETE FROM pending_users WHERE id = $1;
