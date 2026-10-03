package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type editSession struct {
	form        editForm
	confirmKeys confirmBindings
	styles      styleSet
	outer       size
	saving      bool
}

func newEditSession(keys editorBindings, confirmKeys confirmBindings, styles styleSet) (editSession, tea.Cmd) {
	form, cmd := newEditForm(keys)

	return editSession{
		form:        form,
		confirmKeys: confirmKeys,
		styles:      styles,
		outer:       size{width: 0, height: 0},
		saving:      false,
	}, cmd
}

func (s editSession) Update(msg tea.Msg) step {
	switch msg := msg.(type) {
	case tea.KeyPressMsg, tea.PasteMsg:
		return s.formUpdated(msg)
	case tea.WindowSizeMsg:
		return stay(s.resized(sizeOf(msg)))
	case snippetCreatedMsg:
		return s.saved(msg)
	}

	return stay(s)
}

func (s editSession) Received(received outcome.Outcome) step {
	switch received.(type) {
	case outcome.DiscardConfirmed:
		return closing()
	case outcome.QuitAsked:
		if s.form.changed() {
			return s.confirming(quitQuestion, outcome.QuitConfirmed{})
		}
	default:
	}

	return stay(s).Passing(received)
}

func (s editSession) View() string {
	return s.form.view(s.styles, s.outer)
}

func (s editSession) Hints() []key.Binding {
	return s.form.hints()
}

func (s editSession) formUpdated(msg tea.Msg) step {
	next := s

	var (
		request editRequest
		cmd     tea.Cmd
	)

	next.form, request, cmd = s.form.update(msg)

	return next.requested(request).Running(cmd)
}

func (s editSession) requested(request editRequest) step {
	switch request {
	case editSaves:
		return s.saveStarted()
	case editCancels:
		return s.cancelled()
	case editRefusesPaste:
		return stay(s).Passing(outcome.NoticeShown{Text: pasteHasTabsText})
	case editStays:
	}

	return stay(s)
}

func (s editSession) saveStarted() step {
	if s.saving {
		return stay(s)
	}

	next := s
	next.saving = true

	return stay(next).Passing(outcome.SaveRequested{Input: s.form.input()})
}

func (s editSession) cancelled() step {
	switch {
	case s.saving:
		return stay(s)
	case s.form.changed():
		return s.confirming(discardQuestion, outcome.DiscardConfirmed{})
	}

	return closing()
}

func (s editSession) saved(msg snippetCreatedMsg) step {
	if msg.err == nil {
		return closing().Passing(outcome.SnippetSaved{ID: msg.snippet.ID()})
	}

	next := s
	next.saving = false

	fieldErrors := domain.FieldErrors(msg.err)
	if len(fieldErrors) == 0 {
		return stay(next).Passing(outcome.SaveFailed{Err: msg.err})
	}

	next.form = s.form.withInvalid(fieldErrors)

	return stay(next).Passing(outcome.NoticeShown{Text: fieldErrorText(fieldErrors[0])})
}

func (s editSession) confirming(question string, onYes outcome.Outcome) step {
	return stay(s).Opening(newConfirmation(s.confirmKeys, s.styles, question, onYes))
}

func (s editSession) resized(screen size) editSession {
	s.outer = shareOf(screen, editOverlayPercent)
	s.form = s.form.resized(s.outer)

	return s
}
