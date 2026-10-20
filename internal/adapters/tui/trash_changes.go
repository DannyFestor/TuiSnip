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
	operationListTrash = "list trash"
	operationRestore   = "restore from trash"
)

type trashActions struct {
	copier   SnippetCopier
	lister   TrashLister
	restorer TrashRestorer
}

func trashActionsOf(deps Deps) trashActions {
	return trashActions{copier: deps.TrashCopier, lister: deps.TrashLister, restorer: deps.TrashRestorer}
}

func (m Model) loadTrash() tea.Cmd {
	return func() tea.Msg {
		items, err := m.trash.lister.Run(m.ctx, trash.ListInput{})
		if err != nil {
			return operationFailedMsg{operation: operationListTrash, err: err}
		}

		return mainscreen.TrashLoaded{Items: items}
	}
}

func (m Model) restoreFromTrash(requested outcome.TrashRestoreRequested) tea.Cmd {
	return func() tea.Msg {
		restoration, err := m.trash.restorer.Run(m.ctx, requested.Input)
		if err != nil {
			return itemActionFailedMsg{operation: operationRestore, err: err, gone: goneFromTrash(requested.Name)}
		}

		return snippetsChangedMsg{
			selection: browseselection.InTrash(),
			selecting: domain.SnippetID{},
			status:    restoredText(requested, restoration),
		}
	}
}
