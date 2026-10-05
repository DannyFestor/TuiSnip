package tui

import "log/slog"

type Deps struct {
	Lister                FolderSnippetsLister
	TreeLister            FolderTreeLister
	Copier                SnippetCopier
	Creator               SnippetCreator
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
