package editoverlay_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestSession_View(t *testing.T) {
	t.Parallel()

	t.Run("opens on Title", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		assert.Contains(t, screen.Screen(), overlayTitle)
		assert.Contains(t, screen.Screen(), "› Title")
		assert.Contains(t, screen.Screen(), contentEntryHint)
		assert.NotContains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("marks unsaved changes in the title", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Letter('x'))

		assert.Contains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("types Pane keys into the field instead of acting on them", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Typed("q4/")...)

		assert.Empty(t, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "q4/")
	})

	t.Run("hides the entry hint inside Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)

		assert.NotContains(t, screen.Screen(), contentEntryHint)
	})

	t.Run("scrolls a long Title to keep its end and the cursor in the frame", func(t *testing.T) {
		t.Parallel()

		title := strings.Repeat("0123456789", 10)
		screen := editing(t)

		screen.Press(keypress.Typed(title)...)

		assert.Contains(t, screen.Screen(), "│› Title       "+title[len(title)-titleInputWidth:]+" │")
	})

	t.Run("scrolls long Content to keep its last line in the frame", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)

		for number := 1; number <= contentRows+1; number++ {
			screen.Press(keypress.Typed(fmt.Sprintf("line %02d", number))...)
			screen.Press(keypress.Special(tea.KeyEnter))
		}

		assert.Contains(t, screen.Screen(), fmt.Sprintf("line %02d", contentRows+1))
		assert.NotContains(t, screen.Screen(), "line 01")
	})

	t.Run("names only the bound entry key", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeEditor][binding.OpenField] = []string{}
		screen := editingWith(t, keys)

		assert.Contains(t, screen.Screen(), "  down to edit")
	})

	t.Run("names no entry key when none is bound", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeEditor][binding.OpenField] = []string{}
		keys[binding.ScopeEditor][binding.NextField] = []string{}
		screen := editingWith(t, keys)

		assert.NotContains(t, screen.Screen(), "to edit")
	})
}

func TestSession_ShortHelp(t *testing.T) {
	t.Parallel()

	t.Run("shows the field Bindings outside Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		assert.Equal(t, "ctrl+s save · esc cancel · down field", screen.Hints())
	})

	t.Run("shows the content Bindings inside Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)

		assert.Equal(t, "ctrl+s save · esc leave", screen.Hints())
	})
}

func TestSession_save(t *testing.T) {
	t.Parallel()

	t.Run("asks to save every field", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Typed("Prune")...)
		screen.Press(keypress.Special(tea.KeyTab))
		screen.Press(keypress.Typed("Reclaim")...)
		screen.Press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEnter))
		screen.Press(keypress.Typed("docker")...)
		screen.Press(keypress.Special(tea.KeyEnter))
		screen.Press(keypress.Typed("prune")...)
		screen.Press(save())

		want := outcome.SaveRequested{Input: input("Prune", "Reclaim", "docker\nprune")}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
		assert.True(t, screen.IsOpen())
	})

	t.Run("closes and reports the saved Snippet", func(t *testing.T) {
		t.Parallel()

		saved := savedSnippet(t)
		screen := editing(t)

		screen.Press(keypress.Letter('x'), save())
		screen.Send(editoverlay.SaveFinished{Snippet: saved, Err: nil})

		assert.Contains(t, screen.Outcomes(), outcome.Outcome(outcome.SnippetSaved{ID: saved.ID()}))
		assert.False(t, screen.IsOpen())
	})

	t.Run("up on the first line of Content returns to Description", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(
			keypress.Special(tea.KeyDown),
			keypress.Special(tea.KeyDown),
			keypress.Special(tea.KeyDown),
		)
		screen.Press(keypress.Typed("body")...)
		screen.Press(keypress.Special(tea.KeyUp))
		screen.Press(keypress.Typed("more")...)
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "more", "body")}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), contentEntryHint)
	})

	t.Run("up below the first line of Content stays in Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("one")...)
		screen.Press(keypress.Special(tea.KeyEnter))
		screen.Press(keypress.Typed("two")...)
		screen.Press(keypress.Special(tea.KeyUp))
		screen.Press(keypress.Typed(" more")...)
		screen.Press(save())

		want := outcome.SaveRequested{Input: input("", "", "one more\ntwo")}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("up on Title stays on Title", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Special(tea.KeyUp))
		screen.Press(keypress.Typed("kept")...)
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("kept", "", "")}}, screen.Outcomes())
	})

	t.Run("down past Content enters it instead of moving on", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(
			keypress.Special(tea.KeyDown),
			keypress.Special(tea.KeyDown),
			keypress.Special(tea.KeyDown),
		)
		screen.Press(keypress.Typed("body")...)
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "body")}}, screen.Outcomes())
	})

	t.Run("types nothing on Content until it is entered", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Special(tea.KeyDown), keypress.Special(tea.KeyDown))
		screen.Press(keypress.Typed("lost")...)
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "")}}, screen.Outcomes())
	})
}

func TestSession_pendingSave(t *testing.T) {
	t.Parallel()

	t.Run("a second save while one is pending asks once", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Letter('x'), save(), save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("x", "", "")}}, screen.Outcomes())
	})

	t.Run("esc while a save is pending stays without asking", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Letter('x'), save(), keypress.Special(tea.KeyEscape))

		assert.NotContains(t, screen.Screen(), discardQuestion)
		assert.Contains(t, screen.Screen(), overlayTitle)
	})

	t.Run("a quit asked while a save is pending closes with the saved overlay", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Letter('x'), save())
		screen.Offer(outcome.QuitAsked{})
		asked := screen.Screen()
		screen.Send(editoverlay.SaveFinished{Snippet: savedSnippet(t), Err: nil})

		assert.Contains(t, asked, quitQuestion)
		assert.False(t, screen.IsOpen())
	})

	t.Run("a rejected save can be saved again", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(save())
		screen.Send(editoverlay.SaveFinished{Snippet: domain.Snippet{}, Err: errDatabaseLocked})
		screen.Press(save())

		requested := outcome.Outcome(outcome.SaveRequested{Input: input("", "", "")})
		assert.Equal(
			t,
			[]outcome.Outcome{requested, outcome.SaveFailed{Err: errDatabaseLocked}, requested},
			screen.Outcomes(),
		)
	})

	t.Run("a rejected save can be cancelled", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(save())
		screen.Send(editoverlay.SaveFinished{Snippet: domain.Snippet{}, Err: errDatabaseLocked})
		screen.Press(keypress.Special(tea.KeyEscape))

		assert.False(t, screen.IsOpen())
	})
}

func TestSession_saveFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want outcome.Outcome
	}{
		{
			name: "names a blank Title",
			err:  domain.OnField(domain.FieldTitle, value.ErrBlankTitle),
			want: outcome.NoticeShown{Text: "Title is blank"},
		},
		{
			name: "names a long Title",
			err:  domain.OnField(domain.FieldTitle, value.ErrTitleTooLong),
			want: outcome.NoticeShown{Text: "Title is longer than 200 characters"},
		},
		{
			name: "names a long Description",
			err:  domain.OnField(domain.FieldDescription, value.ErrDescriptionTooLong),
			want: outcome.NoticeShown{Text: "Description is longer than 2000 characters"},
		},
		{
			name: "names large Content",
			err:  domain.OnField(domain.FieldContent, value.ErrContentTooLong),
			want: outcome.NoticeShown{Text: "Content is larger than 256 KiB"},
		},
		{
			name: "names the first of several fields",
			err: errors.Join(
				domain.OnField(domain.FieldTitle, value.ErrBlankTitle),
				domain.OnField(domain.FieldContent, value.ErrContentTooLong),
			),
			want: outcome.NoticeShown{Text: "Title is blank"},
		},
		{
			name: "shows a generic notice for a rule it has no words for",
			err:  domain.OnField(domain.FieldLanguage, value.ErrUnknownLanguage),
			want: outcome.NoticeShown{Text: "Something went wrong; see the log"},
		},
		{
			name: "reports any other error as a failure",
			err:  errDatabaseLocked,
			want: outcome.SaveFailed{Err: errDatabaseLocked},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := editing(t)

			screen.Press(save())
			screen.Send(editoverlay.SaveFinished{Snippet: domain.Snippet{}, Err: tt.err})

			assert.Equal(t, tt.want, screen.Outcomes()[len(screen.Outcomes())-1])
			assert.True(t, screen.IsOpen())
		})
	}
}

func TestSession_cancel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		keys        []tea.KeyPressMsg
		wantOverlay bool
		wantText    string
	}{
		{name: "closes at once without changes", keys: []tea.KeyPressMsg{keypress.Special(tea.KeyEscape)}},
		{
			name:        "asks before discarding changes",
			keys:        []tea.KeyPressMsg{keypress.Letter('x'), keypress.Special(tea.KeyEscape)},
			wantOverlay: true,
			wantText:    discardQuestion,
		},
		{
			name: "discards on y",
			keys: []tea.KeyPressMsg{keypress.Letter('x'), keypress.Special(tea.KeyEscape), keypress.Letter('y')},
		},
		{
			name:        "keeps the changes on n",
			keys:        []tea.KeyPressMsg{keypress.Letter('x'), keypress.Special(tea.KeyEscape), keypress.Letter('n')},
			wantOverlay: true,
			wantText:    unsavedTitle,
		},
		{
			name:        "esc inside Content leaves it and keeps the overlay",
			keys:        append(enterContent(), keypress.Special(tea.KeyEscape)),
			wantOverlay: true,
			wantText:    contentEntryHint,
		},
		{
			name: "a second esc cancels",
			keys: append(enterContent(), keypress.Special(tea.KeyEscape), keypress.Special(tea.KeyEscape)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := editing(t)

			screen.Press(tt.keys...)

			assert.Equal(t, tt.wantOverlay, strings.Contains(screen.Screen(), overlayTitle))
			assert.Equal(t, tt.wantOverlay, screen.IsOpen())
			assert.Contains(t, screen.Screen(), tt.wantText)
		})
	}
}

func TestSession_Received(t *testing.T) {
	t.Parallel()

	t.Run("asks before quitting with changes", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Letter('x'))
		screen.Offer(outcome.QuitAsked{})

		assert.Contains(t, screen.Screen(), quitQuestion)
		assert.Equal(t, "y yes · n no", screen.Hints())
		assert.Empty(t, screen.Outcomes())
	})

	t.Run("confirms the quit on y", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Letter('x'))
		screen.Offer(outcome.QuitAsked{})
		screen.Press(keypress.Letter('y'))

		assert.Equal(t, []outcome.Outcome{outcome.QuitConfirmed{}}, screen.Outcomes())
	})

	t.Run("asks to quit instead while asking to discard", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Letter('x'), keypress.Special(tea.KeyEscape))
		screen.Offer(outcome.QuitAsked{})
		screen.Press(keypress.Letter('y'))

		assert.Equal(t, []outcome.Outcome{outcome.QuitConfirmed{}}, screen.Outcomes())
	})

	t.Run("lets the quit through without changes", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Offer(outcome.QuitAsked{})

		assert.Equal(t, []outcome.Outcome{outcome.QuitAsked{}}, screen.Outcomes())
	})

	t.Run("ignores other keys and pastes while asking", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(keypress.Letter('x'))
		screen.Offer(outcome.QuitAsked{})
		screen.Press(keypress.Letter('z'))
		screen.Send(tea.PasteMsg{Content: "pasted"})

		assert.Contains(t, screen.Screen(), quitQuestion)
		assert.NotContains(t, screen.Screen(), "pasted")
	})
}

func TestSession_paste(t *testing.T) {
	t.Parallel()

	t.Run("refuses a paste with tabs into Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: "if x {\n\treturn\n}"})

		want := outcome.NoticeShown{Text: "Pasted text contains tabs; use ctrl+e to edit in $EDITOR"}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
		assert.NotContains(t, screen.Screen(), "return")
		assert.NotContains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("inserts a paste without tabs into Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: "if x {\n    return\n}"})

		assert.Contains(t, screen.Screen(), "return")
	})

	t.Run("pastes text with tabs into Title", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Send(tea.PasteMsg{Content: "Pasted\ttitle"})

		assert.Empty(t, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "Pasted")
	})

	t.Run("names the configured external editor key", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeContent][binding.OpenInEditor] = []string{"ctrl+o"}
		screen := editingWith(t, keys)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: "\t"})

		want := outcome.NoticeShown{Text: "Pasted text contains tabs; use ctrl+o to edit in $EDITOR"}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("names no key when the external editor is unbound", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeContent][binding.OpenInEditor] = []string{}
		screen := editingWith(t, keys)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: "\t"})

		assert.Equal(t, []outcome.Outcome{outcome.NoticeShown{Text: "Pasted text contains tabs"}}, screen.Outcomes())
	})
}
