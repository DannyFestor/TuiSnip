package tui

import "log/slog"

type Deps struct {
	Lister   FolderSnippetsLister
	Copier   SnippetCopier
	Settings Settings
	Logger   *slog.Logger
}
