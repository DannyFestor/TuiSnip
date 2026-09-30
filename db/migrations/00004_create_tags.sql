-- +goose Up
CREATE TABLE tags (
    id         TEXT    NOT NULL PRIMARY KEY CHECK (length(id) = 36),
    name       TEXT    NOT NULL CHECK (name <> '' AND name = trim(name, ' ' || char(9, 10, 11, 12, 13))),
    name_key   TEXT    NOT NULL UNIQUE CHECK (name_key <> ''),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000000000 AS INTEGER)) CHECK (created_at > 0),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000000000 AS INTEGER)),
    CHECK (updated_at >= created_at)
) STRICT;
