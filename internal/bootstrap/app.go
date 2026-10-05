package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/clipboard"
	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/logging"
	"github.com/DannyFestor/TuiSnip/internal/adapters/memsearch"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/adapters/state"
	"github.com/DannyFestor/TuiSnip/internal/adapters/system"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/xdg"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
)

type App struct {
	Create              *snippet.Create
	Capture             *snippet.Capture
	Update              *snippet.Update
	Copy                *snippet.Copy
	Query               *search.Query
	SnippetsInFolder    *browse.SnippetsInFolder
	FolderTree          *browse.FolderTree
	TagList             *browse.TagList
	SnippetsWithTag     *browse.SnippetsWithTag
	CreateFolder        *folder.Create
	RenameFolder        *folder.Rename
	PreviewDeleteFolder *folder.PreviewDelete
	DeleteFolder        *folder.Delete
	CreateTag           *tag.Create
	RenameTag           *tag.Rename
	PreviewDeleteTag    *tag.PreviewDelete
	DeleteTag           *tag.Delete
	SnippetRepository   *sqlite.SnippetRepository
	FolderRepository    *sqlite.FolderRepository
	TagRepository       *sqlite.TagRepository
	model               tui.Model
	database            *sqlite.Database
	log                 *logging.Log
}

func New(ctx context.Context, options Options) (*App, error) {
	log, err := logging.Open(options.Paths.LogFile)
	if err != nil {
		return nil, fmt.Errorf("bootstrap.New: %w", err)
	}

	app, err := openApp(ctx, options, log)
	if err != nil {
		log.Logger().ErrorContext(ctx, "start-up failed", slog.Any(keyError, err))

		return nil, fmt.Errorf("bootstrap.New: %w", errors.Join(err, log.Close()))
	}

	return app, nil
}

func (a *App) TUI() tui.Model {
	return a.model
}

func (a *App) Run(ctx context.Context) error {
	_, err := tea.NewProgram(a.model, tea.WithContext(ctx)).Run()
	a.log.Logger().InfoContext(ctx, "tuisnip stopped")

	if err != nil {
		return fmt.Errorf("bootstrap.App.Run: %w", err)
	}

	return nil
}

func (a *App) Close() error {
	err := errors.Join(a.database.Close(), a.log.Close())
	if err != nil {
		return fmt.Errorf("bootstrap.App.Close: %w", err)
	}

	return nil
}

func openApp(ctx context.Context, options Options, log *logging.Log) (*App, error) {
	logger := log.Logger()
	logStarted(ctx, logger, options.Paths)

	cfg, err := config.Load(ctx, config.Options{Path: options.Paths.ConfigFile, Logger: logger})
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	database, err := sqlite.Open(ctx, sqlite.Options{
		Path:   options.Paths.DatabaseFile,
		Logger: logger,
		Clock:  system.NewClock(),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	remembered := state.Load(ctx, state.Options{Path: options.Paths.StateFile, Logger: logger})

	app, err := wire(ctx, cfg, options, openResources{database: database, log: log, remembered: remembered})
	if err != nil {
		return nil, errors.Join(err, database.Close())
	}

	return app, nil
}

type openResources struct {
	database   *sqlite.Database
	log        *logging.Log
	remembered *state.File
}

func wire(ctx context.Context, cfg config.Config, options Options, opened openResources) (*App, error) {
	logger := opened.log.Logger()

	app, err := newActions(cfg, options, openRepositories(opened.database, logger), logger)
	if err != nil {
		return nil, err
	}

	app.database = opened.database
	app.log = opened.log

	app.model, err = newModel(ctx, cfg, tui.Deps{
		Lister:                app.SnippetsInFolder,
		TreeLister:            app.FolderTree,
		TagLister:             app.TagList,
		TagSnippetsLister:     app.SnippetsWithTag,
		Copier:                app.Copy,
		Creator:               app.Create,
		Capturer:              app.Capture,
		Updater:               app.Update,
		Searcher:              app.Query,
		FolderCreator:         app.CreateFolder,
		FolderRenamer:         app.RenameFolder,
		FolderDeletePreviewer: app.PreviewDeleteFolder,
		FolderDeleter:         app.DeleteFolder,
		TagCreator:            app.CreateTag,
		TagRenamer:            app.RenameTag,
		TagDeletePreviewer:    app.PreviewDeleteTag,
		TagDeleter:            app.DeleteTag,
		SortOrderSaver:        opened.remembered,
		CollapsedFoldersSaver: opened.remembered,
		Settings:              SettingsFrom(cfg, time.Local, rememberedIn(opened.remembered)),
		Logger:                logger,
	})
	if err != nil {
		return nil, fmt.Errorf("build TUI: %w", err)
	}

	return app, nil
}

func rememberedIn(file *state.File) mainscreen.Remembered {
	return mainscreen.Remembered{SortOrder: file.SortOrder(), CollapsedFolders: file.CollapsedFolders()}
}

func newModel(ctx context.Context, cfg config.Config, deps tui.Deps) (tui.Model, error) {
	build := tui.New
	if cfg.Copy.QuitAfter {
		build = tui.NewQuittingAfterCopy
	}

	return build(ctx, deps)
}

type repositories struct {
	snippets *sqlite.SnippetRepository
	folders  *sqlite.FolderRepository
	tags     *sqlite.TagRepository
}

func openRepositories(database *sqlite.Database, logger *slog.Logger) repositories {
	return repositories{
		snippets: sqlite.NewSnippetRepository(database, logger),
		folders:  sqlite.NewFolderRepository(database, logger),
		tags:     sqlite.NewTagRepository(database, logger),
	}
}

func newActions(cfg config.Config, options Options, repos repositories, logger *slog.Logger) (*App, error) {
	clipboardOptions := clipboardOptionsFrom(options, logger)

	copyAction, err := newCopy(cfg.Copy, clipboardOptions, repos.snippets)
	if err != nil {
		return nil, err
	}

	create, createErr := snippet.NewCreate(repos.snippets, system.NewIDs(), system.NewClock())
	capture, captureErr := snippet.NewCapture(clipboard.NewNative(clipboardOptions))
	update, updateErr := snippet.NewUpdate(repos.snippets, system.NewClock())
	query, queryErr := search.NewQuery(memsearch.NewIndex(repos.snippets))
	browsing, browseErr := newBrowseActions(repos)
	createFolder, createFolderErr := folder.NewCreate(repos.folders, system.NewIDs(), system.NewClock())
	renameFolder, renameFolderErr := folder.NewRename(repos.folders, system.NewClock())
	previewDeleteFolder, previewDeleteFolderErr := folder.NewPreviewDelete(repos.folders)
	deleteFolder, deleteFolderErr := folder.NewDelete(repos.folders)
	tags, tagsErr := newTagActions(repos.tags)

	err = errors.Join(createErr, captureErr, updateErr, queryErr, browseErr, tagsErr,
		createFolderErr, renameFolderErr, previewDeleteFolderErr, deleteFolderErr)
	if err != nil {
		return nil, fmt.Errorf("build Actions: %w", err)
	}

	return &App{
		Create:              create,
		Capture:             capture,
		Update:              update,
		Copy:                copyAction,
		Query:               query,
		SnippetsInFolder:    browsing.snippetsInFolder,
		FolderTree:          browsing.folderTree,
		TagList:             browsing.tagList,
		SnippetsWithTag:     browsing.snippetsWithTag,
		CreateFolder:        createFolder,
		RenameFolder:        renameFolder,
		PreviewDeleteFolder: previewDeleteFolder,
		DeleteFolder:        deleteFolder,
		CreateTag:           tags.create,
		RenameTag:           tags.rename,
		PreviewDeleteTag:    tags.previewDelete,
		DeleteTag:           tags.delete,
		SnippetRepository:   repos.snippets,
		FolderRepository:    repos.folders,
		TagRepository:       repos.tags,
		model:               tui.Model{},
		database:            nil,
		log:                 nil,
	}, nil
}

func logStarted(ctx context.Context, logger *slog.Logger, paths xdg.Paths) {
	logger.InfoContext(ctx, "tuisnip started",
		slog.String(keyConfigPath, paths.ConfigFile),
		slog.String(keyDatabasePath, paths.DatabaseFile),
		slog.String(keyLogPath, paths.LogFile),
	)
}
