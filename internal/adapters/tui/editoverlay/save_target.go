package editoverlay

import (
	"errors"
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/confirm"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	goneTitle           = "Gone elsewhere"
	snippetGoneQuestion = "This Snippet was moved to the Trash or deleted in another TuiSnip. " +
		"Save your changes as a new Snippet at the Root?"
	folderGoneQuestion = "Folder %q was moved to the Trash or deleted in another TuiSnip. " +
		"Save the Snippet at the Root instead?"
	savedAtRootFormat = "Saved %q as a new Snippet at the Root"
)

type saveTarget interface {
	savingAs(saving outcome.Overlay, values entered) outcome.Step
	closedAfterSave(saved domain.Snippet) outcome.Step
	reloaded() (outcome.SnippetReloaded, bool)
	goneQuestion(err error) (confirm.Question, bool)
	newFragmentLanguage() value.Language
}

type newSnippet struct {
	destination Destination
}

func (n newSnippet) newFragmentLanguage() value.Language {
	return n.destination.Language
}

func (n newSnippet) savingAs(saving outcome.Overlay, values entered) outcome.Step {
	return outcome.Stay(saving).Passing(outcome.SaveRequested{Input: values.createInput(n.destination.folderID())})
}

func (n newSnippet) closedAfterSave(saved domain.Snippet) outcome.Step {
	return outcome.Close().Passing(outcome.SnippetSaved{ID: saved.ID(), Selection: n.destination.Selection})
}

func (newSnippet) reloaded() (outcome.SnippetReloaded, bool) {
	return outcome.SnippetReloaded{ID: domain.SnippetID{}, Selection: browseselection.Selection{}}, false
}

func (n newSnippet) goneQuestion(err error) (confirm.Question, bool) {
	if !errors.Is(err, domain.ErrFolderGone) {
		return confirm.Question{}, false
	}

	return confirm.Question{Title: goneTitle, Text: fmt.Sprintf(folderGoneQuestion, n.destination.FolderName)}, true
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

func (storedSnippet) goneQuestion(err error) (confirm.Question, bool) {
	if !errors.Is(err, domain.ErrNotFound) {
		return confirm.Question{}, false
	}

	return confirm.Question{Title: goneTitle, Text: snippetGoneQuestion}, true
}

type snippetAtRoot struct{}

func (snippetAtRoot) newFragmentLanguage() value.Language {
	return value.PlainText()
}

func (snippetAtRoot) savingAs(saving outcome.Overlay, values entered) outcome.Step {
	return outcome.Stay(saving).Passing(outcome.SaveRequested{Input: values.createInput(domain.FolderID{})})
}

func (snippetAtRoot) closedAfterSave(saved domain.Snippet) outcome.Step {
	return outcome.Close().Passing(
		outcome.SnippetSaved{ID: saved.ID(), Selection: browseselection.InFolder(domain.FolderID{})},
		outcome.NoticeShown{Text: fmt.Sprintf(savedAtRootFormat, saved.Title().String())},
	)
}

func (snippetAtRoot) reloaded() (outcome.SnippetReloaded, bool) {
	return outcome.SnippetReloaded{ID: domain.SnippetID{}, Selection: browseselection.Selection{}}, false
}

func (snippetAtRoot) goneQuestion(error) (confirm.Question, bool) {
	return confirm.Question{}, false
}
