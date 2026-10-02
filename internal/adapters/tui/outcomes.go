package tui

import (
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type saveRequested struct {
	input snippet.CreateInput
}

type snippetSaved struct {
	id domain.SnippetID
}

type saveFailed struct {
	err error
}

type noticeShown struct {
	text string
}

type searchTyped struct {
	text string
}

type snippetRevealed struct {
	id domain.SnippetID
}

type copyRequested struct {
	id domain.SnippetID
}

type discardConfirmed struct{}

type quitAsked struct{}

type quitConfirmed struct{}
