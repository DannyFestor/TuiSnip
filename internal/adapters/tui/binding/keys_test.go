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

	t.Run("name input hints save and cancel", func(t *testing.T) {
		t.Parallel()

		keys := keysWith(
			bound{scope: binding.ScopeNameInput, name: binding.Accept, keys: []string{"enter"}},
			bound{scope: binding.ScopeNameInput, name: binding.Cancel, keys: []string{"esc"}},
		)

		got := keys.For(binding.ScopeNameInput).ShortHelp()

		assert.Equal(t, []string{"enter save", "esc cancel"}, hintTexts(got))
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

	assert.False(t, binding.AlwaysShown(hints[0]), "search")
	assert.True(t, binding.AlwaysShown(hints[1]), "help")
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
