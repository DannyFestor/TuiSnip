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

-- name: ListSnippetsInFolderByTitle :many
SELECT id, folder_id, title, description, created_at, updated_at
FROM snippets
WHERE folder_id IS sqlc.narg(folder_id)
ORDER BY title COLLATE NOCASE, id;

-- name: ListSnippetsInFolderByUpdated :many
SELECT id, folder_id, title, description, created_at, updated_at
FROM snippets
WHERE folder_id IS sqlc.narg(folder_id)
ORDER BY updated_at DESC, title COLLATE NOCASE, id;

-- name: ListSnippetsInFolderByCreated :many
SELECT id, folder_id, title, description, created_at, updated_at
FROM snippets
WHERE folder_id IS sqlc.narg(folder_id)
ORDER BY created_at DESC, title COLLATE NOCASE, id;

-- name: CountSnippetsByFolder :many
SELECT folder_id, COUNT(*) AS snippet_count
FROM snippets
GROUP BY folder_id;

-- name: UpdateSnippet :execrows
UPDATE snippets
SET title = sqlc.arg(title), description = sqlc.arg(description), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND updated_at = sqlc.arg(loaded_updated_at);
