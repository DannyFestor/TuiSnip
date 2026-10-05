package helpoverlay_test

import (
	"testing"

	"charm.land/bubbles/v2/key"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/helpoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func wide() look.Size {
	return look.Size{Width: 120, Height: 40}
}

func helping(t *testing.T) *overlaytest.Driver {
	t.Helper()

	return helpingWith(t, testsettings.Default(t).Keys, wide())
}

func helpingWith(t *testing.T, keys binding.Keys, screen look.Size) *overlaytest.Driver {
	t.Helper()

	return overlaytest.Open(t, screen, helpoverlay.New(keys, look.NewStyles(look.SchemeDark), listedFor(keys)))
}

func listedFor(keys binding.Keys) [][]key.Binding {
	return append(keys.For(binding.ScopeGlobal).FullHelp(), keys.For(binding.ScopeFolders).FullHelp()...)
}
