-- +goose Up
ALTER TABLE fragments ADD COLUMN title TEXT NOT NULL DEFAULT ''
    CHECK (length(title) <= 200 AND title = trim(title, ' ' || char(9, 10, 11, 12, 13)));
