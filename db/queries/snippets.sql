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

-- name: ListSnippetsWithTagByTitle :many
SELECT snippets.id, snippets.folder_id, snippets.title, snippets.description, snippets.created_at, snippets.updated_at
FROM snippets
JOIN snippet_tag ON snippet_tag.snippet_id = snippets.id
WHERE snippet_tag.tag_id = sqlc.arg(tag_id)
ORDER BY snippets.title COLLATE NOCASE, snippets.id;

-- name: ListSnippetsWithTagByUpdated :many
SELECT snippets.id, snippets.folder_id, snippets.title, snippets.description, snippets.created_at, snippets.updated_at
FROM snippets
JOIN snippet_tag ON snippet_tag.snippet_id = snippets.id
WHERE snippet_tag.tag_id = sqlc.arg(tag_id)
ORDER BY snippets.updated_at DESC, snippets.title COLLATE NOCASE, snippets.id;

-- name: ListSnippetsWithTagByCreated :many
SELECT snippets.id, snippets.folder_id, snippets.title, snippets.description, snippets.created_at, snippets.updated_at
FROM snippets
JOIN snippet_tag ON snippet_tag.snippet_id = snippets.id
WHERE snippet_tag.tag_id = sqlc.arg(tag_id)
ORDER BY snippets.created_at DESC, snippets.title COLLATE NOCASE, snippets.id;

-- name: CountSnippetsByFolder :many
SELECT folder_id, COUNT(*) AS snippet_count
FROM snippets
GROUP BY folder_id;

-- name: UpdateSnippet :execrows
UPDATE snippets
SET title = sqlc.arg(title), description = sqlc.arg(description), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND updated_at = sqlc.arg(loaded_updated_at);

-- name: DeleteSnippet :execrows
DELETE FROM snippets
WHERE id = ?;
