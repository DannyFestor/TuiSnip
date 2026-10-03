package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/searchpopup"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	operationList   = "list snippets"
	operationCopy   = "copy"
	operationSave   = "save snippet"
	operationSearch = "search"
)

type Model struct {
	//nolint:containedctx // Bubble Tea's Update has no context parameter, so the program context travels with the model.
	ctx           context.Context
	lister        FolderSnippetsLister
	copier        SnippetCopier
	creator       SnippetCreator
	searcher      SnippetSearcher
	logger        *slog.Logger
	forcedQuitKey string
	overlays      outcome.Stack
}

func New(ctx context.Context, deps Deps) (Model, error) {
	err := errors.Join(
		domain.RequireDependency("lister", deps.Lister),
		domain.RequireDependency("copier", deps.Copier),
		domain.RequireDependency("creator", deps.Creator),
		domain.RequireDependency("searcher", deps.Searcher),
		requirePointer("logger", deps.Logger),
		requirePointer("location", deps.Settings.Location),
	)
	if err != nil {
		return Model{}, fmt.Errorf("tui.New: %w", err)
	}

	mainScreen := mainscreen.New(deps.Settings.Keys, look.NewStyles(), deps.Settings.Location)
	overlays, _, _ := outcome.NewStack().Pushed(mainScreen)

	return Model{
		ctx:           ctx,
		lister:        deps.Lister,
		copier:        deps.Copier,
		creator:       deps.Creator,
		searcher:      deps.Searcher,
		logger:        deps.Logger,
		forcedQuitKey: deps.Settings.ForcedQuitKey,
		overlays:      overlays,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.loadSnippets(domain.SnippetID{}))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.pressed(msg)
	case tea.WindowSizeMsg, tea.BackgroundColorMsg, tea.PasteMsg,
		editoverlay.SaveFinished, searchpopup.HitsFound, mainscreen.SnippetsLoaded:
		return m.overlaysUpdated(msg)
	case listFailedMsg:
		return m.failed(operationList, msg.err)
	case copyFinishedMsg:
		return m.copyFinished(msg)
	case searchFailedMsg:
		return m.failed(operationSearch, msg.err)
	}

	return m, nil
}

func (m Model) View() tea.View {
	view := tea.NewView(m.overlays.Render())
	view.AltScreen = true

	return view
}

func (m Model) loadSnippets(selecting domain.SnippetID) tea.Cmd {
	return func() tea.Msg {
		snippets, err := m.lister.Run(m.ctx, browse.SnippetsInFolderInput{FolderID: domain.FolderID{}})
		if err != nil {
			return listFailedMsg{err: err}
		}

		return mainscreen.SnippetsLoaded{Snippets: snippets, Selecting: selecting}
	}
}

func (m Model) pressed(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if msg.String() == m.forcedQuitKey {
		return m.settled(m.overlays.Offered(outcome.QuitAsked{}))
	}

	return m.overlaysUpdated(msg)
}

func (m Model) overlaysUpdated(msg tea.Msg) (Model, tea.Cmd) {
	return m.settled(m.overlays.Update(msg))
}

func (m Model) settled(overlays outcome.Stack, outcomes []outcome.Outcome, cmd tea.Cmd) (Model, tea.Cmd) {
	next := m
	next.overlays = overlays

	return next.concludedAll(outcomes, cmd)
}

func (m Model) concludedAll(outcomes []outcome.Outcome, cmd tea.Cmd) (Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, 1+len(outcomes))
	cmds = append(cmds, cmd)

	for _, reported := range outcomes {
		var concludedCmd tea.Cmd

		m, concludedCmd = m.concluded(reported)
		cmds = append(cmds, concludedCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) concluded(reported outcome.Outcome) (Model, tea.Cmd) {
	switch reported := reported.(type) {
	case outcome.SaveRequested:
		return m, m.createSnippet(reported.Input)
	case outcome.SnippetSaved:
		return m, m.loadSnippets(reported.ID)
	case outcome.SaveFailed:
		return m.failed(operationSave, reported.Err)
	case outcome.NoticeShown:
		return m.shown(reported.Text)
	case outcome.SearchTyped:
		return m, m.querySnippets(reported.Text)
	case outcome.SnippetRevealed:
		return m, m.loadSnippets(reported.ID)
	case outcome.CopyRequested:
		return m, m.copySnippet(reported.ID)
	case outcome.DiscardConfirmed:
	case outcome.QuitAsked, outcome.QuitConfirmed:
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) createSnippet(in snippet.CreateInput) tea.Cmd {
	return func() tea.Msg {
		created, err := m.creator.Run(m.ctx, in)

		return editoverlay.SaveFinished{Snippet: created, Err: err}
	}
}

func (m Model) querySnippets(text string) tea.Cmd {
	query := search.QueryInput{Text: text}

	return func() tea.Msg {
		hits, err := m.searcher.Run(m.ctx, query)
		if err != nil {
			return searchFailedMsg{err: err}
		}

		return searchpopup.HitsFound{Text: query.Text, Hits: hits}
	}
}

func (m Model) copySnippet(id domain.SnippetID) tea.Cmd {
	in := snippet.CopyInput{SnippetID: id}

	return func() tea.Msg {
		result, err := m.copier.Run(m.ctx, in)

		return copyFinishedMsg{result: result, err: err}
	}
}

func (m Model) copyFinished(msg copyFinishedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		return m.failed(operationCopy, msg.err)
	}

	text, err := deliveryText(msg.result.Delivery)
	if err != nil {
		return m.failed(operationCopy, err)
	}

	next, cmd := m.shown(text)
	if msg.result.Delivery == domain.CopyDeliverySentToTerminal {
		return next, tea.Batch(cmd, tea.SetClipboard(msg.result.Content.String()))
	}

	return next, cmd
}

func (m Model) failed(operation string, err error) (Model, tea.Cmd) {
	if errors.Is(err, context.Canceled) {
		m.logger.DebugContext(m.ctx, "operation cancelled", slog.String(keyOperation, operation))

		return m, nil
	}

	m.logger.ErrorContext(m.ctx, "operation failed", slog.String(keyOperation, operation), slog.Any(keyError, err))

	return m.shown(failureText(err))
}

func (m Model) shown(text string) (Model, tea.Cmd) {
	return m.overlaysUpdated(mainscreen.StatusShown{Text: text})
}

func requirePointer[T any](name string, pointer *T) error {
	if pointer == nil {
		return fmt.Errorf("%s: %w", name, domain.ErrMissingDependency)
	}

	return nil
}
