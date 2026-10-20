package tui

import tea "charm.land/bubbletea/v2"

type actionFailure struct {
	operation string
	whenGone  gone
}

type actionFailedMsg struct {
	actionFailure

	err error
}

func reporting[T any](failure actionFailure, succeeded func(T) tea.Msg) func(T, error) tea.Msg {
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
