-- +goose Up
CREATE TABLE folders (
    id               TEXT    NOT NULL PRIMARY KEY CHECK (length(id) = 36),
    parent_id        TEXT    REFERENCES folders (id) ON DELETE CASCADE CHECK (length(parent_id) = 36),
    name             TEXT    NOT NULL CHECK (name <> '' AND name = trim(name, ' ' || char(9, 10, 11, 12, 13)) AND length(name) <= 200),
    default_language TEXT    NOT NULL CHECK (default_language <> ''),
    created_at       INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000000000 AS INTEGER)) CHECK (created_at > 0),
    updated_at       INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000000000 AS INTEGER)),
    CHECK (updated_at >= created_at),
    CHECK (parent_id <> id)
) STRICT;

CREATE INDEX folders_parent_id ON folders (parent_id);
