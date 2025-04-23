-- name: CreateDataset :one
INSERT INTO datasets (
    id,
    user_id,
    name,
    description,
    created_at,
    updated_at
) VALUES (
    sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(name), sqlc.arg(description), sqlc.arg(created_at), sqlc.arg(updated_at)
)
RETURNING *;

-- name: GetDatasetByID :one
SELECT * FROM datasets
WHERE id = sqlc.arg(id);

-- name: ListDatasetsForUser :many
SELECT * FROM datasets
WHERE user_id = sqlc.arg(id)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateDataset :one
UPDATE datasets
SET
    name = sqlc.arg(name),
    description = sqlc.arg(description),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteDataset :exec
DELETE FROM datasets
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: SearchDatasetByName :many
SELECT * FROM datasets
WHERE user_id = $1
  AND (
    $2::text IS NULL OR name ILIKE '%' || $2::text || '%'
  )
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;
