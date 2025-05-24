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

-- name: CreateDatasetField :exec
INSERT INTO dataset_fields (id, dataset_id, name, data_type, description, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: CreateDatasetRecord :exec
INSERT INTO dataset_records (id, dataset_id, created_at, updated_at)
VALUES ($1, $2, $3, $4);

-- name: CreateRecordValue :exec
INSERT INTO record_values (record_id, field_id, value)
VALUES ($1, $2, $3);

-- name: GetDatasetFields :many
SELECT * FROM dataset_fields
WHERE dataset_id = $1
ORDER BY name;

-- name: GetDatasetRecords :many
SELECT * FROM dataset_records
WHERE dataset_id = $1
ORDER BY created_at;

-- name: GetRecordValuesByRecordID :many
SELECT * FROM record_values
WHERE record_id = $1;


-- name: GetFieldsByDatasetID :many
SELECT id, name, data_type, description, created_at, dataset_id
FROM dataset_fields
WHERE dataset_id = $1
ORDER BY created_at ASC;

-- name: GetDatasetFieldsForDataset :many
SELECT name FROM dataset_fields WHERE dataset_id = $1 ORDER BY created_at;

-- name: GetRecordsByDatasetID :many
SELECT id, dataset_id, created_at, updated_at
FROM dataset_records
WHERE dataset_id = $1
ORDER BY created_at ASC;

-- name: GetRecordValuesByDatasetID :many
SELECT record_id, field_id, value
FROM record_values
WHERE record_id IN (
    SELECT id FROM dataset_records WHERE dataset_id = $1
);

-- name: UpdateDatasetRows :exec
WITH updated AS (
    UPDATE dataset_records
    SET updated_at = $2
    WHERE dataset_id = $1
    RETURNING id
)
UPDATE record_values
SET value = $3
WHERE record_id IN (SELECT id FROM updated) AND field_id = $4;

-- name: GetDatasetField :one
SELECT * FROM dataset_fields WHERE id = $1 AND dataset_id = $2;

-- name: DeleteDatasetField :exec
DELETE FROM dataset_fields
WHERE id = $1 AND dataset_id = $2;
