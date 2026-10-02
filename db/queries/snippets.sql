-- name: InsertSnippet :exec
INSERT INTO snippets (id, folder_id, title, description, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetSnippet :one
SELECT id, folder_id, title, description, created_at, updated_at
FROM snippets
WHERE id = ?;

-- name: ListSnippets :many
SELECT id, folder_id, title, description, created_at, updated_at
FROM snippets;

-- name: ListSnippetsInFolder :many
SELECT id, folder_id, title, description, created_at, updated_at
FROM snippets
WHERE folder_id IS sqlc.narg(folder_id)
ORDER BY title COLLATE NOCASE, id;
