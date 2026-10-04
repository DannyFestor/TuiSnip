package tui

import "log/slog"

type Deps struct {
	Lister   FolderSnippetsLister
	Tree     FolderTreeLister
	Copier   SnippetCopier
	Creator  SnippetCreator
	Searcher SnippetSearcher
	Settings Settings
	Logger   *slog.Logger
}
