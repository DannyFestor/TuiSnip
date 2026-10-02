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
	edit            editSession
	search          searchPopup
	confirm         confirmation
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
		edit:            noEditSession(),
		search:          noSearchPopup(),
		confirm:         noConfirmation(),
		status:          "",
	}, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.loadSnippets(domain.SnippetID{}))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screen = size{width: msg.Width, height: msg.Height}

		return m.arranged(), nil
	case tea.BackgroundColorMsg:
		m.preview = m.preview.withCodeStyle(codeStyleFor(msg))
		m.search = m.search.withCodeStyle(codeStyleFor(msg))

		return m, nil
	case tea.KeyPressMsg:
		return m.pressed(msg)
	case tea.PasteMsg:
		return m.pasted(msg)
	case snippetsLoadedMsg:
		return m.snippetsLoaded(msg), nil
	case copyFinishedMsg:
		return m.copyFinished(msg)
	case snippetCreatedMsg:
		return m.snippetCreated(msg)
	case searchFinishedMsg:
		return m.searchFinished(msg), nil
	}

	return m, nil
}

func (m Model) View() tea.View {
	view := tea.NewView(m.render())
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
		return m.quitRequested()
	}

	switch {
	case m.confirm.open:
		return m.confirmPressed(msg)
	case m.edit.open:
		return m.editUpdated(m.edit.update(msg))
	case m.search.open:
		return m.searchUpdated(m.search.update(msg))
	}

	return m.panePressed(msg)
}

func (m Model) pasted(msg tea.PasteMsg) (tea.Model, tea.Cmd) {
	switch {
	case m.confirm.open:
		return m, nil
	case m.edit.open:
		return m.editUpdated(m.edit.update(msg))
	case m.search.open:
		return m.searchUpdated(m.search.update(msg))
	}

	return m, nil
}

func (m Model) panePressed(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.newSnippet):
		return m.editOpened()
	case key.Matches(msg, m.keys.openSearch):
		return m.searchOpened()
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

func (m Model) quitRequested() (tea.Model, tea.Cmd) {
	if m.edit.hasUnsavedChanges() {
		return m.confirming(quitQuestion, quitting), nil
	}

	return m, tea.Quit
}

func (m Model) confirming(question string, onYes func(Model) (Model, tea.Cmd)) Model {
	m.confirm = newConfirmation(m.keys.confirm, question, onYes)

	return m
}

func (m Model) confirmPressed(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	onYes := m.confirm.onYes

	switch m.confirm.answer(msg) {
	case answerYes:
		m.confirm = noConfirmation()

		return onYes(m)
	case answerNo:
		m.confirm = noConfirmation()
	case answerPending:
	}

	return m, nil
}

func (m Model) editOpened() (tea.Model, tea.Cmd) {
	edit, cmd := newEditSession(m.keys.editor)
	m.edit = edit.resized(shareOf(m.screen, editOverlayPercent))

	return m, cmd
}

func (m Model) editUpdated(edit editSession, request editRequest, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.edit = edit

	switch request {
	case editSaves:
		return m, tea.Batch(cmd, m.createSnippet())
	case editCancels:
		return m.editClosed(), cmd
	case editAsksDiscard:
		return m.confirming(discardQuestion, discardingEdit), cmd
	case editRefusesPaste:
		m.status = pasteHasTabsText
	case editStays:
	}

	return m, cmd
}

func (m Model) editClosed() Model {
	m.edit = noEditSession()

	return m
}

func (m Model) createSnippet() tea.Cmd {
	in := m.edit.input()

	return func() tea.Msg {
		created, err := m.creator.Run(m.ctx, in)

		return snippetCreatedMsg{snippet: created, err: err}
	}
}

func (m Model) snippetCreated(msg snippetCreatedMsg) (tea.Model, tea.Cmd) {
	edit, outcome := m.edit.saved(msg.err)
	m.edit = edit

	switch outcome {
	case saveSucceeded:
		return m.editClosed(), m.loadSnippets(msg.snippet.ID())
	case saveRejected:
		m.status = edit.notice
	case saveFailed:
		return m.failed(operationSave, msg.err), nil
	}

	return m, nil
}

func (m Model) searchOpened() (tea.Model, tea.Cmd) {
	popup, cmd := newSearchPopup(m.keys.search, m.preview.cleared(), m.list.snippets)
	m.search = popup.resized(m.screen)

	return m, cmd
}

func (m Model) searchUpdated(popup searchPopup, request searchRequest, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.search = popup

	switch request {
	case searchQueries:
		return m, tea.Batch(cmd, m.querySnippets(popup.text()))
	case searchReveals:
		return m.searchRevealed()
	case searchCopies:
		return m.searchCopied()
	case searchCloses:
		return m.searchClosed(), cmd
	case searchStays:
	}

	return m, cmd
}

func (m Model) querySnippets(text string) tea.Cmd {
	in := search.QueryInput{Text: text}

	return func() tea.Msg {
		hits, err := m.searcher.Run(m.ctx, in)

		return searchFinishedMsg{text: in.Text, hits: hits, err: err}
	}
}

func (m Model) searchFinished(msg searchFinishedMsg) Model {
	if msg.err != nil {
		return m.failed(operationSearch, msg.err)
	}

	m.search = m.search.withHits(msg.text, msg.hits)

	return m
}

func (m Model) searchRevealed() (tea.Model, tea.Cmd) {
	selected, ok := m.search.selected()
	if !ok {
		return m, nil
	}

	next := m.searchClosed()
	next.focus = paneSnippet
	next.selectionHolder = paneFolders

	return next.arranged(), next.loadSnippets(selected.ID())
}

func (m Model) searchCopied() (tea.Model, tea.Cmd) {
	selected, ok := m.search.selected()
	if !ok {
		return m, nil
	}

	return m.searchClosed(), m.copySnippet(selected.ID())
}

func (m Model) searchClosed() Model {
	m.search = noSearchPopup()

	return m
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

	if m.edit.open {
		m.edit = m.edit.resized(shareOf(m.screen, editOverlayPercent))
	}

	if m.search.open {
		m.search = m.search.resized(m.screen)
	}

	return m
}

func (m Model) tallLeft() pane {
	if m.focus.inLeftColumn() {
		return m.focus
	}

	return m.selectionHolder
}

func (m Model) render() string {
	view := m.mainScreen()
	for _, layer := range m.overlayViews() {
		view = overlaid(view, m.screen, layer)
	}

	return view
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

func (m Model) overlayViews() []string {
	views := make([]string, 0, overlayLayers)

	if m.edit.open {
		views = append(views, m.edit.view(m.styles, shareOf(m.screen, editOverlayPercent)))
	}

	if m.search.open {
		views = append(views, m.search.view(m.styles))
	}

	if m.confirm.open {
		views = append(views, m.confirm.view(m.styles))
	}

	return views
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
	overlayOpen := m.edit.open || m.search.open || m.confirm.open
	if m.layout.single && !overlayOpen {
		return tooSmallHint
	}

	return hintFor(m.focusedHints(), m.screen.width/hintWidthDivisor)
}

func (m Model) focusedHints() []key.Binding {
	switch {
	case m.confirm.open:
		return m.confirm.hints()
	case m.edit.open:
		return m.edit.hints()
	case m.search.open:
		return m.search.hints()
	}

	return m.keys.paneHints(m.focus)
}

func quitting(m Model) (Model, tea.Cmd) {
	return m, tea.Quit
}

func discardingEdit(m Model) (Model, tea.Cmd) {
	return m.editClosed(), nil
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
