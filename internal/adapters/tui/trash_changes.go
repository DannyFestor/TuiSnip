package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	operationListTrash         = "list trash"
	operationRestore           = "restore from trash"
	operationDeletePermanently = "delete permanently"
	operationEmptyTrash        = "empty trash"
)

type trashActions struct {
	copier   SnippetCopier
	lister   TrashLister
	restorer TrashRestorer
	deleter  TrashDeleter
	emptier  TrashEmptier
}

func trashActionsOf(deps Deps) trashActions {
	return trashActions{
		copier:   deps.TrashCopier,
		lister:   deps.TrashLister,
		restorer: deps.TrashRestorer,
		deleter:  deps.TrashDeleter,
		emptier:  deps.TrashEmptier,
	}
}

func (m Model) runTrashAction(reported outcome.Outcome) tea.Cmd {
	switch reported := reported.(type) {
	case outcome.TrashRestoreRequested:
		return m.restoreFromTrash(reported)
	case outcome.TrashDeleteRequested:
		return m.deleteFromTrash(reported)
	case outcome.TrashEmptyRequested:
		return m.emptyTrash(reported)
	default:
		return nil
	}
}

func (m Model) loadTrash() tea.Cmd {
	return func() tea.Msg {
		trashItems, err := m.trash.lister.Run(m.ctx, trash.ListInput{})
		if err != nil {
			return operationFailedMsg{operation: operationListTrash, err: err}
		}

		return mainscreen.TrashLoaded{Items: trashItems}
	}
}

func (m Model) copyFromTrash(requested outcome.TrashCopyRequested) tea.Cmd {
	return m.copyThrough(m.trash.copier, outcome.CopyRequested(requested), goneFromTrash(requested.Title))
}

func (m Model) restoreFromTrash(requested outcome.TrashRestoreRequested) tea.Cmd {
	report := reporter(
		actionFailure{operation: operationRestore, whenGone: goneFromTrash(requested.Name)},
		func(restoration domain.Restoration) tea.Msg {
			return trashChanged(restoredText(requested, restoration))
		},
	)

	return func() tea.Msg { return report(m.trash.restorer.Run(m.ctx, requested.Input)) }
}

func (m Model) deleteFromTrash(requested outcome.TrashDeleteRequested) tea.Cmd {
	report := reporter(
		actionFailure{operation: operationDeletePermanently, whenGone: goneFromTrash(requested.Name)},
		func(struct{}) tea.Msg { return trashChanged("") },
	)

	return func() tea.Msg { return report(withoutResult(m.trash.deleter.Run(m.ctx, requested.Input))) }
}

func (m Model) emptyTrash(requested outcome.TrashEmptyRequested) tea.Cmd {
	report := reporter(
		actionFailure{operation: operationEmptyTrash, whenGone: gone{notFound: "", folderGone: ""}},
		func(emptied domain.TrashEmptied) tea.Msg { return trashChanged(emptiedText(emptied)) },
	)

	return func() tea.Msg { return report(m.trash.emptier.Run(m.ctx, requested.Input)) }
}

func trashChanged(status string) snippetsChangedMsg {
	return snippetsChangedMsg{selection: browseselection.InTrash(), selecting: domain.SnippetID{}, status: status}
}
