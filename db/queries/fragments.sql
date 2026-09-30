-- name: InsertFragment :exec
INSERT INTO fragments (id, snippet_id, position, language, content, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListFragmentsBySnippet :many
SELECT id, snippet_id, position, language, content, created_at, updated_at
FROM fragments
WHERE snippet_id = ?
ORDER BY position;

-- name: ListFragments :many
SELECT id, snippet_id, position, language, content, created_at, updated_at
FROM fragments
ORDER BY snippet_id, position;
