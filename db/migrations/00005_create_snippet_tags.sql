-- +goose Up
CREATE TABLE snippet_tags (
    snippet_id TEXT NOT NULL REFERENCES snippets (id) ON DELETE CASCADE CHECK (length(snippet_id) = 36),
    tag_id     TEXT NOT NULL REFERENCES tags (id) ON DELETE CASCADE CHECK (length(tag_id) = 36),
    PRIMARY KEY (snippet_id, tag_id)
) STRICT, WITHOUT ROWID;

CREATE INDEX snippet_tags_tag_id ON snippet_tags (tag_id);
