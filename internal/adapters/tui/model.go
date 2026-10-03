package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetlist"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetpane"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	operationList    = "list snippets"
	operationCopy    = "copy"
	operationSave    = "save snippet"
	operationSearch  = "search"
	hintWidthDivisor = 2
)

type Model struct {
	//nolint:containedctx // Bubble Tea's Update has no context parameter, so the program context travels with the model.
	ctx             context.Context
	lister          FolderSnippetsLister
	copier          SnippetCopier
	creator         SnippetCreator
	searcher        SnippetSearcher
	logger          *slog.Logger
	keys            bindings
	styles          look.Styles
	screen          look.Size
	layout          layout
	focus           pane
	selectionHolder pane
	folders         folderPane
	tags            tagPane
	list            snippetlist.List
	preview         snippetpane.Pane
	overlays        overlayStack
	status          string
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

	styles := look.NewStyles()

	return Model{
		ctx:             ctx,
		lister:          deps.Lister,
		copier:          deps.Copier,
		creator:         deps.Creator,
		searcher:        deps.Searcher,
		logger:          deps.Logger,
		keys:            newBindings(deps.Settings),
		styles:          styles,
		screen:          look.Size{Width: 0, Height: 0},
		layout:          arrange(look.Size{Width: 0, Height: 0}, paneFolders, paneFolders),
		focus:           paneFolders,
		selectionHolder: paneFolders,
		folders:         folderPane{rootSnippetCount: 0},
		tags:            tagPane{},
		list:            snippetlist.New(deps.Settings.Keys, styles, snippetlist.Language),
		preview:         snippetpane.New(deps.Settings.Keys, styles, deps.Settings.Location),
		overlays:        newOverlayStack(),
		status:          "",
	}, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.loadSnippets(domain.SnippetID{}))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screen = look.SizeOf(msg)

		return m.arranged().overlaysUpdated(msg)
	case tea.BackgroundColorMsg:
		m.preview, _, _ = m.preview.Update(msg)

		return m.overlaysUpdated(msg)
	case tea.KeyPressMsg:
		return m.pressed(msg)
	case tea.PasteMsg, snippetCreatedMsg:
		return m.overlaysUpdated(msg)
	case snippetsLoadedMsg:
		return m.snippetsLoaded(msg), nil
	case copyFinishedMsg:
		return m.copyFinished(msg)
	case searchFinishedMsg:
		return m.searchFinished(msg)
	}

	return m, nil
}

func (m Model) View() tea.View {
	view := tea.NewView(m.overlays.Render(m.mainScreen()))
	view.AltScreen = true

	return view
}

func (m Model) loadSnippets(selecting domain.SnippetID) tea.Cmd {
	return func() tea.Msg {
		snippets, err := m.lister.Run(m.ctx, browse.SnippetsInFolderInput{FolderID: domain.FolderID{}})

		return snippetsLoadedMsg{snippets: snippets, selecting: selecting, err: err}
	}
}

func (m Model) pressed(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == m.keys.forcedQuit {
		return m.settled(m.overlays.Offered(outcome.QuitAsked{}))
	}

	if m.overlays.Open() {
		return m.overlaysUpdated(msg)
	}

	return m.panePressed(msg)
}

func (m Model) panePressed(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	global := m.keys.global

	switch {
	case global.Matches(msg, binding.Quit):
		return m, tea.Quit
	case global.Matches(msg, binding.NewSnippet):
		return m.opened(newEditSession(m.keys.editor, m.keys.confirm, m.styles))
	case global.Matches(msg, binding.Search):
		return m.opened(newSearchPopup(m.keys.table, m.styles, m.preview.Cleared(), m.list.Snippets()))
	}

	if navigate, ok := m.keys.navigationFor(msg); ok {
		m.focus = navigate(m.focus, m.selectionHolder)

		return m.arranged(), nil
	}

	return m.focusedPressed(msg)
}

func (m Model) focusedPressed(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch m.focus {
	case paneList:
		return m.listUpdated(msg)
	case paneSnippet:
		return m.previewUpdated(msg)
	case paneFolders, paneTags:
	}

	return m, nil
}

func (m Model) listUpdated(msg tea.Msg) (Model, tea.Cmd) {
	next := m
	list, outcomes, cmd := m.list.Update(msg)
	next.list = list

	return next.previewSelected().concludedAll(outcomes, cmd)
}

func (m Model) previewUpdated(msg tea.Msg) (Model, tea.Cmd) {
	next := m
	preview, outcomes, cmd := m.preview.Update(msg)
	next.preview = preview

	return next.concludedAll(outcomes, cmd)
}

func (m Model) opened(opening outcomeOverlay, openCmd tea.Cmd) (Model, tea.Cmd) {
	next, cmd := m.settled(m.overlays.Pushed(opening))

	return next, tea.Batch(openCmd, cmd)
}

func (m Model) overlaysUpdated(msg tea.Msg) (Model, tea.Cmd) {
	return m.settled(m.overlays.Update(msg))
}

func (m Model) settled(overlays overlayStack, outcomes []outcome.Outcome, cmd tea.Cmd) (Model, tea.Cmd) {
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
		return m.failed(operationSave, reported.Err), nil
	case outcome.NoticeShown:
		m.status = reported.Text
	case outcome.SearchTyped:
		return m, m.querySnippets(reported.Text)
	case outcome.SnippetRevealed:
		return m.revealed(reported.ID)
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

		return snippetCreatedMsg{snippet: created, err: err}
	}
}

func (m Model) querySnippets(text string) tea.Cmd {
	in := search.QueryInput{Text: text}

	return func() tea.Msg {
		hits, err := m.searcher.Run(m.ctx, in)

		return searchFinishedMsg{text: in.Text, hits: hits, err: err}
	}
}

func (m Model) searchFinished(msg searchFinishedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		return m.failed(operationSearch, msg.err), nil
	}

	return m.overlaysUpdated(msg)
}

func (m Model) revealed(id domain.SnippetID) (Model, tea.Cmd) {
	next := m
	next.focus = paneSnippet
	next.selectionHolder = paneFolders

	return next.arranged(), next.loadSnippets(id)
}

func (m Model) copySnippet(id domain.SnippetID) tea.Cmd {
	in := snippet.CopyInput{SnippetID: id}

	return func() tea.Msg {
		result, err := m.copier.Run(m.ctx, in)

		return copyFinishedMsg{result: result, err: err}
	}
}

func (m Model) snippetsLoaded(msg snippetsLoadedMsg) Model {
	if msg.err != nil {
		return m.failed(operationList, msg.err)
	}

	next := m
	next.list = m.list.WithSnippets(msg.snippets).WithCursorOn(msg.selecting)
	next.folders = m.folders.withRootSnippetCount(len(msg.snippets))

	return next.previewSelected()
}

func (m Model) copyFinished(msg copyFinishedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		return m.failed(operationCopy, msg.err), nil
	}

	text, err := deliveryText(msg.result.Delivery)
	if err != nil {
		return m.failed(operationCopy, err), nil
	}

	m.status = text
	if msg.result.Delivery == domain.CopyDeliverySentToTerminal {
		return m, tea.SetClipboard(msg.result.Content.String())
	}

	return m, nil
}

func (m Model) failed(operation string, err error) Model {
	if errors.Is(err, context.Canceled) {
		m.logger.DebugContext(m.ctx, "operation cancelled", slog.String(keyOperation, operation))

		return m
	}

	m.logger.ErrorContext(m.ctx, "operation failed", slog.String(keyOperation, operation), slog.Any(keyError, err))
	m.status = failureText(err)

	return m
}

func (m Model) previewSelected() Model {
	selected, ok := m.list.Selected()
	if ok {
		m.preview = m.preview.Showing(selected)
	} else {
		m.preview = m.preview.Cleared()
	}

	return m
}

func (m Model) arranged() Model {
	m.layout = arrange(m.screen, m.focus, m.tallLeft())
	m.list, _, _ = m.list.Update(look.Resized{Box: m.layout.list.Inner()})
	m.preview, _, _ = m.preview.Update(look.Resized{Box: m.layout.snippet.Inner()})

	return m
}

func (m Model) tallLeft() pane {
	if m.focus.inLeftColumn() {
		return m.focus
	}

	return m.selectionHolder
}

func (m Model) mainScreen() string {
	status := statusLine(m.styles, m.status, m.hint(), m.screen.Width)
	if m.layout.single {
		return m.paneFrame(m.focus) + "\n" + status
	}

	left := lipgloss.JoinVertical(lipgloss.Left, m.paneFrame(paneFolders), m.paneFrame(paneTags))
	panes := lipgloss.JoinHorizontal(lipgloss.Top, left, m.paneFrame(paneList), m.paneFrame(paneSnippet))

	return panes + "\n" + status
}

func (m Model) paneFrame(p pane) string {
	outer := m.layout.of(p)
	paneStyle := m.paneStyle(p)

	return look.Frame(paneStyle, paneTitle(p), m.paneBody(p, paneStyle, outer.Inner().Width), outer)
}

func (m Model) paneStyle(p pane) look.FrameStyle {
	if p == m.focus {
		return m.styles.Focused
	}

	return m.styles.Unfocused
}

func (m Model) paneBody(p pane, paneStyle look.FrameStyle, width int) string {
	switch p {
	case paneFolders:
		return m.folders.body(paneStyle, width)
	case paneTags:
		return m.tags.body(m.styles)
	case paneList:
		return m.list.View(paneStyle)
	case paneSnippet:
		return m.preview.View()
	}

	return ""
}

func (m Model) focusedHints() []key.Binding {
	switch m.focus {
	case paneFolders:
		return m.keys.folders.ShortHelp()
	case paneTags:
		return m.keys.tags.ShortHelp()
	case paneList:
		return m.list.ShortHelp()
	case paneSnippet:
		return m.preview.ShortHelp()
	}

	return nil
}

func (m Model) hint() string {
	width := m.screen.Width / hintWidthDivisor

	switch {
	case m.overlays.Open():
		return hintFor(m.overlays.ShortHelp(), width)
	case m.layout.single:
		return tooSmallHint
	}

	return hintFor(m.focusedHints(), width)
}

func paneTitle(p pane) string {
	switch p {
	case paneFolders:
		return folderPaneTitle
	case paneTags:
		return tagPaneTitle
	case paneList:
		return snippetListTitle
	case paneSnippet:
		return snippetPaneTitle
	}

	return ""
}

func requirePointer[T any](name string, pointer *T) error {
	if pointer == nil {
		return fmt.Errorf("%s: %w", name, domain.ErrMissingDependency)
	}

	return nil
}
