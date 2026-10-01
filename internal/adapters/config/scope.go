package config

//go:generate go-enum --marshal --names

// ENUM(global, folders, tags, snippet_list, snippet_pane, editor, content, search, picker, confirm).
type Scope string
