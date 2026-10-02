package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/overlay"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	forcedQuitKey    = "ctrl+c"
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
	styles          styleSet
	screen          size
	layout          layout
	focus           pane
	selectionHolder pane
	folders         folderPane
	tags            tagPane
	list            snippetList
	preview         snippetPane
	overlays        overlay.Stack
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

	styles := newStyleSet()

	return Model{
		ctx:             ctx,
		lister:          deps.Lister,
		copier:          deps.Copier,
		creator:         deps.Creator,
		searcher:        deps.Searcher,
		logger:          deps.Logger,
		keys:            newBindings(deps.Settings),
		styles:          styles,
		screen:          size{width: 0, height: 0},
		layout:          arrange(size{width: 0, height: 0}, paneFolders, paneFolders),
		focus:           paneFolders,
		selectionHolder: paneFolders,
		folders:         folderPane{rootSnippetCount: 0},
		tags:            tagPane{},
		list:            newSnippetList(languageOf),
		preview:         newSnippetPane(styles, deps.Settings.Location),
		overlays:        overlay.NewStack(),
		status:          "",
	}, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.loadSnippets(domain.SnippetID{}))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screen = sizeOf(msg)

		return m.arranged().overlaysUpdated(msg)
	case tea.BackgroundColorMsg:
		m.preview = m.preview.withCodeStyle(codeStyleFor(msg))

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
	if msg.String() == forcedQuitKey {
		return m.settled(m.overlays.Offered(quitAsked{}))
	}

	if m.overlays.Open() {
		return m.overlaysUpdated(msg)
	}

	return m.panePressed(msg)
}

func (m Model) panePressed(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.newSnippet):
		return m.opened(newEditSession(m.keys.editor, m.keys.confirm, m.styles))
	case key.Matches(msg, m.keys.openSearch):
		return m.opened(newSearchPopup(m.keys.search, m.styles, m.preview.cleared(), m.list.snippets))
	case m.copyRequested(msg):
		return m, m.copySelected()
	}

	if navigate, ok := m.keys.navigationFor(msg); ok {
		m.focus = navigate(m.focus, m.selectionHolder)

		return m.arranged(), nil
	}

	if move, ok := m.keys.movementFor(msg); ok {
		return m.moved(move), nil
	}

	return m, nil
}

func (m Model) opened(opening overlay.Overlay, openCmd tea.Cmd) (Model, tea.Cmd) {
	next, cmd := m.settled(m.overlays.Pushed(opening))

	return next, tea.Batch(openCmd, cmd)
}

func (m Model) overlaysUpdated(msg tea.Msg) (Model, tea.Cmd) {
	return m.settled(m.overlays.Update(msg))
}

func (m Model) settled(overlays overlay.Stack, outcomes []overlay.Outcome, cmd tea.Cmd) (Model, tea.Cmd) {
	m.overlays = overlays
	cmds := make([]tea.Cmd, 0, 1+len(outcomes))
	cmds = append(cmds, cmd)

	for _, outcome := range outcomes {
		var concludedCmd tea.Cmd

		m, concludedCmd = m.concluded(outcome)
		cmds = append(cmds, concludedCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) concluded(outcome overlay.Outcome) (Model, tea.Cmd) {
	switch outcome := outcome.(type) {
	case saveRequested:
		return m, m.createSnippet(outcome.input)
	case snippetSaved:
		return m, m.loadSnippets(outcome.id)
	case saveFailed:
		return m.failed(operationSave, outcome.err), nil
	case noticeShown:
		m.status = outcome.text
	case searchTyped:
		return m, m.querySnippets(outcome.text)
	case snippetRevealed:
		return m.revealed(outcome.id)
	case copyRequested:
		return m, m.copySnippet(outcome.id)
	case quitAsked, quitConfirmed:
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

func (m Model) moved(move movement) Model {
	switch m.focus {
	case paneList:
		m.list = m.list.moved(move)

		return m.previewSelected()
	case paneSnippet:
		m.preview = m.preview.moved(move)
	case paneFolders, paneTags:
	}

	return m
}

func (m Model) copyRequested(msg tea.KeyPressMsg) bool {
	return (m.focus == paneList && key.Matches(msg, m.keys.listCopy)) ||
		(m.focus == paneSnippet && key.Matches(msg, m.keys.paneCopy))
}

func (m Model) copySelected() tea.Cmd {
	selected, ok := m.list.selected()
	if !ok {
		return nil
	}

	return m.copySnippet(selected.ID())
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
	next.list = m.list.withSnippets(msg.snippets).withCursorOn(msg.selecting)
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
	selected, ok := m.list.selected()
	if ok {
		m.preview = m.preview.showing(selected)
	} else {
		m.preview = m.preview.cleared()
	}

	return m
}

func (m Model) arranged() Model {
	m.layout = arrange(m.screen, m.focus, m.tallLeft())
	m.list = m.list.resized(innerSize(m.layout.list).height)
	m.preview = m.preview.resized(innerSize(m.layout.snippet))

	return m
}

func (m Model) tallLeft() pane {
	if m.focus.inLeftColumn() {
		return m.focus
	}

	return m.selectionHolder
}

func (m Model) mainScreen() string {
	status := statusLine(m.styles, m.status, m.hint(), m.screen.width)
	if m.layout.single {
		return m.paneFrame(m.focus) + "\n" + status
	}

	left := lipgloss.JoinVertical(lipgloss.Left, m.paneFrame(paneFolders), m.paneFrame(paneTags))
	panes := lipgloss.JoinHorizontal(lipgloss.Top, left, m.paneFrame(paneList), m.paneFrame(paneSnippet))

	return panes + "\n" + status
}

func (m Model) paneFrame(p pane) string {
	outer := m.layout.of(p)
	look := m.styles.lookFor(p, m.focus)

	return frame(look, paneTitle(p), m.paneBody(p, look, innerSize(outer).width), outer)
}

func (m Model) paneBody(p pane, look paneLook, width int) string {
	switch p {
	case paneFolders:
		return m.folders.body(look, width)
	case paneTags:
		return m.tags.body(m.styles)
	case paneList:
		return m.list.body(m.styles, look, width, m.keys.emptyHints)
	case paneSnippet:
		return m.preview.body(m.styles, m.keys.emptyHints)
	}

	return ""
}

func (m Model) hint() string {
	width := m.screen.width / hintWidthDivisor

	switch {
	case m.overlays.Open():
		return hintFor(m.overlays.Hints(), width)
	case m.layout.single:
		return tooSmallHint
	}

	return hintFor(m.keys.paneHints(m.focus), width)
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
