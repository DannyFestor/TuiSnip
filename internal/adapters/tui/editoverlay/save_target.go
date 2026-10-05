package editoverlay

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type saveTarget interface {
	savingAs(saving outcome.Overlay, values entered) outcome.Step
	closedAfterSave(saved domain.Snippet) outcome.Step
	reloaded() (outcome.SnippetReloaded, bool)
}

type newSnippet struct{}

func (newSnippet) savingAs(saving outcome.Overlay, values entered) outcome.Step {
	return outcome.Stay(saving).Passing(outcome.SaveRequested{Input: snippet.CreateInput{
		Title:       values.title,
		Description: values.description,
		Content:     values.content,
	}})
}

func (newSnippet) closedAfterSave(saved domain.Snippet) outcome.Step {
	return outcome.Close().Passing(outcome.SnippetSaved{ID: saved.ID(), FolderID: saved.FolderID()})
}

func (newSnippet) reloaded() (outcome.SnippetReloaded, bool) {
	return outcome.SnippetReloaded{ID: domain.SnippetID{}, Selection: browseselection.Selection{}}, false
}

type storedSnippet struct {
	id              domain.SnippetID
	selection       browseselection.Selection
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

func (s storedSnippet) closedAfterSave(domain.Snippet) outcome.Step {
	reload, _ := s.reloaded()

	return outcome.Close().Passing(reload)
}

func (s storedSnippet) reloaded() (outcome.SnippetReloaded, bool) {
	return outcome.SnippetReloaded{ID: s.id, Selection: s.selection}, true
}
