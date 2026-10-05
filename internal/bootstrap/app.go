package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	tea "charm.land/bubbletea/v2"

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
)

type App struct {
	Create                   *snippet.Create
	Update                   *snippet.Update
	Copy                     *snippet.Copy
	Query                    *search.Query
	SnippetsInFolder         *browse.SnippetsInFolder
	FolderTree               *browse.FolderTree
	TagList                  *browse.TagList
	SnippetsWithTag          *browse.SnippetsWithTag
	CreateFolder             *folder.Create
	RenameFolder             *folder.Rename
	PreviewDeleteFolder      *folder.PreviewDelete
	DeleteFolder             *folder.Delete
	SetFolderDefaultLanguage *folder.SetDefaultLanguage
	SnippetRepository        *sqlite.SnippetRepository
	FolderRepository         *sqlite.FolderRepository
	TagRepository            *sqlite.TagRepository
	model                    tui.Model
	database                 *sqlite.Database
	log                      *logging.Log
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
		Lister:                      app.SnippetsInFolder,
		TreeLister:                  app.FolderTree,
		TagLister:                   app.TagList,
		TagSnippetsLister:           app.SnippetsWithTag,
		Copier:                      app.Copy,
		Creator:                     app.Create,
		Updater:                     app.Update,
		Searcher:                    app.Query,
		FolderCreator:               app.CreateFolder,
		FolderRenamer:               app.RenameFolder,
		FolderDeletePreviewer:       app.PreviewDeleteFolder,
		FolderDeleter:               app.DeleteFolder,
		FolderDefaultLanguageSetter: app.SetFolderDefaultLanguage,
		SortOrderSaver:              opened.remembered,
		CollapsedFoldersSaver:       opened.remembered,
		Settings:                    SettingsFrom(cfg, time.Local, rememberedIn(opened.remembered)),
		Logger:                      logger,
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
	copyAction, err := newCopy(cfg.Copy, options, repos.snippets, logger)
	if err != nil {
		return nil, err
	}

	create, createErr := snippet.NewCreate(repos.snippets, system.NewIDs(), system.NewClock())
	update, updateErr := snippet.NewUpdate(repos.snippets, system.NewClock())
	query, queryErr := search.NewQuery(memsearch.NewIndex(repos.snippets))
	snippetsInFolder, listErr := browse.NewSnippetsInFolder(repos.snippets)
	folderTree, treeErr := browse.NewFolderTree(repos.folders, repos.snippets)
	tagList, tagListErr := browse.NewTagList(repos.tags, repos.snippets)
	snippetsWithTag, withTagErr := browse.NewSnippetsWithTag(repos.snippets)
	folders, foldersErr := newFolderActions(repos.folders)

	err = errors.Join(createErr, updateErr, queryErr, listErr, treeErr, tagListErr, withTagErr, foldersErr)
	if err != nil {
		return nil, fmt.Errorf("build Actions: %w", err)
	}

	return &App{
		Create:                   create,
		Update:                   update,
		Copy:                     copyAction,
		Query:                    query,
		SnippetsInFolder:         snippetsInFolder,
		FolderTree:               folderTree,
		TagList:                  tagList,
		SnippetsWithTag:          snippetsWithTag,
		CreateFolder:             folders.create,
		RenameFolder:             folders.rename,
		PreviewDeleteFolder:      folders.previewDelete,
		DeleteFolder:             folders.remove,
		SetFolderDefaultLanguage: folders.setDefaultLanguage,
		SnippetRepository:        repos.snippets,
		FolderRepository:         repos.folders,
		TagRepository:            repos.tags,
		model:                    tui.Model{},
		database:                 nil,
		log:                      nil,
	}, nil
}

type folderActions struct {
	create             *folder.Create
	rename             *folder.Rename
	previewDelete      *folder.PreviewDelete
	remove             *folder.Delete
	setDefaultLanguage *folder.SetDefaultLanguage
}

func newFolderActions(folders *sqlite.FolderRepository) (folderActions, error) {
	create, createErr := folder.NewCreate(folders, system.NewIDs(), system.NewClock())
	rename, renameErr := folder.NewRename(folders, system.NewClock())
	previewDelete, previewDeleteErr := folder.NewPreviewDelete(folders)
	remove, removeErr := folder.NewDelete(folders)
	setDefaultLanguage, setDefaultLanguageErr := folder.NewSetDefaultLanguage(folders, system.NewClock())

	return folderActions{
		create:             create,
		rename:             rename,
		previewDelete:      previewDelete,
		remove:             remove,
		setDefaultLanguage: setDefaultLanguage,
	}, errors.Join(createErr, renameErr, previewDeleteErr, removeErr, setDefaultLanguageErr)
}

func logStarted(ctx context.Context, logger *slog.Logger, paths xdg.Paths) {
	logger.InfoContext(ctx, "tuisnip started",
		slog.String(keyConfigPath, paths.ConfigFile),
		slog.String(keyDatabasePath, paths.DatabaseFile),
		slog.String(keyLogPath, paths.LogFile),
	)
}
