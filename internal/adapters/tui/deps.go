package tui

import "log/slog"

type Deps struct {
	Lister                FolderSnippetsLister
	TreeLister            FolderTreeLister
	TagLister             TagLister
	TagSnippetsLister     TagSnippetsLister
	Copier                SnippetCopier
	Creator               SnippetCreator
	Capturer              SnippetCapturer
	Updater               SnippetUpdater
	Searcher              SnippetSearcher
	FolderCreator         FolderCreator
	FolderRenamer         FolderRenamer
	FolderDeletePreviewer FolderDeletePreviewer
	FolderDeleter         FolderDeleter
	SortOrderSaver        SortOrderSaver
	CollapsedFoldersSaver CollapsedFoldersSaver
	Settings              Settings
	Logger                *slog.Logger
}
