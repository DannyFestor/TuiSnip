-- name: InsertFragment :exec
INSERT INTO fragments (id, snippet_id, position, language, content, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateFragment :exec
UPDATE fragments
SET language = ?, content = ?, updated_at = ?
WHERE id = ?;

-- name: ListFragmentsBySnippet :many
SELECT id, snippet_id, position, language, content, created_at, updated_at
FROM fragments
WHERE snippet_id = ?
ORDER BY position;

-- name: ListFragments :many
SELECT id, snippet_id, position, language, content, created_at, updated_at
FROM fragments
ORDER BY snippet_id, position;

-- name: ListFragmentsInFolder :many
SELECT fragments.id, fragments.snippet_id, fragments.position, fragments.language, fragments.content,
       fragments.created_at, fragments.updated_at
FROM fragments
JOIN snippets ON snippets.id = fragments.snippet_id
WHERE snippets.folder_id IS sqlc.narg(folder_id)
ORDER BY fragments.snippet_id, fragments.position;

-- name: ListFragmentsWithTag :many
SELECT fragments.id, fragments.snippet_id, fragments.position, fragments.language, fragments.content,
       fragments.created_at, fragments.updated_at
FROM fragments
JOIN snippet_tag ON snippet_tag.snippet_id = fragments.snippet_id
WHERE snippet_tag.tag_id = sqlc.arg(tag_id)
ORDER BY fragments.snippet_id, fragments.position;
