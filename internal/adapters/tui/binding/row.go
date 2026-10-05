package binding

const (
	labelQuit           = "quit"
	labelHelp           = "help"
	labelSearch         = "search"
	labelZoom           = "zoom"
	labelNew            = "new"
	labelCapture        = "Capture"
	labelNewFolder      = "new Folder"
	labelRename         = "rename"
	labelDelete         = "delete"
	labelCollapse       = "collapse"
	labelDefaultLang    = "Default Language"
	labelLanguage       = "Language"
	labelAllLanguages   = "all Languages"
	labelPick           = "pick"
	labelCopy           = "Copy"
	labelEdit           = "edit"
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
	rows := globalRows()

	return append(rows,
		Row{Scope: ScopeFolders, Name: NewFolder, Label: labelNewFolder},
		Row{Scope: ScopeFolders, Name: Rename, Label: labelRename},
		Row{Scope: ScopeFolders, Name: Delete, Label: labelDelete},
		Row{Scope: ScopeFolders, Name: Collapse, Label: labelCollapse},
		Row{Scope: ScopeFolders, Name: Language, Label: labelDefaultLang},
		Row{Scope: ScopeSnippetList, Name: Copy, Label: labelCopy},
		Row{Scope: ScopeSnippetList, Name: Edit, Label: labelEdit},
		Row{Scope: ScopeSnippetList, Name: CycleSort, Label: labelSort},
		Row{Scope: ScopeSnippetPane, Name: Copy, Label: labelCopy},
		Row{Scope: ScopeSnippetPane, Name: Edit, Label: labelEdit},
		Row{Scope: ScopeSnippetPane, Name: Wrap, Label: labelWrap},
		Row{Scope: ScopeEditor, Name: Save, Label: labelSave},
		Row{Scope: ScopeEditor, Name: Cancel, Label: labelCancel},
		Row{Scope: ScopeEditor, Name: NextField, Label: labelField},
		Row{Scope: ScopeEditor, Name: PrevField, Label: unlabelled},
		Row{Scope: ScopeEditor, Name: OpenField, Label: unlabelled},
		Row{Scope: ScopeEditor, Name: PickLanguage, Label: labelLanguage},
		Row{Scope: ScopeEditor, Name: OpenInEditor, Label: labelExternalEditor},
		Row{Scope: ScopeContent, Name: Save, Label: labelSave},
		Row{Scope: ScopeContent, Name: Leave, Label: labelLeave},
		Row{Scope: ScopeContent, Name: Indent, Label: labelIndent},
		Row{Scope: ScopeContent, Name: Dedent, Label: labelDedent},
		Row{Scope: ScopeContent, Name: PickLanguage, Label: labelLanguage},
		Row{Scope: ScopeContent, Name: OpenInEditor, Label: labelExternalEditor},
		Row{Scope: ScopeSearch, Name: Down, Label: labelMove},
		Row{Scope: ScopeSearch, Name: Up, Label: unlabelled},
		Row{Scope: ScopeSearch, Name: Accept, Label: labelReveal},
		Row{Scope: ScopeSearch, Name: Copy, Label: labelCopy},
		Row{Scope: ScopeSearch, Name: Cancel, Label: labelClose},
		Row{Scope: ScopePicker, Name: Down, Label: labelMove},
		Row{Scope: ScopePicker, Name: Up, Label: unlabelled},
		Row{Scope: ScopePicker, Name: Accept, Label: labelPick},
		Row{Scope: ScopePicker, Name: Cancel, Label: labelClose},
		Row{Scope: ScopePicker, Name: ShowAllLanguages, Label: labelAllLanguages},
		Row{Scope: ScopeNameInput, Name: Accept, Label: labelSave},
		Row{Scope: ScopeNameInput, Name: Cancel, Label: labelCancel},
		Row{Scope: ScopeConfirm, Name: Yes, Label: labelYes},
		Row{Scope: ScopeConfirm, Name: No, Label: labelNo},
	)
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
