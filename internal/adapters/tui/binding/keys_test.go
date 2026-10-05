package binding_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/test/keypress"
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
			bound{scope: binding.ScopeSnippetList, name: binding.CycleSort, keys: []string{"o"}},
		)

		got := keys.For(binding.ScopeSnippetList).ShortHelp()

		assert.Equal(t, []string{"c Copy", "a new", "o sort", "ctrl+f search"}, hintTexts(got))
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

	t.Run("name input hints save and cancel", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeNameInput, name: binding.Accept, keys: []string{"enter"}},
			bound{scope: binding.ScopeNameInput, name: binding.Cancel, keys: []string{"esc"}},
		)

		got := keys.For(binding.ScopeNameInput).ShortHelp()

		assert.Equal(t, []string{"enter save", "esc cancel"}, hintTexts(got))
	})

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

	t.Run("disables a hint whose Binding has no keys", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeConfirm, name: binding.Yes, keys: []string{"y"}},
			bound{scope: binding.ScopeConfirm, name: binding.No, keys: []string{}},
		)

		got := keys.For(binding.ScopeConfirm).ShortHelp()

		assert.Equal(t, []string{"y yes"}, hintTexts(got))
	})

	t.Run("full help is the short hint as one column", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeConfirm, name: binding.Yes, keys: []string{"y"}},
			bound{scope: binding.ScopeConfirm, name: binding.No, keys: []string{"n"}},
		)
		scope := keys.For(binding.ScopeConfirm)

		assert.Equal(t, [][]key.Binding{scope.ShortHelp()}, scope.FullHelp())
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
