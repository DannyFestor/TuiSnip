package binding

const (
	labelEmptyListNewSnippet = "new Snippet"
	labelEmptyListSearch     = "Search"
)

type rowRef struct {
	scope Scope
	name  string
}

type labelledRef struct {
	row   rowRef
	label string
}

func statusHintList(scope Scope) []labelledRef {
	if scope.isPane() {
		return paneHintList(scope)
	}

	return nonPaneHintList(scope)
}

func paneHintList(scope Scope) []labelledRef {
	return append(paneOwnHintList(scope), withRowLabels(refsIn(ScopeGlobal, Search, Help)...)...)
}

func paneOwnHintList(scope Scope) []labelledRef {
	switch scope {
	case ScopeFolders:
		open := labelledRef{row: rowRef{scope: ScopeGlobal, name: Open}, label: labelOpen}

		return append([]labelledRef{open}, withRowLabels(refsIn(scope, NewFolder, Rename, Delete)...)...)
	case ScopeSnippetList:
		return withRowLabels(
			rowRef{scope: scope, name: Copy},
			rowRef{scope: ScopeGlobal, name: NewSnippet},
			rowRef{scope: scope, name: CycleSort},
		)
	case ScopeSnippetPane:
		return withRowLabels(rowRef{scope: scope, name: Copy})
	case ScopeTags, ScopeGlobal, ScopeEditor, ScopeContent, ScopeSearch, ScopePicker, ScopeNameInput, ScopeConfirm:
	}

	return nil
}

func nonPaneHintList(scope Scope) []labelledRef {
	switch scope {
	case ScopeEditor:
		return withRowLabels(refsIn(scope, Save, Cancel, NextField)...)
	case ScopeContent:
		return withRowLabels(refsIn(scope, Save, Leave)...)
	case ScopeSearch:
		return withRowLabels(refsIn(scope, Down, Accept, Copy, Cancel)...)
	case ScopeNameInput:
		return withRowLabels(refsIn(scope, Accept, Cancel)...)
	case ScopeConfirm:
		return withRowLabels(refsIn(scope, Yes, No)...)
	case ScopeGlobal, ScopeFolders, ScopeTags, ScopeSnippetList, ScopeSnippetPane, ScopePicker:
	}

	return nil
}

func emptyListHintList() []labelledRef {
	return []labelledRef{
		{row: rowRef{scope: ScopeGlobal, name: NewSnippet}, label: labelEmptyListNewSnippet},
		{row: rowRef{scope: ScopeGlobal, name: Capture}, label: labelCapture},
		{row: rowRef{scope: ScopeGlobal, name: Search}, label: labelEmptyListSearch},
		{row: rowRef{scope: ScopeFolders, name: NewFolder}, label: labelNewFolder},
		{row: rowRef{scope: ScopeGlobal, name: Help}, label: labelHelp},
	}
}

func refsIn(scope Scope, names ...string) []rowRef {
	refs := make([]rowRef, 0, len(names))
	for _, name := range names {
		refs = append(refs, rowRef{scope: scope, name: name})
	}

	return refs
}

func withRowLabels(refs ...rowRef) []labelledRef {
	labelled := make([]labelledRef, 0, len(refs))
	for _, ref := range refs {
		labelled = append(labelled, labelledRef{row: ref, label: labelOf(ref)})
	}

	return labelled
}

func labelOf(ref rowRef) string {
	for _, row := range Rows() {
		if row.Scope == ref.scope && row.Name == ref.name {
			return row.Label
		}
	}

	return unlabelled
}
