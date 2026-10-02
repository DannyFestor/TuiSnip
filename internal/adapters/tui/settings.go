package tui

import "time"

type Settings struct {
	Global      GlobalKeyMap
	Folders     FoldersKeyMap
	SnippetList SnippetListKeyMap
	SnippetPane SnippetPaneKeyMap
	Location    *time.Location
}
