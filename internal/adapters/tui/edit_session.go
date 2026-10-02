package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/overlay"
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

func (s editSession) Update(msg tea.Msg) overlay.Step {
	switch msg := msg.(type) {
	case tea.KeyPressMsg, tea.PasteMsg:
		return s.formUpdated(msg)
	case tea.WindowSizeMsg:
		return overlay.Stay(s.resized(size{width: msg.Width, height: msg.Height}))
	case snippetCreatedMsg:
		return s.saved(msg)
	}

	return overlay.Stay(s)
}

func (s editSession) Received(outcome overlay.Outcome) overlay.Step {
	switch outcome.(type) {
	case discardConfirmed:
		return overlay.Close()
	case quitAsked:
		if s.form.changed() {
			return s.confirming(quitQuestion, quitConfirmed{})
		}
	}

	return overlay.Stay(s).Passing(outcome)
}

func (s editSession) View() string {
	return s.form.view(s.styles, s.outer)
}

func (s editSession) Hints() []key.Binding {
	return s.form.hints()
}

func (s editSession) formUpdated(msg tea.Msg) overlay.Step {
	next := s

	var (
		request editRequest
		cmd     tea.Cmd
	)

	next.form, request, cmd = s.form.update(msg)

	return next.requested(request).Running(cmd)
}

func (s editSession) requested(request editRequest) overlay.Step {
	switch request {
	case editSaves:
		return s.saveStarted()
	case editCancels:
		return s.cancelled()
	case editRefusesPaste:
		return overlay.Stay(s).Passing(noticeShown{text: pasteHasTabsText})
	case editStays:
	}

	return overlay.Stay(s)
}

func (s editSession) saveStarted() overlay.Step {
	if s.saving {
		return overlay.Stay(s)
	}

	next := s
	next.saving = true

	return overlay.Stay(next).Passing(saveRequested{input: s.form.input()})
}

func (s editSession) cancelled() overlay.Step {
	switch {
	case s.saving:
		return overlay.Stay(s)
	case s.form.changed():
		return s.confirming(discardQuestion, discardConfirmed{})
	}

	return overlay.Close()
}

func (s editSession) saved(msg snippetCreatedMsg) overlay.Step {
	if msg.err == nil {
		return overlay.Close().Passing(snippetSaved{id: msg.snippet.ID()})
	}

	next := s
	next.saving = false

	fieldErrors := domain.FieldErrors(msg.err)
	if len(fieldErrors) == 0 {
		return overlay.Stay(next).Passing(saveFailed{err: msg.err})
	}

	next.form = s.form.withInvalid(fieldErrors)

	return overlay.Stay(next).Passing(noticeShown{text: fieldErrorText(fieldErrors[0])})
}

func (s editSession) confirming(question string, onYes overlay.Outcome) overlay.Step {
	return overlay.Stay(s).Opening(newConfirmation(s.confirmKeys, s.styles, question, onYes))
}

func (s editSession) resized(screen size) editSession {
	s.outer = shareOf(screen, editOverlayPercent)
	s.form = s.form.resized(s.outer)

	return s
}
