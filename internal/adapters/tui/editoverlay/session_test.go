package editoverlay_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/screencolour"
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

		assert.Contains(t, screen.Screen(), "│› Title          "+title[len(title)-titleInputWidth:]+" │")
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

		assert.NotRegexp(t, `Content .*to edit`, screen.Screen())
	})

	t.Run("draws in the new Styles", func(t *testing.T) {
		t.Parallel()

		light := look.NewStyles(look.SchemeLight)
		screen := editing(t)

		screen.Send(look.Restyled{Styles: light})

		assert.Equal(t, editingStyled(t, testsettings.Default(t).Keys, light).StyledScreen(), screen.StyledScreen())
	})

	t.Run("draws Content being typed in the new Styles", func(t *testing.T) {
		t.Parallel()

		light := look.NewStyles(look.SchemeLight)
		screen := editing(t)
		screen.Press(twoContentLines()...)

		screen.Send(look.Restyled{Styles: light})

		styledFromStart := editingStyled(t, testsettings.Default(t).Keys, light)
		styledFromStart.Press(twoContentLines()...)
		assert.Equal(t, styledFromStart.StyledScreen(), screen.StyledScreen())
	})

	t.Run("numbers the lines of Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(twoContentLines()...)

		assert.Regexp(t, `1 a\s`, screen.Screen())
		assert.Regexp(t, `2 b\s`, screen.Screen())
	})
}

func TestSession_ShortHelp(t *testing.T) {
	t.Parallel()

	t.Run("shows the field Bindings outside Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		want := "ctrl+s save · esc cancel · down field · ctrl+t Tags · ctrl+l Language · ctrl+e external editor · " +
			"alt+right next Fragment · alt+left previous Fragment · ctrl+n add Fragment · ctrl+x remove Fragment · " +
			"alt+shift+left move Fragment left · alt+shift+right move Fragment right"
		assert.Equal(t, want, screen.Hints())
	})

	t.Run("shows the content Bindings inside Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)

		want := "ctrl+s save · esc leave · tab indent · shift+tab dedent · ctrl+e external editor"
		assert.Equal(t, want, screen.Hints())
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
		screen.Press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyDown), keypress.Special(tea.KeyDown))
		screen.Press(keypress.Special(tea.KeyDown), keypress.Special(tea.KeyEnter))
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

		reported := outcome.SnippetSaved{ID: saved.ID(), Selection: destination().Selection}
		assert.Contains(t, screen.Outcomes(), outcome.Outcome(reported))
		assert.False(t, screen.IsOpen())
	})

	t.Run("files a Snippet for a Tag at the Root carrying the Tag, and returns to the Tag", func(t *testing.T) {
		t.Parallel()

		tag := testkit.Tag(t, testkit.TagSpec{Name: "docker"})
		selection := browseselection.WithTag(tag.ID())
		opened, _ := editoverlay.New(
			testsettings.Default(t).Keys,
			look.NewStyles(look.SchemeDark),
			editoverlay.Options{},
			editoverlay.Destination{Selection: selection, Tags: []domain.Tag{tag}, Language: value.PlainText()},
		)
		screen := overlaytest.Open(t, screenSize(), opened)
		saved := savedSnippet(t)

		screen.Press(keypress.Letter('x'), save())
		screen.Send(editoverlay.SaveFinished{Snippet: saved, Err: nil})

		created := snippet.CreateInput{
			Title:       "x",
			Description: "",
			Fragments:   []snippet.FragmentInput{{Language: value.PlainText().String()}},
			FolderID:    domain.FolderID{},
			Tags:        []domain.Tag{tag},
			NewTags:     nil,
		}
		assert.Equal(t, []outcome.Outcome{
			outcome.SaveRequested{Input: created},
			outcome.SnippetSaved{ID: saved.ID(), Selection: selection},
		}, screen.Outcomes())
	})

	t.Run("up on the first line of Content returns to Language", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("body")...)
		screen.Press(keypress.Special(tea.KeyUp))
		screen.Press(keypress.Typed("more")...)
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "body")}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "› Language")
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

		screen.Press(toContent()...)
		screen.Press(keypress.Special(tea.KeyDown))
		screen.Press(keypress.Typed("body")...)
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "body")}}, screen.Outcomes())
	})

	t.Run("types nothing on Content until it is entered", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(toContent()...)
		screen.Press(keypress.Typed("lost")...)
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "")}}, screen.Outcomes())
	})
}

func TestSession_folderGoneElsewhere(t *testing.T) {
	t.Parallel()

	t.Run("offers to save a new Snippet at the Root when its Folder is gone", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)
		screen.Press(keypress.Letter('x'), save())
		screen.Send(editoverlay.SaveFinished{Snippet: domain.Snippet{}, Err: errFolderGone})

		assert.Contains(t, screen.Screen(), goneElsewhereTitle)
		assert.Contains(t, screen.Screen(), folderGoneQuestion)
	})

	t.Run("asks to save it at the Root on yes", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)
		screen.Press(keypress.Letter('x'), save())
		screen.Send(editoverlay.SaveFinished{Snippet: domain.Snippet{}, Err: errFolderGone})
		screen.Press(keypress.Letter('y'))

		atRoot := input("x", "", "")
		atRoot.FolderID = domain.FolderID{}
		assert.Equal(t, outcome.Outcome(outcome.SaveRequested{Input: atRoot}), screen.Outcomes()[1])
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

	t.Run("inserts a paste with tabs into Content and keeps the tabs", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: "if x {\n\treturn\n}"})
		screen.Press(save())

		want := outcome.SaveRequested{Input: input("", "", "if x {\n\treturn\n}")}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("pastes text with tabs into Title", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Send(tea.PasteMsg{Content: "Pasted\ttitle"})

		assert.Empty(t, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "Pasted")
	})

	t.Run("refuses a paste over 256 KiB into Content and leaves it unchanged", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: textOfSize(maxContentBytes + 1)})

		assert.Equal(t, []outcome.Outcome{outcome.NoticeShown{Text: pasteTooLarge}}, screen.Outcomes())
		assert.NotContains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("refuses a paste that makes Content larger than 256 KiB", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)
		existing := textOfSize(maxContentBytes / 2)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: existing})
		screen.Send(tea.PasteMsg{Content: textOfSize(maxContentBytes/2 + 1)})
		screen.Press(save())

		want := []outcome.Outcome{
			outcome.NoticeShown{Text: pasteTooLarge},
			outcome.SaveRequested{Input: input("", "", existing)},
		}
		assert.Equal(t, want, screen.Outcomes())
	})

	t.Run("inserts a paste that makes Content exactly 256 KiB", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)
		half := textOfSize(maxContentBytes / 2)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: half})
		screen.Send(tea.PasteMsg{Content: half})
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", half+half)}}, screen.Outcomes())
	})

	t.Run("counts the selection a paste replaces", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)
		full := textOfSize(maxContentBytes-len("\nend")) + "\nend"

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: full})
		screen.Press(keypress.Shift(tea.KeyLeft), keypress.Shift(tea.KeyLeft), keypress.Shift(tea.KeyLeft))
		screen.Send(tea.PasteMsg{Content: "fin"})
		screen.Press(save())

		want := strings.TrimSuffix(full, "end") + "fin"
		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", want)}}, screen.Outcomes())
	})

	t.Run("lets typing take Content past 256 KiB", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)
		full := textOfSize(maxContentBytes)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: full})
		screen.Press(keypress.Typed("!")...)
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", full+"!")}}, screen.Outcomes())
	})
}

func TestSession_indent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pasted  string
		pressed tea.KeyPressMsg
		want    string
	}{
		{
			name:    "indent adds four spaces to the cursor's line only",
			pasted:  "first\n  second\nthird",
			pressed: keypress.Special(tea.KeyTab),
			want:    "first\n      second\nthird",
		},
		{
			name:    "dedent removes four spaces from the cursor's line only",
			pasted:  "    first\n      second\n    third",
			pressed: keypress.Shift(tea.KeyTab),
			want:    "    first\n  second\n    third",
		},
		{
			name:    "dedent removes the spaces a line has when it has fewer than four",
			pasted:  "first\n  second\nthird",
			pressed: keypress.Shift(tea.KeyTab),
			want:    "first\nsecond\nthird",
		},
		{
			name:    "dedent removes one tab from a line that starts with a tab",
			pasted:  "first\n\t\t  second\nthird",
			pressed: keypress.Shift(tea.KeyTab),
			want:    "first\n\t  second\nthird",
		},
		{
			name:    "dedent removes the leading spaces before a tab",
			pasted:  "first\n  \tsecond\nthird",
			pressed: keypress.Shift(tea.KeyTab),
			want:    "first\n\tsecond\nthird",
		},
		{
			name:    "dedent leaves an unindented line unchanged",
			pasted:  "first\nsecond\nthird",
			pressed: keypress.Shift(tea.KeyTab),
			want:    "first\nsecond\nthird",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := editing(t)

			screen.Press(enterContent()...)
			screen.Send(tea.PasteMsg{Content: tt.pasted})
			screen.Press(keypress.Special(tea.KeyUp), tt.pressed, save())

			assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", tt.want)}}, screen.Outcomes())
		})
	}

	t.Run("indent keeps the cursor on the same character", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("ab")...)
		screen.Press(keypress.Special(tea.KeyLeft), keypress.Special(tea.KeyTab), keypress.Letter('x'), save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "    axb")}}, screen.Outcomes())
	})

	t.Run("dedent keeps the cursor on the same character", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: "    ab"})
		screen.Press(keypress.Special(tea.KeyLeft), keypress.Shift(tea.KeyTab), keypress.Letter('x'), save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "axb")}}, screen.Outcomes())
	})

	t.Run("dedent of a tab keeps the cursor on the same character", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: "\tab"})
		screen.Press(keypress.Special(tea.KeyLeft), keypress.Shift(tea.KeyTab), keypress.Letter('x'), save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "axb")}}, screen.Outcomes())
	})

	t.Run("dedent keeps the cursor on the line when it sat in the removed spaces", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: "      a"})
		screen.Press(keypress.Special(tea.KeyHome), keypress.Special(tea.KeyRight), keypress.Shift(tea.KeyTab))
		screen.Press(keypress.Letter('x'), save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "x  a")}}, screen.Outcomes())
	})
}

func TestSession_language(t *testing.T) {
	t.Parallel()

	t.Run("starts a new Snippet in plain text and names the keys that pick", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		assert.Contains(t, screen.Screen(), "  Language       plaintext   (enter or ctrl+l to pick)")
	})

	t.Run("starts a new Snippet in the Destination's Language with nothing unsaved", func(t *testing.T) {
		t.Parallel()

		screen := editingIn(t, destinationInGo(t))

		assert.Contains(t, screen.Screen(), "  Language       Go   (enter or ctrl+l to pick)")
		assert.NotContains(t, screen.Screen(), unsavedTitle)

		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: inputIn("", "", "Go", "")}}, screen.Outcomes())
	})

	t.Run("saves a new Snippet in a picked Language instead of the Destination's", func(t *testing.T) {
		t.Parallel()

		screen := editingIn(t, destinationInGo(t))

		screen.Press(pickLanguage())
		screen.Press(keypress.Typed("bash")...)
		screen.Press(keypress.Special(tea.KeyEnter), save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: inputIn("", "", "Bash", "")}}, screen.Outcomes())
	})

	t.Run("names only the bound pick key", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeEditor][binding.OpenField] = []string{}
		screen := editingWith(t, keys)

		assert.Contains(t, screen.Screen(), "plaintext   (ctrl+l to pick)")
	})

	t.Run("names no pick key when none is bound", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeEditor][binding.OpenField] = []string{}
		keys[binding.ScopeEditor][binding.PickLanguage] = []string{}
		screen := editingWith(t, keys)

		assert.NotContains(t, screen.Screen(), "to pick")
	})

	t.Run("enter on Language opens the Language picker", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(toLanguage()...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Contains(t, screen.Screen(), languagePickerTitle)
		assert.Equal(t, pickerHints, screen.Hints())
	})

	t.Run("down on Language moves on to Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(toContent()...)

		assert.Contains(t, screen.Screen(), "› Content")
		assert.NotContains(t, screen.Screen(), languagePickerTitle)
	})

	t.Run("pick_language opens the Language picker from Title", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(pickLanguage())

		assert.Contains(t, screen.Screen(), languagePickerTitle)
	})

	t.Run("pick_language opens the Language picker from inside Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Press(pickLanguage())

		assert.Contains(t, screen.Screen(), languagePickerTitle)
	})

	t.Run("opens the Language picker with the configured key", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeEditor][binding.PickLanguage] = []string{"ctrl+g"}
		screen := editingWith(t, keys)

		screen.Press(keypress.Ctrl('g'))

		assert.Contains(t, screen.Screen(), languagePickerTitle)
	})

	t.Run("a picked Language is an unsaved change that the save carries", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(pickLanguage())
		screen.Press(keypress.Typed("bash")...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Contains(t, screen.Screen(), unsavedTitle)
		assert.Contains(t, screen.Screen(), "Language       Bash")

		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: inputIn("", "", "Bash", "")}}, screen.Outcomes())
	})

	t.Run("closing the Language picker keeps the Language", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(pickLanguage(), keypress.Special(tea.KeyEscape))

		assert.True(t, screen.IsOpen())
		assert.NotContains(t, screen.Screen(), unsavedTitle)
		assert.NotContains(t, screen.Screen(), languagePickerTitle)
	})

	t.Run("keeps typing into Content after a pick from inside it", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Press(keypress.Letter('a'), pickLanguage(), keypress.Special(tea.KeyEnter), keypress.Letter('b'))
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "ab")}}, screen.Outcomes())
	})

	t.Run("offers only the curated Languages", func(t *testing.T) {
		t.Parallel()

		screen := editingOffering(t, []value.Language{language(t, "YAML"), language(t, "Go")})

		screen.Press(pickLanguage(), keypress.Special(tea.KeyDown), keypress.Special(tea.KeyEnter), save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: inputIn("", "", "Go", "")}}, screen.Outcomes())
	})
}

func BenchmarkSession_keypress(b *testing.B) {
	toStart := []tea.KeyPressMsg{keypress.Ctrl(tea.KeyHome)}
	tests := []struct {
		name    string
		moves   []tea.KeyPressMsg
		pressed tea.KeyPressMsg
	}{
		{name: "typing at the end of 256 KiB", moves: nil, pressed: keypress.Letter('x')},
		{name: "typing at the start of 256 KiB", moves: toStart, pressed: keypress.Letter('x')},
		{name: "breaking a line at the start of 256 KiB", moves: toStart, pressed: keypress.Special(tea.KeyEnter)},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			screen := editingStored(b, storedSnippet(b, typicalGoOfSize(maxContentBytes)))
			screen.Press(enterContent()...)
			screen.Press(tt.moves...)
			screen.RunCommands()
			require.NotEmpty(
				b,
				screencolour.Foreground(b, screen.StyledScreen(), "return"),
				"Content is not highlighted",
			)

			for b.Loop() {
				screen.Press(tt.pressed)
				_ = screen.StyledScreen()
			}
		})
	}
}

func typicalGoOfSize(size int) string {
	const block = `func handle%d(w http.ResponseWriter, r *http.Request) {
	if err := process(r, %d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	fmt.Fprintf(w, "handled %d\n")
}

`

	var source strings.Builder

	for index := 0; source.Len() < size; index++ {
		fmt.Fprintf(&source, block, index, index, index)
	}

	return source.String()[:size]
}
