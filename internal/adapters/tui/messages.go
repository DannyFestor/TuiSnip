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

type searchFailedMsg struct {
	err error
}

type folderTreeChangedMsg struct {
	selecting domain.FolderID
}

type folderRenamedMsg struct{}

type folderChangeFailedMsg struct {
	operation string
	err       error
}
