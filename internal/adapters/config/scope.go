package config

import "slices"

//go:generate go-enum --marshal --names

// ENUM(global, folders, tags, snippet_list, trash, snippet_pane, editor, content, search, picker, name_input, confirm).
type Scope string

func (s Scope) isTextEntry() bool {
	return slices.Contains([]Scope{ScopeEditor, ScopeContent, ScopeSearch, ScopePicker, ScopeNameInput}, s)
}

func (s Scope) isPane() bool {
	return slices.Contains([]Scope{ScopeFolders, ScopeTags, ScopeSnippetList, ScopeTrash, ScopeSnippetPane}, s)
}
