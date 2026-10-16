-- name: InsertFolder :exec
INSERT INTO folders (id, parent_id, name, default_language, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetFolder :one
SELECT id, parent_id, name, default_language, created_at, updated_at, deleted_at, former_parent_path
FROM live_folders
WHERE id = ?;

-- name: ListLiveFolders :many
SELECT id, parent_id, name, default_language, created_at, updated_at, deleted_at, former_parent_path
FROM live_folders;

-- name: ListLibraryFolders :many
SELECT id, parent_id, name, default_language, created_at, updated_at, deleted_at, former_parent_path
FROM folders;

-- name: CountFolders :one
SELECT COUNT(*) FROM folders;

-- name: CountSubfolders :one
WITH RECURSIVE subtree (id) AS (
    SELECT live_folders.id FROM live_folders WHERE live_folders.id = sqlc.arg(id)
    UNION ALL
    SELECT live_folders.id FROM live_folders JOIN subtree ON live_folders.parent_id = subtree.id
)
SELECT COUNT(*) - 1 FROM subtree;

-- name: CountSnippetsInSubtree :one
WITH RECURSIVE subtree (id) AS (
    SELECT live_folders.id FROM live_folders WHERE live_folders.id = sqlc.arg(id)
    UNION ALL
    SELECT live_folders.id FROM live_folders JOIN subtree ON live_folders.parent_id = subtree.id
)
SELECT COUNT(*) FROM live_snippets WHERE live_snippets.folder_id IN (SELECT subtree.id FROM subtree);

-- name: DeleteFolder :execrows
DELETE FROM folders
WHERE id = ?;

-- name: UpdateFolder :execrows
UPDATE folders
SET name = ?, default_language = ?, updated_at = ?
WHERE id = ?;

-- name: ListDescendantFolderIDs :many
WITH RECURSIVE subtree (id) AS (
    SELECT folders.id FROM folders WHERE folders.parent_id = sqlc.arg(id)
    UNION
    SELECT folders.id FROM folders JOIN subtree ON folders.parent_id = subtree.id
)
SELECT subtree.id FROM subtree;

-- name: MoveFolder :execrows
UPDATE folders
SET parent_id = ?
WHERE id = ?;
