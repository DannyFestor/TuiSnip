package tui

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type listFailedMsg struct {
	err error
}

type treeFailedMsg struct {
	err error
}

type copyFinishedMsg struct {
	result snippet.CopyResult
}

type captureFinishedMsg struct {
	content string
	err     error
}

type externalEditFinishedMsg struct {
	asked   outcome.ExternalEditAsked
	content string
	err     error
}

type searchFailedMsg struct {
	err error
}

type statusCarrier interface {
	statusText() string
}

type folderTreeChangedMsg struct {
	selecting domain.FolderID
	status    string
}

func (m folderTreeChangedMsg) statusText() string {
	return m.status
}

type folderEditedMsg struct{}

type tagCreatedMsg struct {
	id domain.TagID
}

type tagsChangedMsg struct {
	selecting domain.TagID
}

type snippetsChangedMsg struct {
	selection browseselection.Selection
	selecting domain.SnippetID
	status    string
}

func (m snippetsChangedMsg) statusText() string {
	return m.status
}

type operationFailedMsg struct {
	operation string
	err       error
}
