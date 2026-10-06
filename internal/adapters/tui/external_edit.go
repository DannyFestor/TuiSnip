package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
)

func (m Model) editExternally(asked outcome.ExternalEditAsked) tea.Cmd {
	run, err := m.externalEditor.Open(m.ctx, editedFileName(asked.Language), asked.Content)
	if err != nil {
		return func() tea.Msg { return externalEditFinishedMsg{asked: asked, content: "", err: err} }
	}

	session := newEditorSession(run)

	return tea.Exec(session, func(err error) tea.Msg {
		return externalEditFinishedMsg{asked: asked, content: session.edited, err: err}
	})
}

func (m Model) externalEditFinished(msg externalEditFinishedMsg) (Model, tea.Cmd) {
	switch {
	case msg.err != nil:
		return m.externalEditFailed(msg.err)
	case msg.content == msg.asked.Content:
		return m, nil
	}

	edited := outcome.ContentEdited{Asked: msg.asked, Content: msg.content}

	return m.settled(m.editedContentHandler.Handle(edited, m.overlays))
}

func (m Model) externalEditFailed(err error) (Model, tea.Cmd) {
	if refusal, refused := externalEditRefusalText(err); refused {
		return m.shown(refusal)
	}

	return m.failedShowing(operationExternalEdit, err, externalEditFailureText(err))
}
