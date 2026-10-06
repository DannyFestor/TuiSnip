package folderpicker

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Offer struct {
	Title   string
	Tree    browse.Tree
	Current domain.FolderID
	Moving  domain.FolderID
	Picked  func(picked domain.FolderID) outcome.Outcome
}
