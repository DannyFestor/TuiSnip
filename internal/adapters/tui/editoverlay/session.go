package editoverlay

import (
	"errors"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/confirm"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	overlayPercent  = 90
	unsavedTitle    = "Unsaved changes"
	discardQuestion = "Discard the unsaved changes?"
	quitQuestion    = "Quit and discard the unsaved changes?"
	staleTitle      = "Changed elsewhere"
	reloadQuestion  = "This Snippet changed in another TuiSnip. Reload it and discard your changes?"
)

type Session struct {
	keys   binding.Keys
	form   form
	target saveTarget
	styles look.Styles
	outer  look.Size
	saving bool
}

func New(keys binding.Keys, styles look.Styles) (Session, tea.Cmd) {
	blank, cmd := newForm(
		formKeysOf(keys),
		entered{title: "", description: "", content: ""},
		editableContent(),
	)

	return newSession(keys, styles, blank, newSnippet{}), cmd
}

func Editing(keys binding.Keys, styles look.Styles, browsed BrowsedSnippet) (Session, tea.Cmd) {
	stored := browsed.Snippet
	fragment := stored.FirstFragment()
	original := entered{
		title:       stored.Title().String(),
		description: stored.Description().String(),
		content:     fragment.Content().String(),
	}
	filled, cmd := newForm(formKeysOf(keys), original, readOnlyIfTabbed(fragment, styles.CodeStyle))
	target := storedSnippet{id: stored.ID(), selection: browsed.Selection, loadedUpdatedAt: stored.UpdatedAt()}

	return newSession(keys, styles, filled, target), cmd
}

func newSession(keys binding.Keys, styles look.Styles, opened form, target saveTarget) Session {
	return Session{
		keys:   keys,
		form:   opened,
		target: target,
		styles: styles,
		outer:  look.Size{Width: 0, Height: 0},
		saving: false,
	}
}

func formKeysOf(keys binding.Keys) formKeys {
	return formKeys{fields: keys.For(binding.ScopeEditor), content: keys.For(binding.ScopeContent)}
}

func (s Session) Update(msg tea.Msg) outcome.Step {
	switch msg := msg.(type) {
	case tea.KeyPressMsg, tea.PasteMsg:
		return s.formUpdated(msg)
	case look.Resized:
		return outcome.Stay(s.resized(msg.Box))
	case look.Restyled:
		return outcome.Stay(s.restyled(msg.Styles))
	case SaveFinished:
		return s.saved(msg)
	}

	return outcome.Stay(s)
}

func (s Session) Received(received outcome.Outcome) outcome.Step {
	switch received.(type) {
	case outcome.DiscardConfirmed:
		return outcome.Close()
	case outcome.SnippetReloaded:
		return outcome.Close().Passing(received)
	case outcome.QuitAsked:
		if s.form.changed() {
			return s.confirmingUnsaved(quitQuestion, outcome.QuitConfirmed{})
		}
	default:
	}

	return outcome.Stay(s).Passing(received)
}

func (s Session) View() string {
	return s.form.view(s.styles, s.outer)
}

func (s Session) ShortHelp() []key.Binding {
	return s.form.hints()
}

func (s Session) FullHelp() [][]key.Binding {
	return [][]key.Binding{s.ShortHelp()}
}

func (s Session) formUpdated(msg tea.Msg) outcome.Step {
	next := s

	var (
		asked request
		cmd   tea.Cmd
	)

	next.form, asked, cmd = s.form.update(msg)

	return next.requested(asked).Running(cmd)
}

func (s Session) requested(asked request) outcome.Step {
	switch asked {
	case requestSave:
		return s.saveStarted()
	case requestCancel:
		return s.cancelled()
	case requestRefusePasteWithTabs:
		return s.pasteRefused(pasteHasTabs)
	case requestRefuseOverlongPaste:
		return s.pasteRefused(pasteOverflowsContent)
	case requestNothing:
	}

	return outcome.Stay(s)
}

func (s Session) pasteRefused(refusal string) outcome.Step {
	return outcome.Stay(s).Passing(outcome.NoticeShown{Text: refusedPasteText(refusal, s.form.externalEditorKey())})
}

func (s Session) saveStarted() outcome.Step {
	if s.saving {
		return outcome.Stay(s)
	}

	next := s
	next.saving = true

	return s.target.savingAs(next, s.form.entered())
}

func (s Session) cancelled() outcome.Step {
	switch {
	case s.saving:
		return outcome.Stay(s)
	case s.form.changed():
		return s.confirmingUnsaved(discardQuestion, outcome.DiscardConfirmed{})
	}

	return outcome.Close()
}

func (s Session) saved(msg SaveFinished) outcome.Step {
	if msg.Err == nil {
		return s.target.closedAfterSave(msg.Snippet)
	}

	next := s
	next.saving = false

	return next.refused(msg.Err)
}

func (s Session) refused(err error) outcome.Step {
	if reload, ok := s.target.reloaded(); ok && errors.Is(err, domain.ErrConflict) {
		return s.confirming(confirm.Question{Title: staleTitle, Text: reloadQuestion}, reload)
	}

	fieldErrors := domain.FieldErrors(err)
	if len(fieldErrors) == 0 {
		return outcome.Stay(s).Passing(outcome.SaveFailed{Err: err})
	}

	next := s
	next.form = s.form.withInvalid(fieldErrors)

	return outcome.Stay(next).Passing(outcome.NoticeShown{Text: fieldErrorText(fieldErrors[0])})
}

func (s Session) confirmingUnsaved(question string, onYes outcome.Outcome) outcome.Step {
	return s.confirming(confirm.Question{Title: unsavedTitle, Text: question}, onYes)
}

func (s Session) confirming(asked confirm.Question, onYes outcome.Outcome) outcome.Step {
	return outcome.Stay(s).Opening(confirm.New(s.keys, s.styles, asked, onYes))
}

func (s Session) resized(screen look.Size) Session {
	s.outer = screen.Share(overlayPercent)
	s.form = s.form.resized(s.outer)

	return s
}

func (s Session) restyled(styles look.Styles) Session {
	s.styles = styles
	s.form = s.form.restyled(styles.CodeStyle)

	return s
}
