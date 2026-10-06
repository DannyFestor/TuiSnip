package tageditor

import (
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagchoice"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

// offered keeps a created name listed after it is toggled off, so the row
// under the cursor never disappears from under it.
type toggles struct {
	chosen  tagchoice.Chosen
	offered []value.TagName
}

func togglesFrom(chosen tagchoice.Chosen) toggles {
	return toggles{chosen: chosen, offered: chosen.Created()}
}

func (t toggles) withCreated(name value.TagName) toggles {
	t.offered = append(slices.Clone(t.offered), name)
	t.chosen = t.chosen.ToggledNew(name)

	return t
}
