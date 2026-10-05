-- name: InsertFolder :exec
INSERT INTO folders (id, parent_id, name, default_language, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetFolder :one
SELECT id, parent_id, name, default_language, created_at, updated_at
FROM folders
WHERE id = ?;

-- name: ListFolders :many
SELECT id, parent_id, name, default_language, created_at, updated_at
FROM folders;

-- name: UpdateFolder :execrows
UPDATE folders
SET name = ?, default_language = ?, updated_at = ?
WHERE id = ?;
