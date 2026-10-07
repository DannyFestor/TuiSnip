package tui

import "log/slog"

type Deps struct {
	Lister                      FolderSnippetsLister
	TreeLister                  FolderTreeLister
	TagLister                   TagLister
	TagSnippetsLister           TagSnippetsLister
	Copier                      SnippetCopier
	Creator                     SnippetCreator
	Capturer                    SnippetCapturer
	Updater                     SnippetUpdater
	Duplicator                  SnippetDuplicator
	Deleter                     SnippetDeleter
	Mover                       SnippetMover
	Searcher                    SnippetSearcher
	FolderCreator               FolderCreator
	FolderRenamer               FolderRenamer
	FolderDeletePreviewer       FolderDeletePreviewer
	FolderDeleter               FolderDeleter
	FolderDefaultLanguageSetter FolderDefaultLanguageSetter
	FolderMover                 FolderMover
	TagCreator                  TagCreator
	TagRenamer                  TagRenamer
	TagDeletePreviewer          TagDeletePreviewer
	TagDeleter                  TagDeleter
	SortOrderSaver              SortOrderSaver
	CollapsedFoldersSaver       CollapsedFoldersSaver
	ExternalEditor              ExternalEditor
	EditedContentHandler        EditedContentHandler
	Clock                       Clock
	Settings                    Settings
	Logger                      *slog.Logger
}
