package binding

type Scope string

const (
	ScopeGlobal      Scope = "global"
	ScopeFolders     Scope = "folders"
	ScopeTags        Scope = "tags"
	ScopeSnippetList Scope = "snippet_list"
	ScopeSnippetPane Scope = "snippet_pane"
	ScopeEditor      Scope = "editor"
	ScopeContent     Scope = "content"
	ScopeSearch      Scope = "search"
	ScopePicker      Scope = "picker"
	ScopeConfirm     Scope = "confirm"
)
