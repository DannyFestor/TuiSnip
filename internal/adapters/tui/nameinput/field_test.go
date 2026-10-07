package nameinput_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/nameinput"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const overMaxRunes = 201

func TestField_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		initial string
		keys    []tea.KeyPressMsg
		want    nameinput.Result
	}{
		{
			name: "keeps typing on a printable key",
			keys: keypress.Typed("d"),
			want: typing(),
		},
		{
			name: "commits the typed name on accept",
			keys: append(keypress.Typed("docker"), keypress.Special(tea.KeyEnter)),
			want: committed("docker"),
		},
		{
			name:    "commits the name it started with",
			initial: "go",
			keys:    []tea.KeyPressMsg{keypress.Special(tea.KeyEnter)},
			want:    committed("go"),
		},
		{
			name:    "commits an edit of the name it started with",
			initial: "go",
			keys:    append(keypress.Typed("lang"), keypress.Special(tea.KeyEnter)),
			want:    committed("golang"),
		},
		{
			name: "cancels on cancel",
			keys: append(keypress.Typed("docker"), keypress.Special(tea.KeyEscape)),
			want: nameinput.Result{Ending: nameinput.Cancelled, Name: "", Err: nil},
		},
		{
			name: "refuses a blank name",
			keys: append(keypress.Typed("  "), keypress.Special(tea.KeyEnter)),
			want: refused(value.ErrBlankFolderName),
		},
		{
			name: "refuses a name over 200 runes",
			keys: append(keypress.Typed(strings.Repeat("a", overMaxRunes)), keypress.Special(tea.KeyEnter)),
			want: refused(value.ErrFolderNameTooLong),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, got := pressed(newField(t, tt.initial), tt.keys...)

			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("keeps the name after a refusal so it can be fixed", func(t *testing.T) {
		t.Parallel()

		field, refusal := pressed(newField(t, ""), keypress.Special(tea.KeyEnter))
		require.Equal(t, refused(value.ErrBlankFolderName), refusal)

		_, got := pressed(field, append(keypress.Typed("docker"), keypress.Special(tea.KeyEnter))...)

		assert.Equal(t, committed("docker"), got)
	})

	t.Run("commits on a remapped accept key", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeNameInput][binding.Accept] = []string{"ctrl+j"}
		field, _ := nameinput.New(keys, look.NewStyles(look.SchemeDark), validFolderName, "go")

		_, got := pressed(field, keypress.Special(tea.KeyEnter), keypress.Ctrl('j'))

		assert.Equal(t, committed("go"), got)
	})

	t.Run("takes a paste", func(t *testing.T) {
		t.Parallel()

		field, _, _ := newField(t, "").Update(tea.PasteMsg{Content: "docker"})

		_, got := pressed(field, keypress.Special(tea.KeyEnter))

		assert.Equal(t, committed("docker"), got)
	})
}

func typing() nameinput.Result {
	return nameinput.Result{Ending: nameinput.Typing, Name: "", Err: nil}
}

func committed(name string) nameinput.Result {
	return nameinput.Result{Ending: nameinput.Committed, Name: name, Err: nil}
}

func refused(err error) nameinput.Result {
	return nameinput.Result{Ending: nameinput.Refused, Name: "", Err: err}
}

func TestField_View(t *testing.T) {
	t.Parallel()

	field, _ := pressed(newField(t, "go"), keypress.Typed("lang")...)

	t.Run("shows the typed name", func(t *testing.T) {
		t.Parallel()

		assert.Contains(t, ansi.Strip(field.WithWidth(20).View()), "golang")
	})

	t.Run("fits a long name, cursor included, into the width", func(t *testing.T) {
		t.Parallel()

		const width = 4

		narrow, _ := pressed(newField(t, "").WithWidth(width), keypress.Typed("golang")...)

		assert.Equal(t, width, ansi.StringWidth(narrow.View()))
	})

	t.Run("fits a long name it started with into the width", func(t *testing.T) {
		t.Parallel()

		const width = 4

		assert.Equal(t, width, ansi.StringWidth(newField(t, "golang").WithWidth(width).View()))
	})
}

func TestField_WithStyles(t *testing.T) {
	t.Parallel()

	light := look.NewStyles(look.SchemeLight)

	assert.Equal(t, styledField(t, light, "go").View(), newField(t, "go").WithStyles(light).View())
}

func TestField_Typed(t *testing.T) {
	t.Parallel()

	t.Run("holds the name it started with", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "go", newField(t, "go").Typed())
	})

	t.Run("holds the name as typed so far", func(t *testing.T) {
		t.Parallel()

		field, _ := pressed(newField(t, "go"), append(keypress.Typed(" lang"), keypress.Special(tea.KeyBackspace))...)

		assert.Equal(t, "go lan", field.Typed())
	})
}

func TestField_ShortHelp(t *testing.T) {
	t.Parallel()

	field := newField(t, "")

	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeNameInput).ShortHelp(), field.ShortHelp())
	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeNameInput).FullHelp(), field.FullHelp())
}

func newField(t *testing.T, initial string) nameinput.Field {
	t.Helper()

	return styledField(t, look.NewStyles(look.SchemeDark), initial)
}

func styledField(t *testing.T, styles look.Styles, initial string) nameinput.Field {
	t.Helper()

	field, _ := nameinput.New(testsettings.Default(t).Keys, styles, validFolderName, initial)

	return field
}

func pressed(field nameinput.Field, keys ...tea.KeyPressMsg) (nameinput.Field, nameinput.Result) {
	result := typing()

	for _, pressedKey := range keys {
		field, result, _ = field.Update(pressedKey)
	}

	return field, result
}

func validFolderName(raw string) error {
	_, err := value.NewFolderName(raw)

	return err
}
