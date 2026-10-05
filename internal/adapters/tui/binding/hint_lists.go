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
	return append(paneOwnHintList(scope), withRowLabels(refsIn(ScopeGlobal, Zoom, Search, Help)...)...)
}

func paneOwnHintList(scope Scope) []labelledRef {
	open := labelledRef{row: rowRef{scope: ScopeGlobal, name: Open}, label: labelOpen}

	switch scope {
	case ScopeFolders:
		return append([]labelledRef{open}, withRowLabels(refsIn(scope, NewFolder, Rename, Delete)...)...)
	case ScopeTags:
		return append([]labelledRef{open}, withRowLabels(refsIn(scope, NewTag, Rename, Delete)...)...)
	case ScopeSnippetList:
		return withRowLabels(
			rowRef{scope: scope, name: Copy},
			rowRef{scope: scope, name: Edit},
			rowRef{scope: ScopeGlobal, name: NewSnippet},
			rowRef{scope: scope, name: Duplicate},
			rowRef{scope: scope, name: Delete},
			rowRef{scope: scope, name: CycleSort},
		)
	case ScopeSnippetPane:
		return withRowLabels(refsIn(scope, Copy, Edit, Wrap)...)
	case ScopeGlobal, ScopeEditor, ScopeContent, ScopeSearch, ScopePicker, ScopeNameInput, ScopeConfirm:
	}

	return nil
}

func nonPaneHintList(scope Scope) []labelledRef {
	switch scope {
	case ScopeEditor:
		return withRowLabels(refsIn(scope, Save, Cancel, NextField, PickLanguage)...)
	case ScopeContent:
		return withRowLabels(refsIn(scope, Save, Leave, Indent, Dedent)...)
	case ScopeSearch:
		return withRowLabels(refsIn(scope, Down, Accept, Copy, Cancel)...)
	case ScopePicker:
		return withRowLabels(refsIn(scope, Down, Accept, Cancel)...)
	case ScopeNameInput:
		return withRowLabels(refsIn(scope, Accept, Cancel)...)
	case ScopeConfirm:
		return withRowLabels(refsIn(scope, Yes, No)...)
	case ScopeGlobal, ScopeFolders, ScopeTags, ScopeSnippetList, ScopeSnippetPane:
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

func helpOverlayHintList() []labelledRef {
	return []labelledRef{
		{row: rowRef{scope: ScopeGlobal, name: Help}, label: labelClose},
		{row: rowRef{scope: ScopeGlobal, name: Back}, label: labelClose},
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
