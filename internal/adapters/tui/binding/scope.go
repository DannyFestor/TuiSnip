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
	ScopeNameInput   Scope = "name_input"
	ScopeConfirm     Scope = "confirm"
)

func (s Scope) isPane() bool {
	return s == ScopeFolders || s == ScopeTags || s == ScopeSnippetList || s == ScopeSnippetPane
}
