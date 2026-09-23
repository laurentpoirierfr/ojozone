-- name: CreateImport :one
INSERT INTO imports (resource_type, created_by, line_count)
VALUES (sqlc.arg('resource_type')::text, sqlc.narg('created_by')::uuid, sqlc.arg('line_count')::integer)
RETURNING *;

-- name: AddImportRow :exec
INSERT INTO import_rows (import_id, line_number, payload)
VALUES (sqlc.arg('import_id')::uuid, sqlc.arg('line_number')::integer, sqlc.arg('payload')::jsonb);

-- name: GetImport :one
SELECT * FROM imports WHERE id = sqlc.arg('id')::uuid;

-- name: ListImports :many
SELECT * FROM imports ORDER BY created_at DESC LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;

-- name: ListImportRows :many
SELECT id, line_number, payload, valid, error
FROM import_rows
WHERE import_id = sqlc.arg('import_id')::uuid
ORDER BY line_number;

-- name: SetImportRowStatus :exec
UPDATE import_rows
SET valid = sqlc.arg('valid')::boolean, error = sqlc.narg('error')::text
WHERE import_id = sqlc.arg('import_id')::uuid AND line_number = sqlc.arg('line_number')::integer;

-- name: MarkImportValidated :exec
UPDATE imports
SET status = 'validated',
    valid_count = sqlc.arg('valid_count')::integer,
    invalid_count = sqlc.arg('invalid_count')::integer,
    report = sqlc.arg('report')::jsonb,
    validated_at = now()
WHERE id = sqlc.arg('id')::uuid;

-- name: MarkImportPublished :exec
UPDATE imports
SET status = 'published',
    report = sqlc.arg('report')::jsonb,
    published_at = now()
WHERE id = sqlc.arg('id')::uuid;