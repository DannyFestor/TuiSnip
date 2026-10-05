-- name: InsertSnippetTag :exec
INSERT INTO snippet_tag (snippet_id, tag_id)
VALUES (?, ?);

-- name: CountSnippetsByTag :many
SELECT tag_id, COUNT(*) AS snippet_count
FROM snippet_tag
GROUP BY tag_id;

-- name: CountSnippetsWithTag :one
SELECT COUNT(*)
FROM snippet_tag
WHERE tag_id = ?;

-- name: MoveSnippetTags :exec
INSERT OR IGNORE INTO snippet_tag (snippet_id, tag_id)
SELECT moved.snippet_id, sqlc.arg(to_tag_id)
FROM snippet_tag AS moved
WHERE moved.tag_id = sqlc.arg(from_tag_id);
