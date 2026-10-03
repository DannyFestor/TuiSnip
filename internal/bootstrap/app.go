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
	"github.com/DannyFestor/TuiSnip/internal/adapters/system"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/adapters/xdg"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
)

type App struct {
	Create           *snippet.Create
	Copy             *snippet.Copy
	Query            *search.Query
	SnippetsInFolder *browse.SnippetsInFolder
	model            tui.Model
	database         *sqlite.Database
	log              *logging.Log
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

	app, err := wire(ctx, cfg, options, openResources{database: database, log: log})
	if err != nil {
		return nil, errors.Join(err, database.Close())
	}

	return app, nil
}

type openResources struct {
	database *sqlite.Database
	log      *logging.Log
}

func wire(ctx context.Context, cfg config.Config, options Options, opened openResources) (*App, error) {
	logger := opened.log.Logger()

	app, err := newActions(cfg, options, sqlite.NewSnippetRepository(opened.database, logger), logger)
	if err != nil {
		return nil, err
	}

	app.database = opened.database
	app.log = opened.log

	app.model, err = newModel(ctx, cfg, tui.Deps{
		Lister:   app.SnippetsInFolder,
		Copier:   app.Copy,
		Creator:  app.Create,
		Searcher: app.Query,
		Settings: SettingsFrom(cfg, time.Local),
		Logger:   logger,
	})
	if err != nil {
		return nil, fmt.Errorf("build TUI: %w", err)
	}

	return app, nil
}

func newModel(ctx context.Context, cfg config.Config, deps tui.Deps) (tui.Model, error) {
	build := tui.New
	if cfg.Copy.QuitAfter {
		build = tui.NewQuittingAfterCopy
	}

	return build(ctx, deps)
}

func newActions(cfg config.Config, options Options, repo *sqlite.SnippetRepository, logger *slog.Logger) (*App, error) {
	copyAction, err := newCopy(cfg.Copy, options, repo, logger)
	if err != nil {
		return nil, err
	}

	create, createErr := snippet.NewCreate(repo, system.NewIDs(), system.NewClock())
	query, queryErr := search.NewQuery(memsearch.NewIndex(repo))
	snippetsInFolder, listErr := browse.NewSnippetsInFolder(repo)

	err = errors.Join(createErr, queryErr, listErr)
	if err != nil {
		return nil, fmt.Errorf("build Actions: %w", err)
	}

	return &App{
		Create:           create,
		Copy:             copyAction,
		Query:            query,
		SnippetsInFolder: snippetsInFolder,
		model:            tui.Model{},
		database:         nil,
		log:              nil,
	}, nil
}

func logStarted(ctx context.Context, logger *slog.Logger, paths xdg.Paths) {
	logger.InfoContext(ctx, "tuisnip started",
		slog.String(keyConfigPath, paths.ConfigFile),
		slog.String(keyDatabasePath, paths.DatabaseFile),
		slog.String(keyLogPath, paths.LogFile),
	)
}
