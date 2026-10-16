-- name: InsertSnippetTag :exec
INSERT INTO snippet_tag (snippet_id, tag_id)
VALUES (?, ?);

-- name: DeleteSnippetTags :exec
DELETE FROM snippet_tag
WHERE snippet_id = ?;

-- name: CountSnippetsByTag :many
SELECT snippet_tag.tag_id, COUNT(*) AS snippet_count
FROM snippet_tag
JOIN live_snippets ON live_snippets.id = snippet_tag.snippet_id
GROUP BY snippet_tag.tag_id;

-- name: CountSnippetsWithTag :one
SELECT COUNT(*)
FROM snippet_tag
JOIN live_snippets ON live_snippets.id = snippet_tag.snippet_id
WHERE snippet_tag.tag_id = ?;

-- name: MoveSnippetTags :exec
INSERT OR IGNORE INTO snippet_tag (snippet_id, tag_id)
SELECT moved.snippet_id, sqlc.arg(to_tag_id)
FROM snippet_tag AS moved
WHERE moved.tag_id = sqlc.arg(from_tag_id);
