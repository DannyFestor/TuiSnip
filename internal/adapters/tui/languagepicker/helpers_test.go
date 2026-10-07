package languagepicker_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/languagepicker"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	pickerTitle = "Pick a Language"
	noMatches   = "No Language matches."
)

func picking(t *testing.T, curated []value.Language, current value.Language) *overlaytest.Driver {
	t.Helper()

	return pickingWith(t, testsettings.Default(t).Keys, curated, current)
}

func pickingWith(
	t *testing.T,
	keys binding.Keys,
	curated []value.Language,
	current value.Language,
) *overlaytest.Driver {
	t.Helper()

	return pickingStyled(t, keys, look.NewStyles(look.SchemeDark), offerOf(curated, current))
}

func pickingStyled(
	t *testing.T,
	keys binding.Keys,
	styles look.Styles,
	offer languagepicker.Offer,
) *overlaytest.Driver {
	t.Helper()

	opened, _ := languagepicker.New(keys, styles, offer)

	return overlaytest.Open(t, look.Size{Width: 120, Height: 40}, opened)
}

func offerOf(curated []value.Language, current value.Language) languagepicker.Offer {
	return languagepicker.Offer{
		Title:   pickerTitle,
		Curated: curated,
		Current: current,
		Picked:  func(picked value.Language) outcome.Outcome { return outcome.LanguagePicked{Language: picked} },
	}
}

func language(t *testing.T, name string) value.Language {
	t.Helper()

	parsed, err := value.NewLanguage(name)
	require.NoError(t, err)

	return parsed
}

func languages(t *testing.T, names ...string) []value.Language {
	t.Helper()

	parsed := make([]value.Language, 0, len(names))
	for _, name := range names {
		parsed = append(parsed, language(t, name))
	}

	return parsed
}

func pickedOutcome(t *testing.T, name string) outcome.LanguagePicked {
	t.Helper()

	return outcome.LanguagePicked{Language: language(t, name)}
}

func showAll() tea.KeyPressMsg {
	return keypress.Ctrl('a')
}
