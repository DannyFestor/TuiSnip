package tui

import (
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
	err    error
}

type captureFinishedMsg struct {
	content string
	err     error
}

type searchFailedMsg struct {
	err error
}

type folderTreeChangedMsg struct {
	selecting domain.FolderID
}

type folderEditedMsg struct{}

type tagCreatedMsg struct {
	id domain.TagID
}

type tagsChangedMsg struct {
	selecting domain.TagID
}

type operationFailedMsg struct {
	operation string
	err       error
}
