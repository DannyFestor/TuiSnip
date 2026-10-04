package tui

import "log/slog"

type Deps struct {
	Lister        FolderSnippetsLister
	TreeLister    FolderTreeLister
	Copier        SnippetCopier
	Creator       SnippetCreator
	Searcher      SnippetSearcher
	FolderCreator FolderCreator
	FolderRenamer FolderRenamer
	Settings      Settings
	Logger        *slog.Logger
}
