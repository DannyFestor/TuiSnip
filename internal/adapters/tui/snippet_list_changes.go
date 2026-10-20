package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func (m Model) runSnippetListChange(reported outcome.Outcome) tea.Cmd {
	switch reported := reported.(type) {
	case outcome.SnippetDuplicateRequested:
		return m.duplicateSnippet(reported)
	case outcome.SnippetDeleteRequested:
		return m.deleteSnippet(reported)
	case outcome.SnippetMoveRequested:
		return m.moveSnippet(reported)
	default:
		return m.runTrashAction(reported)
	}
}

func (m Model) duplicateSnippet(requested outcome.SnippetDuplicateRequested) tea.Cmd {
	report := reporter(
		actionFailure{operation: operationDuplicate, whenGone: goneElsewhere(snippetNoun, requested.Title)},
		func(duplicate domain.Snippet) tea.Msg {
			return snippetsChangedMsg{selection: requested.Selection, selecting: duplicate.ID(), status: ""}
		},
	)

	return func() tea.Msg { return report(m.duplicator.Run(m.ctx, requested.Input)) }
}

func (m Model) deleteSnippet(requested outcome.SnippetDeleteRequested) tea.Cmd {
	report := reporter(
		actionFailure{operation: operationDeleteSnippet, whenGone: goneElsewhere(snippetNoun, requested.Title)},
		func(struct{}) tea.Msg {
			return snippetsChangedMsg{
				selection: requested.Selection,
				selecting: requested.Selecting,
				status:    movedToTrashText(snippetNoun, requested.Title),
			}
		},
	)

	return func() tea.Msg { return report(withoutResult(m.deleter.Run(m.ctx, requested.Input))) }
}

func (m Model) moveSnippet(requested outcome.SnippetMoveRequested) tea.Cmd {
	report := reporter(
		actionFailure{
			operation: operationMoveSnippet,
			whenGone:  goneElsewhere(snippetNoun, requested.Title).placingInto(requested.TargetName),
		},
		func(domain.Snippet) tea.Msg {
			return snippetsChangedMsg{selection: requested.Selection, selecting: requested.Selecting, status: ""}
		},
	)

	return func() tea.Msg { return report(m.mover.Run(m.ctx, requested.Input)) }
}
