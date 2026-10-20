-- name: InsertTag :exec
INSERT INTO tags (id, name, name_key, created_at, updated_at)
VALUES (?, ?, ?, ?, ?);

-- name: InsertTagUnlessStored :exec
INSERT INTO tags (id, name, name_key, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (id) DO NOTHING;

-- name: ListTags :many
SELECT id, name, name_key, created_at, updated_at
FROM tags;

-- name: CountTags :one
SELECT COUNT(*) FROM tags;

-- name: GetTag :one
SELECT id, name, name_key, created_at, updated_at
FROM tags
WHERE id = ?;

-- name: ListOtherTagsWithNameKey :many
SELECT id, name, name_key, created_at, updated_at
FROM tags
WHERE name_key = sqlc.arg(name_key) AND id <> sqlc.arg(id);

-- name: UpdateTag :execrows
UPDATE tags
SET name = ?, name_key = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteTag :execrows
DELETE FROM tags
WHERE id = ?;

-- name: ListLiveSnippetTags :many
SELECT snippet_tag.snippet_id, sqlc.embed(tags)
FROM snippet_tag
JOIN tags ON tags.id = snippet_tag.tag_id
JOIN live_snippets ON live_snippets.id = snippet_tag.snippet_id;

-- name: ListLibrarySnippetTags :many
SELECT snippet_tag.snippet_id, sqlc.embed(tags)
FROM snippet_tag
JOIN tags ON tags.id = snippet_tag.tag_id;

-- name: ListTrashSnippetTags :many
SELECT snippet_tag.snippet_id, sqlc.embed(tags)
FROM snippet_tag
JOIN tags ON tags.id = snippet_tag.tag_id
WHERE snippet_tag.snippet_id NOT IN (SELECT live_snippets.id FROM live_snippets);

-- name: ListSnippetTrashItemTags :many
SELECT snippet_tag.snippet_id, sqlc.embed(tags)
FROM snippet_tag
JOIN tags ON tags.id = snippet_tag.tag_id
WHERE snippet_tag.snippet_id = ?;

-- name: ListSnippetTagsBySnippet :many
SELECT snippet_tag.snippet_id, sqlc.embed(tags)
FROM snippet_tag
JOIN tags ON tags.id = snippet_tag.tag_id
JOIN live_snippets ON live_snippets.id = snippet_tag.snippet_id
WHERE snippet_tag.snippet_id = ?;

-- name: ListSnippetTagsInFolder :many
SELECT snippet_tag.snippet_id, sqlc.embed(tags)
FROM snippet_tag
JOIN tags ON tags.id = snippet_tag.tag_id
JOIN live_snippets ON live_snippets.id = snippet_tag.snippet_id
WHERE live_snippets.folder_id IS sqlc.narg(folder_id);

-- name: ListSnippetTagsWithTag :many
SELECT snippet_tag.snippet_id, sqlc.embed(tags)
FROM snippet_tag
JOIN tags ON tags.id = snippet_tag.tag_id
JOIN live_snippets ON live_snippets.id = snippet_tag.snippet_id
WHERE snippet_tag.snippet_id IN (SELECT carrying.snippet_id FROM snippet_tag AS carrying WHERE carrying.tag_id = sqlc.arg(tag_id));
