package folderpicker_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpicker"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	pickerTitle = "Move to"
	noMatches   = "No Folder matches."
)

func picking(t *testing.T, sample foldertree.Sample, current, moving domain.FolderID) *overlaytest.Driver {
	t.Helper()

	return pickingIn(t, sample.Tree, current, moving)
}

func pickingIn(t *testing.T, tree browse.Tree, current, moving domain.FolderID) *overlaytest.Driver {
	t.Helper()

	return pickingStyled(t, look.NewStyles(look.SchemeDark), offerOf(tree, current, moving))
}

func pickingStyled(t *testing.T, styles look.Styles, offer folderpicker.Offer) *overlaytest.Driver {
	t.Helper()

	opened, _ := folderpicker.New(testsettings.Default(t).Keys, styles, offer)

	return overlaytest.Open(t, look.Size{Width: 120, Height: 40}, opened)
}

func offerOf(tree browse.Tree, current, moving domain.FolderID) folderpicker.Offer {
	return folderpicker.Offer{
		Title:   pickerTitle,
		Tree:    tree,
		Current: current,
		Moving:  moving,
		Picked:  func(picked domain.FolderID) outcome.Outcome { return pickedOutcome(picked) },
	}
}

func pickedOutcome(picked domain.FolderID) outcome.FolderSelected {
	return outcome.FolderSelected{ID: picked}
}

func down() tea.KeyPressMsg {
	return keypress.Special(tea.KeyDown)
}

func enter() tea.KeyPressMsg {
	return keypress.Special(tea.KeyEnter)
}
