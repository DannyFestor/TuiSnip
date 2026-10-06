package editoverlay

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagchoice"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type entered struct {
	title       string
	description string
	tags        tagchoice.Chosen
	language    value.Language
	content     string
}

func (e entered) equal(other entered) bool {
	return e.title == other.title &&
		e.description == other.description &&
		e.tags.Equal(other.tags) &&
		e.language == other.language &&
		e.content == other.content
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
