package tui

import "log/slog"

type Deps struct {
	Lister   FolderSnippetsLister
	Copier   SnippetCopier
	Creator  SnippetCreator
	Searcher SnippetSearcher
	Settings Settings
	Logger   *slog.Logger
}
