package editoverlay

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type saveTarget interface {
	savingAs(saving outcome.Overlay, values entered) outcome.Step
	closedAfterSave(saved domain.Snippet) outcome.Step
	reloaded() (outcome.SnippetReloaded, bool)
	newFragmentLanguage() value.Language
}

type newSnippet struct {
	destination Destination
}

func (n newSnippet) newFragmentLanguage() value.Language {
	return n.destination.Language
}

func (n newSnippet) savingAs(saving outcome.Overlay, values entered) outcome.Step {
	return outcome.Stay(saving).Passing(outcome.SaveRequested{Input: snippet.CreateInput{
		Title:       values.title,
		Description: values.description,
		Fragments:   values.fragmentInputs(),
		FolderID:    n.destination.folderID(),
		Tags:        values.tags.Stored(),
		NewTags:     values.newTagNames(),
	}})
}

func (n newSnippet) closedAfterSave(saved domain.Snippet) outcome.Step {
	return outcome.Close().Passing(outcome.SnippetSaved{ID: saved.ID(), Selection: n.destination.Selection})
}

func (newSnippet) reloaded() (outcome.SnippetReloaded, bool) {
	return outcome.SnippetReloaded{ID: domain.SnippetID{}, Selection: browseselection.Selection{}}, false
}

type storedSnippet struct {
	id              domain.SnippetID
	selection       browseselection.Selection
	loadedUpdatedAt time.Time
	defaultLanguage value.Language
}

func (s storedSnippet) newFragmentLanguage() value.Language {
	return s.defaultLanguage
}

func (s storedSnippet) savingAs(saving outcome.Overlay, values entered) outcome.Step {
	return outcome.Stay(saving).Passing(outcome.UpdateRequested{Input: snippet.UpdateInput{
		SnippetID:       s.id,
		LoadedUpdatedAt: s.loadedUpdatedAt,
		Title:           values.title,
		Description:     values.description,
		Fragments:       values.fragmentInputs(),
		Tags:            values.tags.Stored(),
		NewTags:         values.newTagNames(),
	}})
}

func (s storedSnippet) closedAfterSave(domain.Snippet) outcome.Step {
	reload, _ := s.reloaded()

	return outcome.Close().Passing(reload)
}

func (s storedSnippet) reloaded() (outcome.SnippetReloaded, bool) {
	return outcome.SnippetReloaded{ID: s.id, Selection: s.selection}, true
}
