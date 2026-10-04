package binding

const (
	labelHelp           = "help"
	labelSearch         = "search"
	labelNew            = "new"
	labelCapture        = "Capture"
	labelNewFolder      = "new Folder"
	labelRename         = "rename"
	labelCopy           = "Copy"
	labelExternalEditor = "external editor"
	labelSave           = "save"
	labelCancel         = "cancel"
	labelField          = "field"
	labelLeave          = "leave"
	labelMove           = "move"
	labelReveal         = "reveal"
	labelOpen           = "open"
	labelClose          = "close"
	labelYes            = "yes"
	labelNo             = "no"
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
		Row{Scope: ScopeSnippetList, Name: Copy, Label: labelCopy},
		Row{Scope: ScopeSnippetPane, Name: Copy, Label: labelCopy},
		Row{Scope: ScopeEditor, Name: Save, Label: labelSave},
		Row{Scope: ScopeEditor, Name: Cancel, Label: labelCancel},
		Row{Scope: ScopeEditor, Name: NextField, Label: labelField},
		Row{Scope: ScopeEditor, Name: PrevField, Label: unlabelled},
		Row{Scope: ScopeEditor, Name: OpenField, Label: unlabelled},
		Row{Scope: ScopeEditor, Name: OpenInEditor, Label: labelExternalEditor},
		Row{Scope: ScopeContent, Name: Save, Label: labelSave},
		Row{Scope: ScopeContent, Name: Leave, Label: labelLeave},
		Row{Scope: ScopeContent, Name: OpenInEditor, Label: labelExternalEditor},
		Row{Scope: ScopeSearch, Name: Down, Label: labelMove},
		Row{Scope: ScopeSearch, Name: Up, Label: unlabelled},
		Row{Scope: ScopeSearch, Name: Accept, Label: labelReveal},
		Row{Scope: ScopeSearch, Name: Copy, Label: labelCopy},
		Row{Scope: ScopeSearch, Name: Cancel, Label: labelClose},
		Row{Scope: ScopeNameInput, Name: Accept, Label: labelSave},
		Row{Scope: ScopeNameInput, Name: Cancel, Label: labelCancel},
		Row{Scope: ScopeConfirm, Name: Yes, Label: labelYes},
		Row{Scope: ScopeConfirm, Name: No, Label: labelNo},
	)
}

func globalRows() []Row {
	unlabelledNames := []string{
		Quit, FocusNext, FocusPrev, FocusRight, FocusLeft, FocusFolders, FocusTags, FocusList, FocusSnippet,
		Open, Back, Down, Up, Top, Bottom, PageDown, PageUp,
	}
	labelled := []Row{
		{Scope: ScopeGlobal, Name: Help, Label: labelHelp},
		{Scope: ScopeGlobal, Name: Search, Label: labelSearch},
		{Scope: ScopeGlobal, Name: NewSnippet, Label: labelNew},
		{Scope: ScopeGlobal, Name: Capture, Label: labelCapture},
	}

	rows := make([]Row, 0, len(labelled)+len(unlabelledNames))
	rows = append(rows, labelled...)

	for _, name := range unlabelledNames {
		rows = append(rows, Row{Scope: ScopeGlobal, Name: name, Label: unlabelled})
	}

	return rows
}
