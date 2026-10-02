package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type editSession struct {
	open   bool
	form   editForm
	saving bool
	notice string
}

func newEditSession(keys editorBindings) (editSession, tea.Cmd) {
	form, cmd := newEditForm(keys)

	return editSession{open: true, form: form, saving: false, notice: ""}, cmd
}

func noEditSession() editSession {
	return editSession{}
}

func (s editSession) update(msg tea.Msg) (editSession, editRequest, tea.Cmd) {
	form, request, cmd := s.form.update(msg)
	s.form = form

	switch request {
	case editSaves:
		next, saveRequest := s.saveRequested()

		return next, saveRequest, cmd
	case editCancels:
		return s, s.cancelRequest(), cmd
	case editStays, editAsksDiscard, editRefusesPaste:
	}

	return s, request, cmd
}

func (s editSession) saved(err error) (editSession, saveOutcome) {
	s.saving = false
	if err == nil {
		return s, saveSucceeded
	}

	fieldErrors := domain.FieldErrors(err)
	if len(fieldErrors) == 0 {
		return s, saveFailed
	}

	s.form = s.form.withInvalid(fieldErrors)
	s.notice = fieldErrorText(fieldErrors[0])

	return s, saveRejected
}

func (s editSession) hasUnsavedChanges() bool {
	return s.open && s.form.changed()
}

func (s editSession) input() snippet.CreateInput {
	return s.form.input()
}

func (s editSession) resized(outer size) editSession {
	s.form = s.form.resized(outer)

	return s
}

func (s editSession) hints() []key.Binding {
	return s.form.hints()
}

func (s editSession) view(styles styleSet, outer size) string {
	return s.form.view(styles, outer)
}

func (s editSession) saveRequested() (editSession, editRequest) {
	if s.saving {
		return s, editStays
	}

	s.saving = true

	return s, editSaves
}

func (s editSession) cancelRequest() editRequest {
	switch {
	case s.saving:
		return editStays
	case s.form.changed():
		return editAsksDiscard
	}

	return editCancels
}
