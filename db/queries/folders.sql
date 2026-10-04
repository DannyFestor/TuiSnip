-- name: InsertFolder :exec
INSERT INTO folders (id, parent_id, name, default_language, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListFolders :many
SELECT id, parent_id, name, default_language, created_at, updated_at
FROM folders;
