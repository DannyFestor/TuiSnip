package folderpane_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const overMaxRunes = 201

func TestPane_UpdateNewFolder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		before []tea.KeyPressMsg
		want   func(foldertree.Sample) domain.FolderID
	}{
		{
			name:   "creates inside the selected Folder",
			before: []tea.KeyPressMsg{keypress.Letter('j'), keypress.Letter('j')},
			want:   func(sample foldertree.Sample) domain.FolderID { return sample.Go.ID() },
		},
		{
			name:   "creates at the Root when the Root is selected",
			before: nil,
			want:   func(foldertree.Sample) domain.FolderID { return domain.FolderID{} },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sample := foldertree.New(t)
			pane := pressed(samplePane(t, sample), tt.before...)
			pane = pressed(pane, newFolderTyped("errors")...)

			pane, outcomes, _ := pane.Update(keypress.Special(tea.KeyEnter))

			assert.Equal(t, []outcome.Outcome{outcome.FolderCreateRequested{
				Input: folder.CreateInput{Name: "errors", ParentID: tt.want(sample)},
			}}, outcomes)
			assert.False(t, pane.Naming())
		})
	}
}

func TestPane_UpdateRename(t *testing.T) {
	t.Parallel()

	t.Run("renames the Folder under the cursor, starting from its name", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := pressed(samplePane(t, sample), keypress.Letter('j'), keypress.Letter('r'))
		pane = pressed(pane, keypress.Typed("s")...)

		pane, outcomes, _ := pane.Update(keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.FolderRenameRequested{
			Input: folder.RenameInput{FolderID: sample.Docker.ID(), Name: "dockers"},
		}}, outcomes)
		assert.False(t, pane.Naming())
	})

	t.Run("leaves the Root alone", func(t *testing.T) {
		t.Parallel()

		pane, outcomes, _ := samplePane(t, foldertree.New(t)).Update(keypress.Letter('r'))

		assert.Empty(t, outcomes)
		assert.False(t, pane.Naming())
	})
}

func TestPane_UpdateWhileNaming(t *testing.T) {
	t.Parallel()

	t.Run("cancel drops the typed row", func(t *testing.T) {
		t.Parallel()

		pane := pressed(samplePane(t, foldertree.New(t)), keypress.Letter('N'), keypress.Letter('x'))

		pane, outcomes, _ := pane.Update(keypress.Special(tea.KeyEscape))

		assert.Empty(t, outcomes)
		assert.False(t, pane.Naming())
		assert.Len(t, viewLines(pane), 4)
	})

	t.Run("movement keys type instead of moving the cursor", func(t *testing.T) {
		t.Parallel()

		pane := pressed(samplePane(t, foldertree.New(t)), keypress.Letter('N'))

		pane, outcomes, _ := pane.Update(keypress.Letter('j'))

		assert.Empty(t, outcomes)
		assert.Equal(t, domain.FolderID{}, pane.Selected())
		assert.True(t, pane.Naming())
	})

	refusals := []struct {
		name  string
		typed string
		want  string
	}{
		{name: "refuses a blank name", typed: "  ", want: "Folder name is blank"},
		{
			name:  "refuses a name over 200 characters",
			typed: strings.Repeat("a", overMaxRunes),
			want:  "Folder name is longer than 200 characters",
		},
	}
	for _, tt := range refusals {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pane := pressed(samplePane(t, foldertree.New(t)), keypress.Letter('N'))
			pane = pressed(pane, keypress.Typed(tt.typed)...)

			pane, outcomes, _ := pane.Update(keypress.Special(tea.KeyEnter))

			assert.Equal(t, []outcome.Outcome{outcome.NoticeShown{Text: tt.want}}, outcomes)
			assert.True(t, pane.Naming(), "the row stays open to fix the name")
		})
	}

	t.Run("takes a paste", func(t *testing.T) {
		t.Parallel()

		pane := pressed(samplePane(t, foldertree.New(t)), keypress.Letter('N'))
		pane, _, _ = pane.Update(tea.PasteMsg{Content: "docker"})

		_, outcomes, _ := pane.Update(keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.FolderCreateRequested{
			Input: folder.CreateInput{Name: "docker", ParentID: domain.FolderID{}},
		}}, outcomes)
	})
}

func TestPane_ViewWhileNaming(t *testing.T) {
	t.Parallel()

	t.Run("types a new Folder on a row below the cursor, one level inside", func(t *testing.T) {
		t.Parallel()

		pane := pressed(
			samplePane(t, foldertree.New(t)),
			keypress.Letter('j'),
			keypress.Letter('j'),
			keypress.Letter('N'),
		)
		pane = pressed(pane, keypress.Typed("errs")...)

		lines := viewLines(pane)

		require.Len(t, lines, 5)
		assert.Equal(t, "▾ go                  2", lines[2])
		assert.Regexp(t, `^    ERRS +0$`, lines[3])
		assert.Equal(t, "    testing           1", lines[4])
	})

	t.Run("types a new Folder at the Root without an indent", func(t *testing.T) {
		t.Parallel()

		pane := pressed(samplePane(t, foldertree.New(t)), keypress.Letter('N'))

		assert.Regexp(t, `^   +0$`, viewLines(pane)[1])
	})

	t.Run("scrolls a long name inside the row instead of cutting it", func(t *testing.T) {
		t.Parallel()

		pane := pressed(samplePane(t, foldertree.New(t)), keypress.Letter('N'))
		pane = pressed(pane, keypress.Typed("abcdefghijklmnopqrstuvwxyz")...)

		assert.Regexp(t, `^  [A-Z]*XYZ +0$`, viewLines(pane)[1])
	})

	t.Run("types a rename in place of the row", func(t *testing.T) {
		t.Parallel()

		pane := pressed(samplePane(t, foldertree.New(t)), keypress.Letter('j'), keypress.Letter('r'))

		lines := viewLines(pane)

		require.Len(t, lines, 4)
		assert.Regexp(t, `^  DOCKER +3$`, lines[1])
	})
}

func TestPane_ShortHelpWhileNaming(t *testing.T) {
	t.Parallel()

	pane := pressed(samplePane(t, foldertree.New(t)), keypress.Letter('N'))

	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeNameInput).ShortHelp(), pane.ShortHelp())
}
