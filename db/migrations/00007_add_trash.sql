-- +goose Up
ALTER TABLE folders ADD COLUMN deleted_at INTEGER
    CHECK (deleted_at IS NULL OR deleted_at >= created_at);

ALTER TABLE folders ADD COLUMN former_parent_path TEXT NOT NULL DEFAULT '[]'
    CHECK (
        json_valid(former_parent_path) AND json_type(former_parent_path) = 'array'
        AND (json_array_length(former_parent_path) = 0 OR (deleted_at IS NOT NULL AND parent_id IS NULL))
    );

CREATE INDEX folders_deleted_at ON folders (deleted_at) WHERE deleted_at IS NOT NULL;

ALTER TABLE snippets ADD COLUMN deleted_at INTEGER
    CHECK (deleted_at IS NULL OR deleted_at >= created_at);

ALTER TABLE snippets ADD COLUMN former_folder_path TEXT NOT NULL DEFAULT '[]'
    CHECK (
        json_valid(former_folder_path) AND json_type(former_folder_path) = 'array'
        AND (json_array_length(former_folder_path) = 0 OR (deleted_at IS NOT NULL AND folder_id IS NULL))
    );

CREATE INDEX snippets_deleted_at ON snippets (deleted_at) WHERE deleted_at IS NOT NULL;

CREATE VIEW live_folders AS
WITH RECURSIVE live (id) AS (
    SELECT folders.id FROM folders WHERE folders.parent_id IS NULL AND folders.deleted_at IS NULL
    UNION ALL
    SELECT folders.id FROM folders JOIN live ON folders.parent_id = live.id WHERE folders.deleted_at IS NULL
)
SELECT folders.id, folders.parent_id, folders.name, folders.default_language, folders.created_at, folders.updated_at,
       folders.deleted_at, folders.former_parent_path
FROM folders
JOIN live ON live.id = folders.id;

CREATE VIEW live_snippets AS
SELECT snippets.id, snippets.folder_id, snippets.title, snippets.description, snippets.created_at, snippets.updated_at,
       snippets.deleted_at, snippets.former_folder_path
FROM snippets
WHERE snippets.deleted_at IS NULL
    AND (snippets.folder_id IS NULL OR snippets.folder_id IN (SELECT live_folders.id FROM live_folders));
