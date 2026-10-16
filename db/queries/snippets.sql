-- name: InsertSnippet :exec
INSERT INTO snippets (id, folder_id, title, description, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetSnippet :one
SELECT id, folder_id, title, description, created_at, updated_at, deleted_at, former_folder_path
FROM live_snippets
WHERE id = ?;

-- name: ListLiveSnippets :many
SELECT id, folder_id, title, description, created_at, updated_at, deleted_at, former_folder_path
FROM live_snippets;

-- name: ListLibrarySnippets :many
SELECT id, folder_id, title, description, created_at, updated_at, deleted_at, former_folder_path
FROM snippets;

-- name: CountSnippets :one
SELECT COUNT(*) FROM snippets;

-- name: ListSnippetsInFolderByTitle :many
SELECT id, folder_id, title, description, created_at, updated_at, deleted_at, former_folder_path
FROM live_snippets
WHERE folder_id IS sqlc.narg(folder_id)
ORDER BY title COLLATE NOCASE, id;

-- name: ListSnippetsInFolderByUpdated :many
SELECT id, folder_id, title, description, created_at, updated_at, deleted_at, former_folder_path
FROM live_snippets
WHERE folder_id IS sqlc.narg(folder_id)
ORDER BY updated_at DESC, title COLLATE NOCASE, id;

-- name: ListSnippetsInFolderByCreated :many
SELECT id, folder_id, title, description, created_at, updated_at, deleted_at, former_folder_path
FROM live_snippets
WHERE folder_id IS sqlc.narg(folder_id)
ORDER BY created_at DESC, title COLLATE NOCASE, id;

-- name: ListSnippetsWithTagByTitle :many
SELECT live_snippets.id, live_snippets.folder_id, live_snippets.title, live_snippets.description,
       live_snippets.created_at, live_snippets.updated_at, live_snippets.deleted_at, live_snippets.former_folder_path
FROM live_snippets
JOIN snippet_tag ON snippet_tag.snippet_id = live_snippets.id
WHERE snippet_tag.tag_id = sqlc.arg(tag_id)
ORDER BY live_snippets.title COLLATE NOCASE, live_snippets.id;

-- name: ListSnippetsWithTagByUpdated :many
SELECT live_snippets.id, live_snippets.folder_id, live_snippets.title, live_snippets.description,
       live_snippets.created_at, live_snippets.updated_at, live_snippets.deleted_at, live_snippets.former_folder_path
FROM live_snippets
JOIN snippet_tag ON snippet_tag.snippet_id = live_snippets.id
WHERE snippet_tag.tag_id = sqlc.arg(tag_id)
ORDER BY live_snippets.updated_at DESC, live_snippets.title COLLATE NOCASE, live_snippets.id;

-- name: ListSnippetsWithTagByCreated :many
SELECT live_snippets.id, live_snippets.folder_id, live_snippets.title, live_snippets.description,
       live_snippets.created_at, live_snippets.updated_at, live_snippets.deleted_at, live_snippets.former_folder_path
FROM live_snippets
JOIN snippet_tag ON snippet_tag.snippet_id = live_snippets.id
WHERE snippet_tag.tag_id = sqlc.arg(tag_id)
ORDER BY live_snippets.created_at DESC, live_snippets.title COLLATE NOCASE, live_snippets.id;

-- name: CountSnippetsByFolder :many
SELECT folder_id, COUNT(*) AS snippet_count
FROM live_snippets
GROUP BY folder_id;

-- name: UpdateSnippet :execrows
UPDATE snippets
SET title = sqlc.arg(title), description = sqlc.arg(description), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND updated_at = sqlc.arg(loaded_updated_at);

-- name: DeleteSnippet :execrows
DELETE FROM snippets
WHERE id = ?;

-- name: MoveSnippet :execrows
UPDATE snippets
SET folder_id = ?
WHERE id = ?;
