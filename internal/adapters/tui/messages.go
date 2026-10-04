package tui

import (
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
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
