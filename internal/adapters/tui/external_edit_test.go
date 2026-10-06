package tui_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

const (
	noEditorText      = "No editor found; set `editor` in config.toml"
	editorFailedText  = "Editor exited with an error; changes discarded"
	editedContent     = "docker image prune\n"
	storedContent     = "docker system prune\n"
	editedTitle       = "Prune everything"
	unsavedEditTitle  = editOverlayTitle + " •"
	externalEditorKey = 'E'
)

var errExecFormat = errors.New("exec format error")

func TestModel_externalEditor_fileName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		language string
		want     string
	}{
		{language: "Go", want: "snippet.go"},
		{language: "Docker", want: "Dockerfile"},
		{language: "plaintext", want: "snippet.txt"},
		{language: "Python 2", want: "snippet.txt"},
		{language: "Groff", want: "snippet.txt"},
	}
	for _, tt := range tests {
		t.Run("names the file "+tt.want+" for "+tt.language, func(t *testing.T) {
			t.Parallel()

			editor := NewMockExternalEditor(t)
			editor.EXPECT().Open(mock.Anything, tt.want, storedContent).Return(nil, domain.ErrNoEditor)
			screen := start(t, externallyEditingModel(t, editor, storedIn(t, tt.language)), wideWidth, wideHeight)

			screen.press(keypress.Letter('3'), keypress.Letter(externalEditorKey))

			assert.Contains(t, screen.screen(), noEditorText)
		})
	}
}

func TestModel_externalEditor_failures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		openErr  error
		runErr   error
		wantText string
	}{
		{name: "says no editor was found", openErr: domain.ErrNoEditor, wantText: noEditorText},
		{
			name:     "shows the error an editor could not start with",
			runErr:   domain.EditorStartError{Err: errExecFormat},
			wantText: errExecFormat.Error(),
		},
		{
			name:     "discards the content of an editor that exited with an error",
			runErr:   domain.ErrEditorFailed,
			wantText: editorFailedText,
		},
		{name: "shows a generic failure for anything else", runErr: errDatabaseLocked, wantText: genericFailure},
	}
	for _, tt := range tests {
		t.Run(tt.name+" and opens nothing", func(t *testing.T) {
			t.Parallel()

			editor := NewMockExternalEditor(t)
			editor.EXPECT().
				Open(mock.Anything, mock.Anything, mock.Anything).
				Return(editorRun(editedContent, wrapped(tt.runErr)), wrapped(tt.openErr))
			screen := start(t, externallyEditingModel(t, editor, storedIn(t, "Bash")), wideWidth, wideHeight)

			screen.press(keypress.Letter('3'), keypress.Letter(externalEditorKey))

			assert.Contains(t, screen.screen(), tt.wantText)
			assert.NotContains(t, screen.screen(), editOverlayTitle)
		})
	}
}

func TestModel_externalEditor_edited(t *testing.T) {
	t.Parallel()

	t.Run("opens the edited content in the edit overlay as an unsaved change", func(t *testing.T) {
		t.Parallel()

		screen := start(
			t,
			externallyEditingModel(t, editorReturning(t, editedContent), storedIn(t, "Bash")),
			wideWidth,
			wideHeight,
		)

		screen.press(keypress.Letter('3'), keypress.Letter(externalEditorKey))

		assert.Contains(t, screen.screen(), unsavedEditTitle)
		assert.Contains(t, screen.screen(), "docker image prune")
	})

	t.Run("puts content edited from the edit overlay back into it", func(t *testing.T) {
		t.Parallel()

		screen := start(
			t,
			externallyEditingModel(t, editorReturning(t, editedContent), storedIn(t, "Bash")),
			wideWidth,
			wideHeight,
		)

		screen.press(keypress.Letter('3'), keypress.Letter('e'), keypress.Ctrl('e'))

		assert.Contains(t, screen.screen(), unsavedEditTitle)
		assert.Contains(t, screen.screen(), "docker image prune")
	})

	t.Run("does nothing when the content is unchanged", func(t *testing.T) {
		t.Parallel()

		screen := start(
			t,
			externallyEditingModel(t, editorReturning(t, storedContent), storedIn(t, "Bash")),
			wideWidth,
			wideHeight,
		)

		screen.press(keypress.Letter('3'), keypress.Letter(externalEditorKey))

		assert.NotContains(t, screen.screen(), editOverlayTitle)
	})

	t.Run("suspends the program while the editor runs", func(t *testing.T) {
		t.Parallel()

		model := externallyEditingModel(t, editorReturning(t, editedContent), storedIn(t, "Bash"))
		program := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(wideWidth, wideHeight))
		waitForOutput(t, program, editedTitle)

		program.Send(keypress.Letter('3'))
		program.Send(keypress.Letter(externalEditorKey))

		waitForOutput(t, program, unsavedEditTitle)
		program.Send(keypress.Ctrl('c'))
		program.Send(keypress.Letter('y'))
		program.WaitFinished(t, teatest.WithFinalTimeout(programTimeout))
	})
}

func externallyEditingModel(t *testing.T, editor tui.ExternalEditor, snippets ...domain.Snippet) tui.Model {
	t.Helper()

	return modelWith(t, actions{
		lister:         listerOf(t, snippets...),
		treeLister:     treeOf(t, emptyTree()),
		copier:         NewMockSnippetCopier(t),
		creator:        NewMockSnippetCreator(t),
		searcher:       NewMockSnippetSearcher(t),
		externalEditor: editor,
	})
}

func storedIn(t *testing.T, language string) domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return testkit.Snippet(t, testkit.SnippetSpec{
		ID:       ids.NewSnippetID(),
		Title:    editedTitle,
		Fragment: testkit.FragmentSpec{ID: ids.NewFragmentID(), Language: language, Content: storedContent},
	})
}

func editorReturning(t *testing.T, edited string) *MockExternalEditor {
	t.Helper()

	editor := NewMockExternalEditor(t)
	editor.EXPECT().Open(mock.Anything, "snippet.sh", storedContent).Return(editorRun(edited, nil), nil)

	return editor
}

func editorRun(edited string, err error) tui.EditorRun {
	return func(io.Reader, io.Writer, io.Writer) (string, error) {
		if err != nil {
			return "", err
		}

		return edited, nil
	}
}

func wrapped(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("editor.Launcher.Open: %w", err)
}

func waitForOutput(t *testing.T, program *teatest.TestModel, text string) {
	t.Helper()

	teatest.WaitFor(t, program.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte(text))
	}, teatest.WithDuration(programTimeout))
}
