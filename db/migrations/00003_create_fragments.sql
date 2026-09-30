-- +goose Up
CREATE TABLE fragments (
    id         TEXT    NOT NULL PRIMARY KEY CHECK (length(id) = 36),
    snippet_id TEXT    NOT NULL REFERENCES snippets (id) ON DELETE CASCADE CHECK (length(snippet_id) = 36),
    position   INTEGER NOT NULL CHECK (position >= 0),
    language   TEXT    NOT NULL CHECK (language <> ''),
    content    TEXT    NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000000000 AS INTEGER)) CHECK (created_at > 0),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000000000 AS INTEGER)),
    CHECK (updated_at >= created_at),
    UNIQUE (snippet_id, position)
) STRICT;
