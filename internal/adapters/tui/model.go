package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/arrived"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/savegate"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/searchpopup"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	operationList                 = "list snippets"
	operationTree                 = "list folders"
	operationTags                 = "list tags"
	operationCopy                 = "copy"
	operationCapture              = "capture"
	operationSave                 = "save snippet"
	operationDuplicate            = "duplicate snippet"
	operationDeleteSnippet        = "delete snippet"
	operationMoveSnippet          = "move snippet"
	operationSearch               = "search"
	operationCreateFolder         = "create folder"
	operationRenameFolder         = "rename folder"
	operationDeleteFolder         = "delete folder"
	operationSetDefaultLanguage   = "set folder default language"
	operationMoveFolder           = "move folder"
	operationCreateTag            = "create tag"
	operationRenameTag            = "rename tag"
	operationPreviewTagDelete     = "preview tag delete"
	operationDeleteTag            = "delete tag"
	operationSaveSortOrder        = "save sort order"
	operationCycleSortOrder       = "cycle sort order"
	operationShowSortOrder        = "show sort order"
	operationSaveCollapsedFolders = "save collapsed folders"
	operationExternalEdit         = "edit in external editor"
)

type Model struct {
	//nolint:containedctx // Bubble Tea's Update has no context parameter, so the program context travels with the model.
	ctx                   context.Context
	lister                FolderSnippetsLister
	treeLister            FolderTreeLister
	tagLister             TagLister
	tagSnippetsLister     TagSnippetsLister
	copier                SnippetCopier
	trash                 trashActions
	creator               SnippetCreator
	capturer              SnippetCapturer
	updater               SnippetUpdater
	duplicator            SnippetDuplicator
	deleter               SnippetDeleter
	mover                 SnippetMover
	searcher              SnippetSearcher
	folderCreator         FolderCreator
	folderRenamer         FolderRenamer
	folderDeleter         FolderDeleter
	defaultLanguageSetter FolderDefaultLanguageSetter
	folderMover           FolderMover
	tagCreator            TagCreator
	tagRenamer            TagRenamer
	tagDeletePreviewer    TagDeletePreviewer
	tagDeleter            TagDeleter
	sortOrderSaver        SortOrderSaver
	collapsedFoldersSaver CollapsedFoldersSaver
	collapsedFoldersGate  *savegate.Gate
	externalEditor        ExternalEditor
	editedContentHandler  EditedContentHandler
	mouse                 mouse
	sortOrder             domain.SortOrder
	logger                *slog.Logger
	forcedQuitKey         string
	theme                 look.Theme
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

	mainScreen, err := newMainScreen(deps.Settings)
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
		trash:                 trashActionsOf(deps),
		creator:               deps.Creator,
		capturer:              deps.Capturer,
		updater:               deps.Updater,
		duplicator:            deps.Duplicator,
		deleter:               deps.Deleter,
		mover:                 deps.Mover,
		searcher:              deps.Searcher,
		folderCreator:         deps.FolderCreator,
		folderRenamer:         deps.FolderRenamer,
		folderDeleter:         deps.FolderDeleter,
		defaultLanguageSetter: deps.FolderDefaultLanguageSetter,
		folderMover:           deps.FolderMover,
		tagCreator:            deps.TagCreator,
		tagRenamer:            deps.TagRenamer,
		tagDeletePreviewer:    deps.TagDeletePreviewer,
		tagDeleter:            deps.TagDeleter,
		sortOrderSaver:        deps.SortOrderSaver,
		collapsedFoldersSaver: deps.CollapsedFoldersSaver,
		collapsedFoldersGate:  &savegate.Gate{},
		externalEditor:        deps.ExternalEditor,
		editedContentHandler:  deps.EditedContentHandler,
		mouse:                 mouse{on: deps.Settings.Mouse, clock: deps.Clock, clicks: pointer.Clicks{}},
		sortOrder:             deps.Settings.Remembered.SortOrder,
		logger:                deps.Logger,
		forcedQuitKey:         deps.Settings.ForcedQuitKey,
		theme:                 deps.Settings.Theme,
		afterCopy:             afterCopy,
		overlays:              overlays,
	}, nil
}

func newMainScreen(settings Settings) (mainscreen.Screen, error) {
	screen, err := mainscreen.New(settings.Keys, look.NewStyles(settings.Theme.Scheme()), mainscreen.Options{
		Location:   settings.Location,
		Remembered: settings.Remembered,
		Languages:  settings.Languages,
	})
	if err != nil {
		return mainscreen.Screen{}, fmt.Errorf("new main screen: %w", err)
	}

	return screen, nil
}

func missingDependencies(deps Deps) error {
	return errors.Join(
		domain.RequireDependency("lister", deps.Lister),
		domain.RequireDependency("treeLister", deps.TreeLister),
		domain.RequireDependency("tagLister", deps.TagLister),
		domain.RequireDependency("tagSnippetsLister", deps.TagSnippetsLister),
		domain.RequireDependency("copier", deps.Copier),
		domain.RequireDependency("trashCopier", deps.TrashCopier),
		domain.RequireDependency("trashLister", deps.TrashLister),
		domain.RequireDependency("trashRestorer", deps.TrashRestorer),
		domain.RequireDependency("creator", deps.Creator),
		domain.RequireDependency("capturer", deps.Capturer),
		domain.RequireDependency("updater", deps.Updater),
		domain.RequireDependency("duplicator", deps.Duplicator),
		domain.RequireDependency("deleter", deps.Deleter),
		domain.RequireDependency("mover", deps.Mover),
		domain.RequireDependency("searcher", deps.Searcher),
		domain.RequireDependency("folderCreator", deps.FolderCreator),
		domain.RequireDependency("folderRenamer", deps.FolderRenamer),
		domain.RequireDependency("folderDeleter", deps.FolderDeleter),
		domain.RequireDependency("folderDefaultLanguageSetter", deps.FolderDefaultLanguageSetter),
		domain.RequireDependency("folderMover", deps.FolderMover),
		domain.RequireDependency("tagCreator", deps.TagCreator),
		domain.RequireDependency("tagRenamer", deps.TagRenamer),
		domain.RequireDependency("tagDeletePreviewer", deps.TagDeletePreviewer),
		domain.RequireDependency("tagDeleter", deps.TagDeleter),
		domain.RequireDependency("sortOrderSaver", deps.SortOrderSaver),
		domain.RequireDependency("collapsedFoldersSaver", deps.CollapsedFoldersSaver),
		domain.RequireDependency("externalEditor", deps.ExternalEditor),
		domain.RequireDependency("editedContentHandler", deps.EditedContentHandler),
		domain.RequireDependency("clock", deps.Clock),
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
	case tea.MouseClickMsg, tea.MouseWheelMsg:
		return m.pointed(msg)
	case tea.BackgroundColorMsg:
		return m.restyledOn(msg)
	case tea.WindowSizeMsg, tea.PasteMsg,
		editoverlay.SaveFinished, searchpopup.HitsFound, mainscreen.SnippetsLoaded, mainscreen.TreeLoaded,
		mainscreen.TreeChanged, mainscreen.TrashLoaded, mainscreen.TagsLoaded,
		mainscreen.TagCreated, mainscreen.TagsChanged, mainscreen.TagDeletePreviewed:
		return m.overlaysUpdatedSharingArrivals(msg)
	case folderTreeChangedMsg, folderEditedMsg, tagCreatedMsg, tagsChangedMsg, snippetsChangedMsg:
		return m.reloadedAfter(msg)
	case operationFailedMsg, actionFailedMsg, listFailedMsg, treeFailedMsg, searchFailedMsg:
		return m.failedWith(msg)
	case copyFinishedMsg:
		return m.copyFinished(msg)
	case captureFinishedMsg:
		return m.captureFinished(msg)
	case externalEditFinishedMsg:
		return m.externalEditFinished(msg)
	}

	return m, nil
}

func (m Model) View() tea.View {
	view := tea.NewView(m.overlays.Render())
	view.AltScreen = true
	view.MouseMode = m.mouse.mode()

	return view
}

func (m Model) pointed(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		return m.clicked(msg)
	case tea.MouseWheelMsg:
		if wheel, ok := m.mouse.wheeled(msg); ok {
			return m.overlaysUpdated(wheel)
		}
	}

	return m, nil
}

func (m Model) clicked(msg tea.MouseClickMsg) (Model, tea.Cmd) {
	next := m

	var (
		click pointer.Clicked
		ok    bool
	)

	next.mouse, click, ok = m.mouse.clicked(msg)
	if !ok {
		return next, nil
	}

	return next.overlaysUpdated(click)
}

func (m Model) restyledOn(background tea.BackgroundColorMsg) (Model, tea.Cmd) {
	return m.overlaysUpdated(look.Restyled{Styles: look.NewStyles(m.theme.SchemeOn(background))})
}

func (m Model) failedWith(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case operationFailedMsg:
		return m.failed(msg.operation, msg.err)
	case actionFailedMsg:
		return m.actionFailed(msg)
	case listFailedMsg:
		return m.failed(operationList, msg.err)
	case treeFailedMsg:
		return m.failed(operationTree, msg.err)
	case searchFailedMsg:
		return m.failed(operationSearch, msg.err)
	default:
		return m, nil
	}
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

func (m Model) reloadedAfter(msg tea.Msg) (Model, tea.Cmd) {
	reload := m.reloadAfter(msg)

	carrier, ok := msg.(statusCarrier)
	if !ok || carrier.statusText() == "" {
		return m, reload
	}

	next, cmd := m.shown(carrier.statusText())

	return next, tea.Batch(cmd, reload)
}

func (m Model) reloadAfter(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case folderTreeChangedMsg:
		return tea.Batch(m.loadTreeSelecting(msg.selecting), m.loadTags())
	case folderEditedMsg:
		return m.loadTree()
	case tagCreatedMsg:
		return m.listTags(func(tags []browse.TagCount) tea.Msg { return mainscreen.TagCreated{Tags: tags, ID: msg.id} })
	case tagsChangedMsg:
		return m.listTags(func(tags []browse.TagCount) tea.Msg {
			return mainscreen.TagsChanged{Tags: tags, Selecting: msg.selecting}
		})
	case snippetsChangedMsg:
		return m.reloadSelecting(msg.selection, msg.selecting)
	default:
		return nil
	}
}

func (m Model) loadTags() tea.Cmd {
	return m.listTags(func(tags []browse.TagCount) tea.Msg { return mainscreen.TagsLoaded{Tags: tags} })
}

func (m Model) listTags(loaded func(tags []browse.TagCount) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		tags, err := m.tagLister.Run(m.ctx, browse.TagListInput{})
		if err != nil {
			return operationFailedMsg{operation: operationTags, err: err}
		}

		return loaded(tags)
	}
}

func (m Model) loadSnippets(selection browseselection.Selection, selecting domain.SnippetID) tea.Cmd {
	return m.loadSnippetsShowing(selection, selecting, domain.FragmentID{})
}

func (m Model) loadSnippetsShowing(
	selection browseselection.Selection, selecting domain.SnippetID, showing domain.FragmentID,
) tea.Cmd {
	order := m.sortOrder
	loaded := func(snippets []domain.Snippet, err error) tea.Msg {
		if err != nil {
			return listFailedMsg{err: err}
		}

		return mainscreen.SnippetsLoaded{
			Selection: selection, Snippets: snippets, Selecting: selecting, Showing: showing, Order: order,
		}
	}

	if selection.Trash() {
		return m.loadTrash()
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

// Listings run concurrently, so one asked for before a sort cycle can arrive after the one asked for since.
func (m Model) snippetsLoaded(loaded mainscreen.SnippetsLoaded) (Model, tea.Cmd) {
	if loaded.Order != m.sortOrder {
		return m, nil
	}

	return m.overlaysUpdated(loaded)
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

func (m Model) overlaysUpdatedSharingArrivals(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case mainscreen.SnippetsLoaded:
		return m.snippetsLoaded(msg)
	case mainscreen.TreeLoaded:
		return m.sharingArrival(msg, arrived.Tree{Tree: msg.Tree})
	case mainscreen.TreeChanged:
		return m.sharingArrival(msg, arrived.Tree{Tree: msg.Tree})
	case mainscreen.TagsLoaded:
		return m.sharingArrival(msg, arrived.Tags{Tags: msg.Tags})
	case mainscreen.TagCreated:
		return m.sharingArrival(msg, arrived.Tags{Tags: msg.Tags})
	case mainscreen.TagsChanged:
		return m.sharingArrival(msg, arrived.Tags{Tags: msg.Tags})
	}

	return m.overlaysUpdated(msg)
}

func (m Model) sharingArrival(msg, arrival tea.Msg) (Model, tea.Cmd) {
	screenUpdated, screenCmd := m.overlaysUpdated(msg)
	overlaysUpdated, overlaysCmd := screenUpdated.overlaysUpdated(arrival)

	return overlaysUpdated, tea.Batch(screenCmd, overlaysCmd)
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
	case outcome.SaveRequested, outcome.UpdateRequested, outcome.SearchTyped, outcome.CopyRequested,
		outcome.TrashCopyRequested, outcome.CaptureAsked, outcome.ExternalEditAsked, outcome.SnippetDuplicateRequested, outcome.SnippetDeleteRequested,
		outcome.SnippetMoveRequested, outcome.TrashRestoreRequested:
		return m, m.runSnippetAction(reported)
	case outcome.SnippetSaved, outcome.SnippetReloaded, outcome.SnippetRevealed, outcome.FolderSelected,
		outcome.TagSelected, outcome.TrashSelected, outcome.TagsChanged, outcome.SortCycleAsked, outcome.ReloadAsked:
		return m.relisted(reported)
	case outcome.SaveFailed, outcome.SortOrderRejected:
		return m.reportedFailure(reported)
	case outcome.NoticeShown:
		return m.shown(reported.Text)
	case outcome.FolderCreateRequested, outcome.FolderRenameRequested,
		outcome.FolderDeleteRequested, outcome.DefaultLanguageRequested,
		outcome.FolderMoveRequested:
		return m, m.runFolderAction(reported)
	case outcome.TagCreateRequested, outcome.TagRenameRequested,
		outcome.TagDeleteAsked, outcome.TagDeleteRequested:
		return m, m.runTagAction(reported)
	case outcome.CollapsedFoldersChanged:
		return m, m.saveCollapsedFolders(reported.IDs)
	case outcome.DiscardConfirmed, outcome.FragmentRemoveConfirmed, outcome.LanguagePicked, outcome.TagsEdited,
		outcome.ContentEdited:
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
		return m, m.reloadSelecting(reported.Selection, reported.ID)
	case outcome.SnippetReloaded:
		return m, m.reloadSelecting(reported.Selection, reported.ID)
	case outcome.ReloadAsked:
		return m, m.reloadSelecting(reported.Selection, reported.Selecting)
	case outcome.SnippetRevealed:
		return m, m.loadSnippetsShowing(browseselection.InFolder(reported.FolderID), reported.ID, reported.FragmentID)
	case outcome.FolderSelected, outcome.TagSelected, outcome.TrashSelected:
		return m, m.selectionChanged(reported)
	case outcome.TagsChanged:
		return m, m.loadSnippets(reported.Selection, reported.Selecting)
	case outcome.SortCycleAsked:
		return m.sortCycled(reported)
	default:
		return m, nil
	}
}

func (m Model) selectionChanged(reported outcome.Outcome) tea.Cmd {
	switch reported := reported.(type) {
	case outcome.FolderSelected:
		return m.loadSnippets(browseselection.InFolder(reported.ID), domain.SnippetID{})
	case outcome.TagSelected:
		return m.loadSnippets(browseselection.WithTag(reported.ID), domain.SnippetID{})
	case outcome.TrashSelected:
		return tea.Batch(m.loadTree(), m.loadTrash())
	default:
		return nil
	}
}

func (m Model) reloadSelecting(selection browseselection.Selection, id domain.SnippetID) tea.Cmd {
	return tea.Batch(m.loadTree(), m.loadTags(), m.loadSnippets(selection, id))
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
		return m.copySnippet(reported)
	case outcome.TrashCopyRequested:
		return m.copyFromTrash(reported)
	case outcome.CaptureAsked:
		return m.captureSnippet()
	case outcome.ExternalEditAsked:
		return m.editExternally(reported)
	default:
		return m.runSnippetListChange(reported)
	}
}

func (m Model) runTagAction(reported outcome.Outcome) tea.Cmd {
	switch reported := reported.(type) {
	case outcome.TagCreateRequested:
		return m.createTag(reported.Input)
	case outcome.TagRenameRequested:
		return m.renameTag(reported.Input)
	case outcome.TagDeleteAsked:
		return m.previewTagDelete(tag.PreviewDeleteInput{TagID: reported.ID})
	case outcome.TagDeleteRequested:
		return m.deleteTag(reported.Input)
	default:
		return nil
	}
}

func (m Model) createTag(in tag.CreateInput) tea.Cmd {
	return func() tea.Msg {
		created, err := m.tagCreator.Run(m.ctx, in)
		if err != nil {
			return operationFailedMsg{operation: operationCreateTag, err: err}
		}

		return tagCreatedMsg{id: created.ID()}
	}
}

func (m Model) renameTag(in tag.RenameInput) tea.Cmd {
	return func() tea.Msg {
		survivor, err := m.tagRenamer.Run(m.ctx, in)
		if err != nil {
			return operationFailedMsg{operation: operationRenameTag, err: err}
		}

		return tagsChangedMsg{selecting: survivor.ID()}
	}
}

func (m Model) previewTagDelete(in tag.PreviewDeleteInput) tea.Cmd {
	return func() tea.Msg {
		preview, err := m.tagDeletePreviewer.Run(m.ctx, in)
		if err != nil {
			return operationFailedMsg{operation: operationPreviewTagDelete, err: err}
		}

		return mainscreen.TagDeletePreviewed{Preview: preview}
	}
}

func (m Model) deleteTag(in tag.DeleteInput) tea.Cmd {
	return func() tea.Msg {
		err := m.tagDeleter.Run(m.ctx, in)
		if err != nil {
			return operationFailedMsg{operation: operationDeleteTag, err: err}
		}

		return tagsChangedMsg{selecting: domain.TagID{}}
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

func (m Model) copySnippet(requested outcome.CopyRequested) tea.Cmd {
	return m.copyThrough(m.copier, requested, goneElsewhere(snippetNoun, requested.Title))
}

func (m Model) copyThrough(copier SnippetCopier, requested outcome.CopyRequested, whenGone gone) tea.Cmd {
	input := snippet.CopyInput{SnippetID: requested.ID, FragmentID: requested.FragmentID}
	copied := reporting(
		actionFailure{operation: operationCopy, whenGone: whenGone},
		func(result snippet.CopyResult) tea.Msg { return copyFinishedMsg{result: result} },
	)

	return func() tea.Msg { return copied(copier.Run(m.ctx, input)) }
}

func (m Model) copyFinished(msg copyFinishedMsg) (Model, tea.Cmd) {
	text, err := deliveryText(msg.result.Delivery)
	if err != nil {
		return m.failed(operationCopy, err)
	}

	next, cmd := m.shown(text)

	return next, tea.Sequence(tea.Batch(cmd, terminalClipboard(msg.result)), m.afterCopy)
}

func (m Model) captureSnippet() tea.Cmd {
	return func() tea.Msg {
		content, err := m.capturer.Run(m.ctx, snippet.CaptureInput{})

		return captureFinishedMsg{content: content.String(), err: err}
	}
}

func (m Model) captureFinished(msg captureFinishedMsg) (Model, tea.Cmd) {
	if msg.err == nil {
		return m.overlaysUpdated(mainscreen.Captured{Content: msg.content})
	}

	if refusal, refused := captureRefusalText(msg.err); refused {
		return m.shown(refusal)
	}

	return m.failedShowing(operationCapture, msg.err, captureFailureText(msg.err))
}

func terminalClipboard(result snippet.CopyResult) tea.Cmd {
	if result.Delivery != domain.CopyDeliverySentToTerminal {
		return nil
	}

	return tea.SetClipboard(result.Content.String())
}

func (m Model) failed(operation string, err error) (Model, tea.Cmd) {
	return m.failedShowing(operation, err, failureText(err))
}

func (m Model) failedShowing(operation string, err error, text string) (Model, tea.Cmd) {
	if errors.Is(err, context.Canceled) {
		m.logger.DebugContext(m.ctx, "operation cancelled", slog.String(keyOperation, operation))

		return m, nil
	}

	m.logger.ErrorContext(m.ctx, "operation failed", slog.String(keyOperation, operation), slog.Any(keyError, err))

	return m.shown(text)
}

func (m Model) shown(text string) (Model, tea.Cmd) {
	return m.overlaysUpdated(mainscreen.StatusShown{Text: text})
}

func requirePointer[T any](name string, required *T) error {
	if required == nil {
		return fmt.Errorf("%s: %w", name, domain.ErrMissingDependency)
	}

	return nil
}
