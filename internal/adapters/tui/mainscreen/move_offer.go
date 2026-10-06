package mainscreen

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpicker"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	moveTitlePrefix = "Move "
	moveTitleSuffix = " to"
)

func folderMoveOffer(moving domain.Folder, tree browse.Tree) folderpicker.Offer {
	return folderpicker.Offer{
		Title:   moveTitle(moving.Name().String()),
		Tree:    tree,
		Current: moving.ParentID(),
		Moving:  moving.ID(),
		Picked: func(picked domain.FolderID) outcome.Outcome {
			return outcome.FolderMoveRequested{Input: folder.MoveInput{FolderID: moving.ID(), ParentID: picked}}
		},
	}
}

func snippetMoveOffer(moving domain.Snippet, tree browse.Tree) folderpicker.Offer {
	return folderpicker.Offer{
		Title:   moveTitle(moving.Title().String()),
		Tree:    tree,
		Current: moving.FolderID(),
		Moving:  domain.FolderID{},
		Picked: func(picked domain.FolderID) outcome.Outcome {
			return outcome.SnippetMoveRequested{Input: snippet.MoveInput{SnippetID: moving.ID(), FolderID: picked}}
		},
	}
}

func moveTitle(name string) string {
	return moveTitlePrefix + name + moveTitleSuffix
}
