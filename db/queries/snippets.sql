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
