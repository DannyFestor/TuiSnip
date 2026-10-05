package tui

import "log/slog"

type Deps struct {
	Lister                      FolderSnippetsLister
	TreeLister                  FolderTreeLister
	TagLister                   TagLister
	TagSnippetsLister           TagSnippetsLister
	Copier                      SnippetCopier
	Creator                     SnippetCreator
	Updater                     SnippetUpdater
	Searcher                    SnippetSearcher
	FolderCreator               FolderCreator
	FolderRenamer               FolderRenamer
	FolderDeletePreviewer       FolderDeletePreviewer
	FolderDeleter               FolderDeleter
	FolderDefaultLanguageSetter FolderDefaultLanguageSetter
	SortOrderSaver              SortOrderSaver
	CollapsedFoldersSaver       CollapsedFoldersSaver
	Settings                    Settings
	Logger                      *slog.Logger
}
