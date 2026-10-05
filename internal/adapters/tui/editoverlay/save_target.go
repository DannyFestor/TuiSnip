package editoverlay

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type saveTarget interface {
	savingAs(saving outcome.Overlay, values entered) outcome.Step
}

type newSnippet struct{}

func (newSnippet) savingAs(saving outcome.Overlay, values entered) outcome.Step {
	return outcome.Stay(saving).Passing(outcome.SaveRequested{Input: snippet.CreateInput{
		Title:       values.title,
		Description: values.description,
		Content:     values.content,
	}})
}

type storedSnippet struct {
	id              domain.SnippetID
	loadedUpdatedAt time.Time
}

func (s storedSnippet) savingAs(saving outcome.Overlay, values entered) outcome.Step {
	return outcome.Stay(saving).Passing(outcome.UpdateRequested{Input: snippet.UpdateInput{
		SnippetID:       s.id,
		LoadedUpdatedAt: s.loadedUpdatedAt,
		Title:           values.title,
		Description:     values.description,
		Content:         values.content,
	}})
}
