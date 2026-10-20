package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
)

type actionFailure struct {
	operation string
	whenGone  gone
}

type actionFailedMsg struct {
	actionFailure

	err error
}

func reporter[T any](failure actionFailure, succeeded func(T) tea.Msg) func(T, error) tea.Msg {
	return func(result T, err error) tea.Msg {
		if err != nil {
			return actionFailedMsg{actionFailure: failure, err: err}
		}

		return succeeded(result)
	}
}

func withoutResult(err error) (struct{}, error) {
	return struct{}{}, err
}

func (m Model) actionFailed(msg actionFailedMsg) (Model, tea.Cmd) {
	text, isGone := msg.whenGone.text(msg.err)
	if !isGone {
		return m.failed(msg.operation, msg.err)
	}

	m.logger.InfoContext(m.ctx, "gone elsewhere",
		slog.String(keyOperation, msg.operation), slog.Any(keyError, msg.err))

	return m.overlaysUpdated(mainscreen.GoneElsewhereShown{Text: text})
}
