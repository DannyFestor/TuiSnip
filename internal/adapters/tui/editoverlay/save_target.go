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
	staleTitle     = "Changed elsewhere"
	reloadQuestion = "This Snippet changed in another TuiSnip. Reload it and discard your changes?"

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
	confirmationAfter(refusal error) (confirmation, bool)
	newFragmentLanguage() value.Language
}

type confirmation struct {
	question confirm.Question
	onYes    outcome.Outcome
}

func savingAtRoot(question string) confirmation {
	return confirmation{
		question: confirm.Question{Title: goneTitle, Text: question},
		onYes:    outcome.SaveAtRootConfirmed{},
	}
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

func (n newSnippet) confirmationAfter(refusal error) (confirmation, bool) {
	if !errors.Is(refusal, domain.ErrFolderGone) {
		return confirmation{}, false
	}

	return savingAtRoot(fmt.Sprintf(folderGoneQuestion, n.destination.FolderName)), true
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
	return outcome.Close().Passing(s.reloaded())
}

func (s storedSnippet) confirmationAfter(refusal error) (confirmation, bool) {
	switch {
	case errors.Is(refusal, domain.ErrConflict):
		return confirmation{
			question: confirm.Question{Title: staleTitle, Text: reloadQuestion},
			onYes:    s.reloaded(),
		}, true
	case errors.Is(refusal, domain.ErrNotFound):
		return savingAtRoot(snippetGoneQuestion), true
	}

	return confirmation{}, false
}

func (s storedSnippet) reloaded() outcome.SnippetReloaded {
	return outcome.SnippetReloaded{ID: s.id, Selection: s.selection}
}

type snippetAtRoot struct {
	newSnippet
}

func atRoot() snippetAtRoot {
	return snippetAtRoot{destination: Destination{
		Selection:  browseselection.InFolder(domain.FolderID{}),
		FolderName: "",
		Tags:       nil,
		Language:   value.PlainText(),
	}}
}

func (r snippetAtRoot) closedAfterSave(saved domain.Snippet) outcome.Step {
	return r.newSnippet.closedAfterSave(saved).
		Passing(outcome.NoticeShown{Text: fmt.Sprintf(savedAtRootFormat, saved.Title().String())})
}
