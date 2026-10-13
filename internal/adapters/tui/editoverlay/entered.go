package editoverlay

import (
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagchoice"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
)

type entered struct {
	title       string
	description string
	tags        tagchoice.Chosen
	fragments   []fragmentDraft
}

func (e entered) equal(other entered) bool {
	return e.title == other.title &&
		e.description == other.description &&
		e.tags.Equal(other.tags) &&
		slices.Equal(e.fragments, other.fragments)
}

func (e entered) fragmentInputs() []snippet.FragmentInput {
	inputs := make([]snippet.FragmentInput, 0, len(e.fragments))
	for _, draft := range e.fragments {
		inputs = append(inputs, draft.input())
	}

	return inputs
}

func (e entered) newTagNames() []string {
	created := e.tags.Created()
	if len(created) == 0 {
		return nil
	}

	names := make([]string, 0, len(created))

	for _, name := range created {
		names = append(names, name.String())
	}

	return names
}
