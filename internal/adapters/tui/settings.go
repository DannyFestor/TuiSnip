package tui

import "time"

type Settings struct {
	Global      GlobalKeyMap
	Folders     FoldersKeyMap
	SnippetList SnippetListKeyMap
	SnippetPane SnippetPaneKeyMap
	Editor      EditorKeyMap
	Content     ContentKeyMap
	Search      SearchKeyMap
	Confirm     ConfirmKeyMap
	Location    *time.Location
}
