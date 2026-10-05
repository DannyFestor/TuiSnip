package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/savegate"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/searchpopup"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	operationList                 = "list snippets"
	operationTree                 = "list folders"
	operationTags                 = "list tags"
	operationCopy                 = "copy"
	operationSave                 = "save snippet"
	operationSearch               = "search"
	operationCreateFolder         = "create folder"
	operationRenameFolder         = "rename folder"
	operationPreviewDelete        = "preview folder delete"
	operationDeleteFolder         = "delete folder"
	operationSaveSortOrder        = "save sort order"
	operationCycleSortOrder       = "cycle sort order"
	operationShowSortOrder        = "show sort order"
	operationSaveCollapsedFolders = "save collapsed folders"
)

type Model struct {
	//nolint:containedctx // Bubble Tea's Update has no context parameter, so the program context travels with the model.
	ctx                   context.Context
	lister                FolderSnippetsLister
	treeLister            FolderTreeLister
	tagLister             TagLister
	tagSnippetsLister     TagSnippetsLister
	copier                SnippetCopier
	creator               SnippetCreator
	updater               SnippetUpdater
	searcher              SnippetSearcher
	folderCreator         FolderCreator
	folderRenamer         FolderRenamer
	folderDeletePreviewer FolderDeletePreviewer
	folderDeleter         FolderDeleter
	sortOrderSaver        SortOrderSaver
	collapsedFoldersSaver CollapsedFoldersSaver
	collapsedFoldersGate  *savegate.Gate
	sortOrder             domain.SortOrder
	logger                *slog.Logger
	forcedQuitKey         string
	afterCopy             tea.Cmd
	overlays              outcome.Stack
}

func New(ctx context.Context, deps Deps) (Model, error) {
	model, err := modelEndingCopyWith(ctx, deps, nil)
	if err != nil {
		return Model{}, fmt.Errorf("tui.New: %w", err)
	}

	return model, nil
}

func NewQuittingAfterCopy(ctx context.Context, deps Deps) (Model, error) {
	model, err := modelEndingCopyWith(ctx, deps, tea.Quit)
	if err != nil {
		return Model{}, fmt.Errorf("tui.NewQuittingAfterCopy: %w", err)
	}

	return model, nil
}

func modelEndingCopyWith(ctx context.Context, deps Deps, afterCopy tea.Cmd) (Model, error) {
	err := missingDependencies(deps)
	if err != nil {
		return Model{}, err
	}

	mainScreen, err := mainscreen.New(
		deps.Settings.Keys,
		look.NewStyles(),
		deps.Settings.Location,
		deps.Settings.Remembered,
	)
	if err != nil {
		return Model{}, fmt.Errorf("main screen: %w", err)
	}

	overlays, _, _ := outcome.NewStack().Pushed(mainScreen)

	return Model{
		ctx:                   ctx,
		lister:                deps.Lister,
		treeLister:            deps.TreeLister,
		tagLister:             deps.TagLister,
		tagSnippetsLister:     deps.TagSnippetsLister,
		copier:                deps.Copier,
		creator:               deps.Creator,
		updater:               deps.Updater,
		searcher:              deps.Searcher,
		folderCreator:         deps.FolderCreator,
		folderRenamer:         deps.FolderRenamer,
		folderDeletePreviewer: deps.FolderDeletePreviewer,
		folderDeleter:         deps.FolderDeleter,
		sortOrderSaver:        deps.SortOrderSaver,
		collapsedFoldersSaver: deps.CollapsedFoldersSaver,
		collapsedFoldersGate:  &savegate.Gate{},
		sortOrder:             deps.Settings.Remembered.SortOrder,
		logger:                deps.Logger,
		forcedQuitKey:         deps.Settings.ForcedQuitKey,
		afterCopy:             afterCopy,
		overlays:              overlays,
	}, nil
}

func missingDependencies(deps Deps) error {
	return errors.Join(
		domain.RequireDependency("lister", deps.Lister),
		domain.RequireDependency("treeLister", deps.TreeLister),
		domain.RequireDependency("tagLister", deps.TagLister),
		domain.RequireDependency("tagSnippetsLister", deps.TagSnippetsLister),
		domain.RequireDependency("copier", deps.Copier),
		domain.RequireDependency("creator", deps.Creator),
		domain.RequireDependency("updater", deps.Updater),
		domain.RequireDependency("searcher", deps.Searcher),
		domain.RequireDependency("folderCreator", deps.FolderCreator),
		domain.RequireDependency("folderRenamer", deps.FolderRenamer),
		domain.RequireDependency("folderDeletePreviewer", deps.FolderDeletePreviewer),
		domain.RequireDependency("folderDeleter", deps.FolderDeleter),
		domain.RequireDependency("sortOrderSaver", deps.SortOrderSaver),
		domain.RequireDependency("collapsedFoldersSaver", deps.CollapsedFoldersSaver),
		requirePointer("logger", deps.Logger),
		requirePointer("location", deps.Settings.Location),
	)
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
		m.loadTree(),
		m.loadTags(),
		m.loadSnippets(browseselection.Selection{}, domain.SnippetID{}),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.pressed(msg)
	case tea.WindowSizeMsg, tea.BackgroundColorMsg, tea.PasteMsg,
		editoverlay.SaveFinished, searchpopup.HitsFound, mainscreen.SnippetsLoaded, mainscreen.TreeLoaded,
		mainscreen.TreeChanged, mainscreen.FolderDeletePreviewed, mainscreen.TagsLoaded:
		return m.overlaysUpdatedSharingTree(msg)
	case folderTreeChangedMsg:
		return m, tea.Batch(m.loadTreeSelecting(msg.selecting), m.loadTags())
	case folderRenamedMsg:
		return m, m.loadTree()
	case operationFailedMsg:
		return m.failed(msg.operation, msg.err)
	case listFailedMsg:
		return m.failed(operationList, msg.err)
	case treeFailedMsg:
		return m.failed(operationTree, msg.err)
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

func (m Model) loadTree() tea.Cmd {
	return m.listTree(func(tree browse.Tree) tea.Msg { return mainscreen.TreeLoaded{Tree: tree} })
}

func (m Model) loadTreeSelecting(id domain.FolderID) tea.Cmd {
	return m.listTree(func(tree browse.Tree) tea.Msg { return mainscreen.TreeChanged{Tree: tree, Selecting: id} })
}

func (m Model) listTree(loaded func(tree browse.Tree) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		tree, err := m.treeLister.Run(m.ctx, browse.FolderTreeInput{})
		if err != nil {
			return treeFailedMsg{err: err}
		}

		return loaded(tree)
	}
}

func (m Model) loadTags() tea.Cmd {
	return func() tea.Msg {
		tags, err := m.tagLister.Run(m.ctx, browse.TagListInput{})
		if err != nil {
			return operationFailedMsg{operation: operationTags, err: err}
		}

		return mainscreen.TagsLoaded{Tags: tags}
	}
}

func (m Model) loadSnippets(selection browseselection.Selection, selecting domain.SnippetID) tea.Cmd {
	order := m.sortOrder
	loaded := func(snippets []domain.Snippet, err error) tea.Msg {
		if err != nil {
			return listFailedMsg{err: err}
		}

		return mainscreen.SnippetsLoaded{Selection: selection, Snippets: snippets, Selecting: selecting, Order: order}
	}

	if tagID, ok := selection.Tag(); ok {
		return func() tea.Msg {
			return loaded(m.tagSnippetsLister.Run(m.ctx, browse.SnippetsWithTagInput{TagID: tagID, Order: order}))
		}
	}

	folderID, _ := selection.Folder()

	return func() tea.Msg {
		return loaded(m.lister.Run(m.ctx, browse.SnippetsInFolderInput{FolderID: folderID, Order: order}))
	}
}

func (m Model) sortCycled(asked outcome.SortCycleAsked) (Model, tea.Cmd) {
	order, err := m.sortOrder.Next()
	if err != nil {
		return m.failed(operationCycleSortOrder, err)
	}

	next := m
	next.sortOrder = order

	return next, tea.Batch(next.saveSortOrder(), next.loadSnippets(asked.Selection, asked.Selecting))
}

func (m Model) saveSortOrder() tea.Cmd {
	order := m.sortOrder

	return remembering(operationSaveSortOrder, func() error {
		return m.sortOrderSaver.SaveSortOrder(m.ctx, order)
	})
}

func (m Model) saveCollapsedFolders(ids []domain.FolderID) tea.Cmd {
	ticket := m.collapsedFoldersGate.Ticket()

	return remembering(operationSaveCollapsedFolders, func() error {
		return m.collapsedFoldersGate.Save(ticket, func() error {
			return m.collapsedFoldersSaver.SaveCollapsedFolders(m.ctx, ids)
		})
	})
}

func remembering(operation string, save func() error) tea.Cmd {
	return func() tea.Msg {
		err := save()
		if err != nil {
			return operationFailedMsg{operation: operation, err: err}
		}

		return nil
	}
}

func (m Model) pressed(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if msg.String() == m.forcedQuitKey {
		return m.settled(m.overlays.Offered(outcome.QuitAsked{}))
	}

	return m.overlaysUpdated(msg)
}

func (m Model) overlaysUpdatedSharingTree(msg tea.Msg) (Model, tea.Cmd) {
	tree, ok := arrivedTree(msg)
	if !ok {
		return m.overlaysUpdated(msg)
	}

	return m.treeArrived(msg, tree)
}

func arrivedTree(msg tea.Msg) (browse.Tree, bool) {
	switch msg := msg.(type) {
	case mainscreen.TreeLoaded:
		return msg.Tree, true
	case mainscreen.TreeChanged:
		return msg.Tree, true
	}

	return browse.Tree{}, false
}

// The Search popup can't import mainscreen, so it hears about the tree through its own message.
func (m Model) treeArrived(msg tea.Msg, tree browse.Tree) (Model, tea.Cmd) {
	screenUpdated, screenCmd := m.overlaysUpdated(msg)
	popupUpdated, popupCmd := screenUpdated.overlaysUpdated(searchpopup.TreeLoaded{Tree: tree})

	return popupUpdated, tea.Batch(screenCmd, popupCmd)
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
	case outcome.SaveRequested, outcome.UpdateRequested, outcome.SearchTyped, outcome.CopyRequested:
		return m, m.runSnippetAction(reported)
	case outcome.SnippetSaved, outcome.SnippetReloaded, outcome.SnippetRevealed, outcome.FolderSelected,
		outcome.TagSelected, outcome.SortCycleAsked:
		return m.relisted(reported)
	case outcome.SaveFailed, outcome.SortOrderRejected:
		return m.reportedFailure(reported)
	case outcome.NoticeShown:
		return m.shown(reported.Text)
	case outcome.FolderCreateRequested, outcome.FolderRenameRequested,
		outcome.FolderDeleteAsked, outcome.FolderDeleteRequested:
		return m, m.runFolderAction(reported)
	case outcome.CollapsedFoldersChanged:
		return m, m.saveCollapsedFolders(reported.IDs)
	case outcome.DiscardConfirmed:
	case outcome.QuitAsked, outcome.QuitConfirmed:
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) reportedFailure(reported outcome.Outcome) (Model, tea.Cmd) {
	switch reported := reported.(type) {
	case outcome.SaveFailed:
		return m.failed(operationSave, reported.Err)
	case outcome.SortOrderRejected:
		return m.failed(operationShowSortOrder, reported.Err)
	default:
		return m, nil
	}
}

func (m Model) relisted(reported outcome.Outcome) (Model, tea.Cmd) {
	switch reported := reported.(type) {
	case outcome.SnippetSaved:
		return m, m.reloadSelecting(browseselection.InFolder(reported.FolderID), reported.ID)
	case outcome.SnippetReloaded:
		return m, m.reloadSelecting(browseselection.InFolder(reported.FolderID), reported.ID)
	case outcome.SnippetRevealed:
		return m, m.loadSnippets(browseselection.InFolder(reported.FolderID), reported.ID)
	case outcome.FolderSelected:
		return m, m.loadSnippets(browseselection.InFolder(reported.ID), domain.SnippetID{})
	case outcome.TagSelected:
		return m, m.loadSnippets(browseselection.WithTag(reported.ID), domain.SnippetID{})
	case outcome.SortCycleAsked:
		return m.sortCycled(reported)
	default:
		return m, nil
	}
}

func (m Model) reloadSelecting(selection browseselection.Selection, id domain.SnippetID) tea.Cmd {
	return tea.Batch(m.loadTree(), m.loadSnippets(selection, id))
}

func (m Model) createSnippet(in snippet.CreateInput) tea.Cmd {
	return func() tea.Msg {
		created, err := m.creator.Run(m.ctx, in)

		return editoverlay.SaveFinished{Snippet: created, Err: err}
	}
}

func (m Model) updateSnippet(in snippet.UpdateInput) tea.Cmd {
	return func() tea.Msg {
		updated, err := m.updater.Run(m.ctx, in)

		return editoverlay.SaveFinished{Snippet: updated, Err: err}
	}
}

func (m Model) runSnippetAction(reported outcome.Outcome) tea.Cmd {
	switch reported := reported.(type) {
	case outcome.SaveRequested:
		return m.createSnippet(reported.Input)
	case outcome.UpdateRequested:
		return m.updateSnippet(reported.Input)
	case outcome.SearchTyped:
		return m.querySnippets(reported.Text)
	case outcome.CopyRequested:
		return m.copySnippet(reported.ID)
	default:
		return nil
	}
}

func (m Model) runFolderAction(reported outcome.Outcome) tea.Cmd {
	switch reported := reported.(type) {
	case outcome.FolderCreateRequested:
		return m.createFolder(reported.Input)
	case outcome.FolderRenameRequested:
		return m.renameFolder(reported.Input)
	case outcome.FolderDeleteAsked:
		return m.previewFolderDelete(folder.PreviewDeleteInput{FolderID: reported.ID})
	case outcome.FolderDeleteRequested:
		return m.deleteFolder(reported.Input, reported.ParentID)
	default:
		return nil
	}
}

func (m Model) createFolder(in folder.CreateInput) tea.Cmd {
	return func() tea.Msg {
		created, err := m.folderCreator.Run(m.ctx, in)
		if err != nil {
			return operationFailedMsg{operation: operationCreateFolder, err: err}
		}

		return folderTreeChangedMsg{selecting: created.ID()}
	}
}

func (m Model) previewFolderDelete(in folder.PreviewDeleteInput) tea.Cmd {
	return func() tea.Msg {
		preview, err := m.folderDeletePreviewer.Run(m.ctx, in)
		if err != nil {
			return operationFailedMsg{operation: operationPreviewDelete, err: err}
		}

		return mainscreen.FolderDeletePreviewed{Preview: preview}
	}
}

func (m Model) deleteFolder(in folder.DeleteInput, parentID domain.FolderID) tea.Cmd {
	return func() tea.Msg {
		err := m.folderDeleter.Run(m.ctx, in)
		if err != nil {
			return operationFailedMsg{operation: operationDeleteFolder, err: err}
		}

		return folderTreeChangedMsg{selecting: parentID}
	}
}

func (m Model) renameFolder(in folder.RenameInput) tea.Cmd {
	return func() tea.Msg {
		_, err := m.folderRenamer.Run(m.ctx, in)
		if err != nil {
			return operationFailedMsg{operation: operationRenameFolder, err: err}
		}

		return folderRenamedMsg{}
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

	return next, tea.Sequence(tea.Batch(cmd, terminalClipboard(msg.result)), m.afterCopy)
}

func terminalClipboard(result snippet.CopyResult) tea.Cmd {
	if result.Delivery != domain.CopyDeliverySentToTerminal {
		return nil
	}

	return tea.SetClipboard(result.Content.String())
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
