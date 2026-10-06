package binding

const (
	labelQuit           = "quit"
	labelHelp           = "help"
	labelSearch         = "search"
	labelZoom           = "zoom"
	labelNew            = "new"
	labelCapture        = "Capture"
	labelNewFolder      = "new Folder"
	labelNewTag         = "new Tag"
	labelRename         = "rename"
	labelDelete         = "delete"
	labelCollapse       = "collapse"
	labelDefaultLang    = "Default Language"
	labelLanguage       = "Language"
	labelAllLanguages   = "all Languages"
	labelPick           = "pick"
	labelCopy           = "Copy"
	labelEdit           = "edit"
	labelDuplicate      = "duplicate"
	labelSort           = "sort"
	labelWrap           = "wrap"
	labelExternalEditor = "external editor"
	labelSave           = "save"
	labelCancel         = "cancel"
	labelField          = "field"
	labelLeave          = "leave"
	labelIndent         = "indent"
	labelDedent         = "dedent"
	labelMove           = "move"
	labelReveal         = "reveal"
	labelOpen           = "open"
	labelClose          = "close"
	labelYes            = "yes"
	labelNo             = "no"
	labelNextPane       = "next Pane"
	labelPreviousPane   = "previous Pane"
	labelPaneRight      = "Pane to the right"
	labelPaneLeft       = "Pane to the left"
	labelFolders        = "Folders"
	labelTags           = "Tags"
	labelSnippetList    = "Snippet list"
	labelSnippetPane    = "Snippet pane"
	labelBack           = "back"
	labelDown           = "down"
	labelUp             = "up"
	labelFirstRow       = "first row"
	labelLastRow        = "last row"
	labelPageDown       = "page down"
	labelPageUp         = "page up"
	unlabelled          = ""
)

type Row struct {
	Scope Scope
	Name  string
	Label string
}

func Rows() []Row {
	return append(append(globalRows(), paneRows()...), overlayRows()...)
}

func paneRows() []Row {
	return []Row{
		{Scope: ScopeFolders, Name: NewFolder, Label: labelNewFolder},
		{Scope: ScopeFolders, Name: Rename, Label: labelRename},
		{Scope: ScopeFolders, Name: Delete, Label: labelDelete},
		{Scope: ScopeFolders, Name: Move, Label: labelMove},
		{Scope: ScopeFolders, Name: Collapse, Label: labelCollapse},
		{Scope: ScopeFolders, Name: Language, Label: labelDefaultLang},
		{Scope: ScopeTags, Name: NewTag, Label: labelNewTag},
		{Scope: ScopeTags, Name: Rename, Label: labelRename},
		{Scope: ScopeTags, Name: Delete, Label: labelDelete},
		{Scope: ScopeSnippetList, Name: Copy, Label: labelCopy},
		{Scope: ScopeSnippetList, Name: Edit, Label: labelEdit},
		{Scope: ScopeSnippetList, Name: OpenInEditor, Label: labelExternalEditor},
		{Scope: ScopeSnippetList, Name: Move, Label: labelMove},
		{Scope: ScopeSnippetList, Name: Duplicate, Label: labelDuplicate},
		{Scope: ScopeSnippetList, Name: Delete, Label: labelDelete},
		{Scope: ScopeSnippetList, Name: CycleSort, Label: labelSort},
		{Scope: ScopeSnippetPane, Name: Copy, Label: labelCopy},
		{Scope: ScopeSnippetPane, Name: Edit, Label: labelEdit},
		{Scope: ScopeSnippetPane, Name: OpenInEditor, Label: labelExternalEditor},
		{Scope: ScopeSnippetPane, Name: Wrap, Label: labelWrap},
	}
}

func overlayRows() []Row {
	return []Row{
		{Scope: ScopeEditor, Name: Save, Label: labelSave},
		{Scope: ScopeEditor, Name: Cancel, Label: labelCancel},
		{Scope: ScopeEditor, Name: NextField, Label: labelField},
		{Scope: ScopeEditor, Name: PrevField, Label: unlabelled},
		{Scope: ScopeEditor, Name: OpenField, Label: unlabelled},
		{Scope: ScopeEditor, Name: PickLanguage, Label: labelLanguage},
		{Scope: ScopeEditor, Name: OpenInEditor, Label: labelExternalEditor},
		{Scope: ScopeContent, Name: Save, Label: labelSave},
		{Scope: ScopeContent, Name: Leave, Label: labelLeave},
		{Scope: ScopeContent, Name: Indent, Label: labelIndent},
		{Scope: ScopeContent, Name: Dedent, Label: labelDedent},
		{Scope: ScopeContent, Name: PickLanguage, Label: labelLanguage},
		{Scope: ScopeContent, Name: OpenInEditor, Label: labelExternalEditor},
		{Scope: ScopeSearch, Name: Down, Label: labelMove},
		{Scope: ScopeSearch, Name: Up, Label: unlabelled},
		{Scope: ScopeSearch, Name: Accept, Label: labelReveal},
		{Scope: ScopeSearch, Name: Copy, Label: labelCopy},
		{Scope: ScopeSearch, Name: Cancel, Label: labelClose},
		{Scope: ScopePicker, Name: Down, Label: labelMove},
		{Scope: ScopePicker, Name: Up, Label: unlabelled},
		{Scope: ScopePicker, Name: Accept, Label: labelPick},
		{Scope: ScopePicker, Name: Cancel, Label: labelClose},
		{Scope: ScopePicker, Name: ShowAllLanguages, Label: labelAllLanguages},
		{Scope: ScopeNameInput, Name: Accept, Label: labelSave},
		{Scope: ScopeNameInput, Name: Cancel, Label: labelCancel},
		{Scope: ScopeConfirm, Name: Yes, Label: labelYes},
		{Scope: ScopeConfirm, Name: No, Label: labelNo},
	}
}

func globalRows() []Row {
	return []Row{
		{Scope: ScopeGlobal, Name: Quit, Label: labelQuit},
		{Scope: ScopeGlobal, Name: Help, Label: labelHelp},
		{Scope: ScopeGlobal, Name: Search, Label: labelSearch},
		{Scope: ScopeGlobal, Name: Zoom, Label: labelZoom},
		{Scope: ScopeGlobal, Name: NewSnippet, Label: labelNew},
		{Scope: ScopeGlobal, Name: Capture, Label: labelCapture},
		{Scope: ScopeGlobal, Name: FocusNext, Label: labelNextPane},
		{Scope: ScopeGlobal, Name: FocusPrev, Label: labelPreviousPane},
		{Scope: ScopeGlobal, Name: FocusRight, Label: labelPaneRight},
		{Scope: ScopeGlobal, Name: FocusLeft, Label: labelPaneLeft},
		{Scope: ScopeGlobal, Name: FocusFolders, Label: labelFolders},
		{Scope: ScopeGlobal, Name: FocusTags, Label: labelTags},
		{Scope: ScopeGlobal, Name: FocusList, Label: labelSnippetList},
		{Scope: ScopeGlobal, Name: FocusSnippet, Label: labelSnippetPane},
		{Scope: ScopeGlobal, Name: Open, Label: labelOpen},
		{Scope: ScopeGlobal, Name: Back, Label: labelBack},
		{Scope: ScopeGlobal, Name: Down, Label: labelDown},
		{Scope: ScopeGlobal, Name: Up, Label: labelUp},
		{Scope: ScopeGlobal, Name: Top, Label: labelFirstRow},
		{Scope: ScopeGlobal, Name: Bottom, Label: labelLastRow},
		{Scope: ScopeGlobal, Name: PageDown, Label: labelPageDown},
		{Scope: ScopeGlobal, Name: PageUp, Label: labelPageUp},
	}
}
