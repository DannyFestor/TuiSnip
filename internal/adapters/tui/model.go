package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	forcedQuitKey    = "ctrl+c"
	operationList    = "list snippets"
	operationCopy    = "copy"
	hintWidthDivisor = 2
)

type Model struct {
	//nolint:containedctx // Bubble Tea's Update has no context parameter, so the program context travels with the model.
	ctx             context.Context
	lister          FolderSnippetsLister
	copier          SnippetCopier
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
	status          string
}

func New(ctx context.Context, deps Deps) (Model, error) {
	err := errors.Join(
		domain.RequireDependency("lister", deps.Lister),
		domain.RequireDependency("copier", deps.Copier),
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
		logger:          deps.Logger,
		keys:            newBindings(deps.Settings),
		styles:          styles,
		screen:          size{width: 0, height: 0},
		layout:          arrange(size{width: 0, height: 0}, paneFolders, paneFolders),
		focus:           paneFolders,
		selectionHolder: paneFolders,
		folders:         folderPane{rootSnippetCount: 0},
		tags:            tagPane{},
		list:            snippetList{snippets: nil, cursor: 0, offset: 0, height: 0},
		preview:         newSnippetPane(styles, deps.Settings.Location),
		status:          "",
	}, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.loadSnippets())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screen = size{width: msg.Width, height: msg.Height}

		return m.arranged(), nil
	case tea.BackgroundColorMsg:
		m.preview = m.preview.withCodeStyle(codeStyleFor(msg))

		return m, nil
	case tea.KeyPressMsg:
		return m.pressed(msg)
	case snippetsLoadedMsg:
		return m.snippetsLoaded(msg), nil
	case copyFinishedMsg:
		return m.copyFinished(msg)
	}

	return m, nil
}

func (m Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true

	return view
}

func (m Model) loadSnippets() tea.Cmd {
	return func() tea.Msg {
		snippets, err := m.lister.Run(m.ctx, browse.SnippetsInFolderInput{FolderID: domain.FolderID{}})

		return snippetsLoadedMsg{snippets: snippets, err: err}
	}
}

func (m Model) pressed(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == forcedQuitKey || key.Matches(msg, m.keys.quit) {
		return m, tea.Quit
	}

	if navigate, ok := m.keys.navigationFor(msg); ok {
		m.focus = navigate(m.focus, m.selectionHolder)

		return m.arranged(), nil
	}

	if move, ok := m.keys.movementFor(msg); ok {
		return m.moved(move), nil
	}

	if m.copyRequested(msg) {
		return m, m.copySelected()
	}

	return m, nil
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

	in := snippet.CopyInput{SnippetID: selected.ID()}

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
	next.list = m.list.withSnippets(msg.snippets)
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

func (m Model) render() string {
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

func (m Model) hint() string {
	if m.layout.single {
		return tooSmallHint
	}

	return hintFor(m.focusedHints(), m.screen.width/hintWidthDivisor)
}

func (m Model) focusedHints() []key.Binding {
	switch m.focus {
	case paneList:
		return []key.Binding{m.keys.listCopy}
	case paneSnippet:
		return []key.Binding{m.keys.paneCopy}
	case paneFolders, paneTags:
	}

	return nil
}

func requirePointer[T any](name string, pointer *T) error {
	if pointer == nil {
		return fmt.Errorf("%s: %w", name, domain.ErrMissingDependency)
	}

	return nil
}
