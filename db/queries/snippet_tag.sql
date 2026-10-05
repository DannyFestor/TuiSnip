-- name: InsertSnippetTag :exec
INSERT INTO snippet_tag (snippet_id, tag_id)
VALUES (?, ?);

-- name: CountSnippetsByTag :many
SELECT tag_id, COUNT(*) AS snippet_count
FROM snippet_tag
GROUP BY tag_id;
