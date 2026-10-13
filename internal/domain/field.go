package domain

//go:generate go-enum --marshal

// ENUM(title, description, content, language, fragment_title, folder_name, tag_name).
type Field string
