package binding_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

type bound struct {
	scope binding.Scope
	name  string
	keys  []string
}

func TestKeys_For(t *testing.T) {
	t.Parallel()

	t.Run("hints follow the Scope's hint list, global rows included", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeGlobal, name: binding.NewSnippet, keys: []string{"a"}},
			bound{scope: binding.ScopeGlobal, name: binding.Search, keys: []string{"ctrl+f", "/"}},
			bound{scope: binding.ScopeSnippetList, name: binding.Copy, keys: []string{"c"}},
			bound{scope: binding.ScopeSnippetList, name: binding.Edit, keys: []string{"x"}},
			bound{scope: binding.ScopeSnippetList, name: binding.CycleSort, keys: []string{"o"}},
		)

		got := keys.For(binding.ScopeSnippetList).ShortHelp()

		assert.Equal(t, []string{"c Copy", "x edit", "a new", "o sort", "ctrl+f search"}, hintTexts(got))
	})

	t.Run("Snippet pane hints Copy and edit before search", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeGlobal, name: binding.Search, keys: []string{"/"}},
			bound{scope: binding.ScopeSnippetPane, name: binding.Copy, keys: []string{"y"}},
			bound{scope: binding.ScopeSnippetPane, name: binding.Edit, keys: []string{"e"}},
		)

		got := keys.For(binding.ScopeSnippetPane).ShortHelp()

		assert.Equal(t, []string{"y Copy", "e edit", "/ search"}, hintTexts(got))
	})

	t.Run("Folders hints open, new Folder, rename and delete before search", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeGlobal, name: binding.Open, keys: []string{"enter"}},
			bound{scope: binding.ScopeGlobal, name: binding.Search, keys: []string{"/"}},
			bound{scope: binding.ScopeFolders, name: binding.NewFolder, keys: []string{"N"}},
			bound{scope: binding.ScopeFolders, name: binding.Rename, keys: []string{"r"}},
			bound{scope: binding.ScopeFolders, name: binding.Delete, keys: []string{"d"}},
		)

		got := keys.For(binding.ScopeFolders).ShortHelp()

		assert.Equal(t, []string{"enter open", "N new Folder", "r rename", "d delete", "/ search"}, hintTexts(got))
	})

	t.Run("every pane Scope's hints end with help", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys

		for _, scope := range []binding.Scope{
			binding.ScopeFolders, binding.ScopeTags, binding.ScopeSnippetList, binding.ScopeSnippetPane,
		} {
			hints := hintTexts(keys.For(scope).ShortHelp())
			assert.Equal(t, "? help", hints[len(hints)-1], "%s", scope)
		}
	})

	t.Run("Snippet pane hints Copy, edit and wrap before zoom", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeGlobal, name: binding.Zoom, keys: []string{"z"}},
			bound{scope: binding.ScopeSnippetPane, name: binding.Copy, keys: []string{"y"}},
			bound{scope: binding.ScopeSnippetPane, name: binding.Edit, keys: []string{"e"}},
			bound{scope: binding.ScopeSnippetPane, name: binding.Wrap, keys: []string{"w"}},
		)

		got := keys.For(binding.ScopeSnippetPane).ShortHelp()

		assert.Equal(t, []string{"y Copy", "e edit", "w wrap", "z zoom"}, hintTexts(got))
	})

	zoomHints := []struct {
		scope binding.Scope
		want  []string
	}{
		{scope: binding.ScopeFolders, want: []string{"z zoom", "/ search"}},
		{scope: binding.ScopeTags, want: []string{"z zoom", "/ search"}},
		{scope: binding.ScopeSnippetList, want: []string{"z zoom", "/ search"}},
		{scope: binding.ScopeSnippetPane, want: []string{"z zoom", "/ search"}},
	}

	for _, tt := range zoomHints {
		t.Run("hints zoom before search in "+string(tt.scope), func(t *testing.T) {
			t.Parallel()

			keys := keysWith(
				bound{scope: binding.ScopeGlobal, name: binding.Zoom, keys: []string{"z"}},
				bound{scope: binding.ScopeGlobal, name: binding.Search, keys: []string{"/"}},
			)

			got := keys.For(tt.scope).ShortHelp()

			assert.Equal(t, tt.want, hintTexts(got))
		})
	}

	t.Run("Tags hints open, new Tag, rename and delete before search", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeGlobal, name: binding.Open, keys: []string{"enter"}},
			bound{scope: binding.ScopeGlobal, name: binding.Search, keys: []string{"/"}},
			bound{scope: binding.ScopeTags, name: binding.NewTag, keys: []string{"N"}},
			bound{scope: binding.ScopeTags, name: binding.Rename, keys: []string{"r"}},
			bound{scope: binding.ScopeTags, name: binding.Delete, keys: []string{"d"}},
		)

		got := keys.For(binding.ScopeTags).ShortHelp()

		assert.Equal(t, []string{"enter open", "N new Tag", "r rename", "d delete", "/ search"}, hintTexts(got))
	})

	t.Run("name input hints save and cancel", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeNameInput, name: binding.Accept, keys: []string{"enter"}},
			bound{scope: binding.ScopeNameInput, name: binding.Cancel, keys: []string{"esc"}},
		)

		got := keys.For(binding.ScopeNameInput).ShortHelp()

		assert.Equal(t, []string{"enter save", "esc cancel"}, hintTexts(got))
	})

	t.Run("editor hints save, cancel, field and Language", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeEditor, name: binding.Save, keys: []string{"ctrl+s"}},
			bound{scope: binding.ScopeEditor, name: binding.Cancel, keys: []string{"esc"}},
			bound{scope: binding.ScopeEditor, name: binding.NextField, keys: []string{"down"}},
			bound{scope: binding.ScopeEditor, name: binding.PickLanguage, keys: []string{"ctrl+g"}},
		)

		got := keys.For(binding.ScopeEditor).ShortHelp()

		assert.Equal(t, []string{"ctrl+s save", "esc cancel", "down field", "ctrl+g Language"}, hintTexts(got))
	})

	t.Run("picker hints move, pick and close", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopePicker, name: binding.Down, keys: []string{"ctrl+n"}},
			bound{scope: binding.ScopePicker, name: binding.Accept, keys: []string{"enter"}},
			bound{scope: binding.ScopePicker, name: binding.Cancel, keys: []string{"esc"}},
			bound{scope: binding.ScopePicker, name: binding.ShowAllLanguages, keys: []string{"ctrl+a"}},
		)

		got := keys.For(binding.ScopePicker).ShortHelp()

		assert.Equal(t, []string{"ctrl+n move", "enter pick", "esc close"}, hintTexts(got))
	})

	t.Run("Folders full help lists the Default Language Binding", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(bound{scope: binding.ScopeFolders, name: binding.Language, keys: []string{"L"}})

		got := keys.For(binding.ScopeFolders).FullHelp()

		assert.Equal(t, [][]string{{"L Default Language"}}, columnTexts(got))
	})

	for _, scope := range []binding.Scope{binding.ScopeFolders, binding.ScopeSnippetList} {
		t.Run(string(scope)+" full help lists the move Binding", func(t *testing.T) {
			t.Parallel()

			keys := keysWith(bound{scope: scope, name: binding.Move, keys: []string{"m"}})

			got := keys.For(scope).FullHelp()

			assert.Equal(t, [][]string{{"m move"}}, columnTexts(got))
		})
	}

	t.Run("Content hints save, leave, indent and dedent", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeContent, name: binding.Save, keys: []string{"ctrl+s"}},
			bound{scope: binding.ScopeContent, name: binding.Leave, keys: []string{"esc"}},
			bound{scope: binding.ScopeContent, name: binding.Indent, keys: []string{"ctrl+]"}},
			bound{scope: binding.ScopeContent, name: binding.Dedent, keys: []string{"ctrl+["}},
		)

		got := keys.For(binding.ScopeContent).ShortHelp()

		assert.Equal(t, []string{"ctrl+s save", "esc leave", "ctrl+] indent", "ctrl+[ dedent"}, hintTexts(got))
	})

	externalEditorHints := []struct {
		scope binding.Scope
		after string
		want  []string
	}{
		{scope: binding.ScopeSnippetList, after: binding.Edit, want: []string{"e edit", "E external editor"}},
		{scope: binding.ScopeSnippetPane, after: binding.Edit, want: []string{"e edit", "E external editor"}},
		{scope: binding.ScopeEditor, after: binding.PickLanguage, want: []string{"e Language", "E external editor"}},
		{scope: binding.ScopeContent, after: binding.Dedent, want: []string{"e dedent", "E external editor"}},
	}

	for _, tt := range externalEditorHints {
		t.Run("hints the external editor after "+tt.after+" in "+string(tt.scope), func(t *testing.T) {
			t.Parallel()

			keys := keysWith(
				bound{scope: tt.scope, name: binding.OpenInEditor, keys: []string{"E"}},
				bound{scope: tt.scope, name: tt.after, keys: []string{"e"}},
			)

			got := keys.For(tt.scope).ShortHelp()

			assert.Equal(t, tt.want, hintTexts(got))
		})
	}

	t.Run("disables a hint whose Binding has no keys", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeConfirm, name: binding.Yes, keys: []string{"y"}},
			bound{scope: binding.ScopeConfirm, name: binding.No, keys: []string{}},
		)

		got := keys.For(binding.ScopeConfirm).ShortHelp()

		assert.Equal(t, []string{"y yes"}, hintTexts(got))
	})

	t.Run("full help lists every row of the Scope with all its keys, as one column", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeFolders, name: binding.NewFolder, keys: []string{"N"}},
			bound{scope: binding.ScopeFolders, name: binding.Rename, keys: []string{"r", "f2"}},
			bound{scope: binding.ScopeFolders, name: binding.Delete, keys: []string{"d"}},
			bound{scope: binding.ScopeFolders, name: binding.Collapse, keys: []string{"space"}},
		)

		got := keys.For(binding.ScopeFolders).FullHelp()

		assert.Equal(t, [][]string{{"N new Folder", "r, f2 rename", "d delete", "space collapse"}}, columnTexts(got))
	})

	t.Run("full help leaves out a Binding with no keys", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeConfirm, name: binding.Yes, keys: []string{}},
			bound{scope: binding.ScopeConfirm, name: binding.No, keys: []string{"n", "esc"}},
		)

		got := keys.For(binding.ScopeConfirm).FullHelp()

		assert.Equal(t, [][]string{{"n, esc no"}}, columnTexts(got))
	})

	t.Run("matches any of a Binding's keys", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(bound{scope: binding.ScopeSearch, name: binding.Down, keys: []string{"down", "ctrl+n"}})

		assert.True(t, keys.For(binding.ScopeSearch).Matches(keypress.Ctrl('n'), binding.Down))
	})

	t.Run("matches only the Scope's own rows", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeGlobal, name: binding.Search, keys: []string{"/"}},
			bound{scope: binding.ScopeSnippetList, name: binding.Copy, keys: []string{"y"}},
		)

		assert.False(t, keys.For(binding.ScopeSnippetList).Matches(keypress.Letter('/'), binding.Search))
		assert.False(t, keys.For(binding.ScopeFolders).Matches(keypress.Letter('y'), binding.Copy))
	})

	t.Run("names a Binding's first key", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeContent, name: binding.OpenInEditor, keys: []string{"ctrl+o", "ctrl+e"}},
		)

		assert.Equal(t, "ctrl+o", keys.For(binding.ScopeContent).FirstKey(binding.OpenInEditor))
	})

	t.Run("names no key for an unbound Binding", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(bound{scope: binding.ScopeContent, name: binding.OpenInEditor, keys: []string{}})

		assert.Empty(t, keys.For(binding.ScopeContent).FirstKey(binding.OpenInEditor))
	})
}

func TestKeys_EmptyListHints(t *testing.T) {
	t.Parallel()

	keys := keysWith(
		bound{scope: binding.ScopeGlobal, name: binding.NewSnippet, keys: []string{"n"}},
		bound{scope: binding.ScopeGlobal, name: binding.Capture, keys: []string{"p"}},
		bound{scope: binding.ScopeGlobal, name: binding.Search, keys: []string{"/"}},
		bound{scope: binding.ScopeGlobal, name: binding.Help, keys: []string{}},
		bound{scope: binding.ScopeFolders, name: binding.NewFolder, keys: []string{"N"}},
	)

	got := keys.EmptyListHints()

	assert.Equal(t, []string{"n new Snippet", "p Capture", "/ Search", "N new Folder"}, hintTexts(got))
}

func TestKeys_HelpOverlayHints(t *testing.T) {
	t.Parallel()

	keys := keysWith(
		bound{scope: binding.ScopeGlobal, name: binding.Help, keys: []string{"f1", "?"}},
		bound{scope: binding.ScopeGlobal, name: binding.Back, keys: []string{"backspace"}},
	)

	got := keys.HelpOverlayHints()

	assert.Equal(t, []string{"f1 close", "backspace close"}, hintTexts(got))
}

func TestAlwaysShown(t *testing.T) {
	t.Parallel()

	hints := keysWith(
		bound{scope: binding.ScopeGlobal, name: binding.Help, keys: []string{"f1"}},
		bound{scope: binding.ScopeGlobal, name: binding.Search, keys: []string{"/"}},
	).For(binding.ScopeTags).ShortHelp()

	assert.False(t, binding.AlwaysShown(hints[len(hints)-2]), "search")
	assert.True(t, binding.AlwaysShown(hints[len(hints)-1]), "help")
}

func keysWith(bindings ...bound) binding.Keys {
	keys := make(binding.Keys)

	for _, each := range bindings {
		if keys[each.scope] == nil {
			keys[each.scope] = make(map[string][]string)
		}

		keys[each.scope][each.name] = each.keys
	}

	return keys
}

func hintTexts(hints []key.Binding) []string {
	texts := make([]string, 0, len(hints))

	for _, hint := range hints {
		if hint.Enabled() {
			texts = append(texts, hint.Help().Key+" "+hint.Help().Desc)
		}
	}

	return texts
}

func columnTexts(columns [][]key.Binding) [][]string {
	texts := make([][]string, 0, len(columns))

	for _, column := range columns {
		texts = append(texts, hintTexts(column))
	}

	return texts
}
