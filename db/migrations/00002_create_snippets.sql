-- +goose Up
CREATE TABLE snippets (
    id          TEXT    NOT NULL PRIMARY KEY CHECK (length(id) = 36),
    folder_id   TEXT    REFERENCES folders (id) ON DELETE CASCADE CHECK (length(folder_id) = 36),
    title       TEXT    NOT NULL CHECK (title <> '' AND title = trim(title, ' ' || char(9, 10, 11, 12, 13)) AND length(title) <= 200),
    description TEXT    NOT NULL DEFAULT '' CHECK (length(description) <= 2000),
    created_at  INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000000000 AS INTEGER)) CHECK (created_at > 0),
    updated_at  INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000000000 AS INTEGER)),
    CHECK (updated_at >= created_at)
) STRICT;

CREATE INDEX snippets_folder_id ON snippets (folder_id);
