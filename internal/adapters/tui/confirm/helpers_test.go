package confirm_test

import (
	"testing"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/confirm"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const question = "Discard the unsaved changes?"

func screenSize() look.Size {
	return look.Size{Width: 80, Height: 24}
}

func asking(t *testing.T) *overlaytest.Driver {
	t.Helper()

	return askingWith(t, testsettings.Default(t).Keys, question)
}

func askingWith(t *testing.T, keys binding.Keys, asked string) *overlaytest.Driver {
	t.Helper()

	opened := confirm.New(keys, look.NewStyles(), asked, outcome.DiscardConfirmed{})

	return overlaytest.Open(t, screenSize(), opened)
}
