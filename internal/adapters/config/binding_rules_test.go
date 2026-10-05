package config_test

import (
	"bytes"
	"errors"
	"maps"
	"path/filepath"
	"slices"
	"testing"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
)

func TestLoad_BindingRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{
			name:     "rejects two Bindings in one Scope sharing a key",
			contents: "[bindings.snippet_list]\nedit = [\"y\"]\n",
			want:     `bindings.snippet_list.edit: "y" is also bound to bindings.snippet_list.copy`,
		},
		{
			name:     "rejects a global key that a pane Scope also binds",
			contents: "[bindings.snippet_list]\nedit = [\"/\"]\n",
			want:     `bindings.global.search: "/" is also bound to bindings.snippet_list.edit`,
		},
		{
			name:     "compares keys after reordering their modifiers",
			contents: "[bindings.global]\nhelp = [\"ctrl+shift+f\"]\nsearch = [\"shift+ctrl+f\"]\n",
			want:     `bindings.global.search: "ctrl+shift+f" is also bound to bindings.global.help`,
		},
		{
			name:     "rejects a printable key without a modifier in a text-entry Scope",
			contents: "[bindings.editor]\nsave = [\"s\"]\n",
			want:     `bindings.editor.save: "s" types text in editor; add a modifier`,
		},
		{
			name:     "rejects space without a modifier in a text-entry Scope",
			contents: "[bindings.picker]\naccept = [\"space\"]\n",
			want:     `bindings.picker.accept: "space" types text in picker; add a modifier`,
		},
		{
			name:     "rejects a printable key without a modifier in the name input",
			contents: "[bindings.name_input]\naccept = [\"a\"]\n",
			want:     `bindings.name_input.accept: "a" types text in name_input; add a modifier`,
		},
		{
			name:     "rejects binding ctrl+c",
			contents: "[bindings.confirm]\nno = [\"ctrl+c\"]\n",
			want:     `bindings.confirm.no: "ctrl+c" always quits and cannot be bound`,
		},
		{
			name:     "reports every bad key of a Binding by its key path",
			contents: "[bindings.snippet_list]\ncopy = [\"shift+y\", \"escape\"]\n",
			want: `bindings.snippet_list.copy: "shift+y" never matches; write "Y"` + "\n" +
				`bindings.snippet_list.copy: "escape" never matches; write "esc"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := config.Load(t.Context(), testOptions(writeConfig(t, tt.contents), &bytes.Buffer{}))

			invalid, ok := errors.AsType[config.InvalidError](err)
			require.True(t, ok, "want InvalidError, got %v", err)
			assert.EqualError(t, invalid.Problems, tt.want)
		})
	}
}

func TestLoad_BindingRulesAllow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		contents string
	}{
		{
			name:     "allows one key in two pane Scopes",
			contents: "[bindings.folders]\ndelete = [\"x\"]\n[bindings.snippet_list]\ndelete = [\"x\"]\n",
		},
		{
			name:     "allows a global key in a text-entry Scope",
			contents: "[bindings.search]\ncopy = [\"ctrl+d\"]\n",
		},
		{
			name:     "allows a modified printable key in a text-entry Scope",
			contents: "[bindings.editor]\nsave = [\"alt+s\"]\n",
		},
		{
			name:     "allows a printable key without a modifier in confirm",
			contents: "[bindings.confirm]\nyes = [\"j\"]\n",
		},
		{
			name:     "allows one key listed twice in a Binding",
			contents: "[bindings.snippet_list]\ncopy = [\"y\", \"y\"]\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loadValid(t, tt.contents)
		})
	}
}

func TestLoad_StoresKeysAsBubbleTeaWritesThem(t *testing.T) {
	t.Parallel()

	got := loadValid(t, "[bindings.global]\nsearch = [\"shift+ctrl+f\"]\n")

	assert.Equal(t, []string{"ctrl+shift+f"}, keyNames(got, config.ScopeGlobal, config.BindingSearch))
}

func FuzzLoadBindings(f *testing.F) {
	f.Add(uint8(0), uint8(0), "q", "ctrl+q")
	f.Add(uint8(3), uint8(1), "y", "ctrl+c")
	f.Add(uint8(5), uint8(4), "s", "space")
	f.Add(uint8(0), uint8(11), "shift+ctrl+f", "/")

	defaults := loadDefaults(f)

	f.Fuzz(func(t *testing.T, scopeIndex, bindingIndex uint8, first, second string) {
		scope := config.ScopeNames()[int(scopeIndex)%len(config.ScopeNames())]
		names := bindingNamesIn(defaults, scope)
		binding := names[int(bindingIndex)%len(names)]

		got, err := config.Load(
			t.Context(),
			testOptions(writeBindings(t, scope, binding, first, second), &bytes.Buffer{}),
		)
		if err != nil {
			assert.ErrorAs(t, err, new(config.InvalidError))

			return
		}

		assertBindingRulesHold(t, got)
	})
}

func loadDefaults(f *testing.F) config.Config {
	f.Helper()

	got, err := config.Load(f.Context(), testOptions(filepath.Join(f.TempDir(), configFileName), &bytes.Buffer{}))
	require.NoError(f, err)

	return got
}

func bindingNamesIn(cfg config.Config, scopeName string) []string {
	scope := config.Scope(scopeName)

	names := make([]string, 0, len(cfg.Bindings[scope]))
	for binding := range cfg.Bindings[scope] {
		names = append(names, binding.String())
	}

	slices.Sort(names)

	return names
}

func writeBindings(t *testing.T, scope, binding string, keys ...string) string {
	t.Helper()

	var contents bytes.Buffer
	require.NoError(t, toml.NewEncoder(&contents).Encode(map[string]map[string]map[string][]string{
		"bindings": {scope: {binding: keys}},
	}))

	return writeConfig(t, contents.String())
}

func assertBindingRulesHold(t *testing.T, cfg config.Config) {
	t.Helper()

	paneKeys := map[string]bool{}
	globalKeys := map[string]bool{}

	for _, scope := range slices.Sorted(maps.Keys(cfg.Bindings)) {
		scopeKeys := assertScopeRulesHold(t, scope, cfg.Bindings[scope])

		switch {
		case scope == config.ScopeGlobal:
			globalKeys = scopeKeys
		case isPaneScope(scope):
			maps.Copy(paneKeys, scopeKeys)
		}
	}

	for key := range globalKeys {
		assert.False(t, paneKeys[key], "global key %q is also bound in a pane Scope", key)
	}
}

func isPaneScope(scope config.Scope) bool {
	return slices.Contains(
		[]config.Scope{config.ScopeFolders, config.ScopeTags, config.ScopeSnippetList, config.ScopeSnippetPane},
		scope,
	)
}

func assertScopeRulesHold(t *testing.T, scope config.Scope, bindings map[config.Binding][]config.Key) map[string]bool {
	t.Helper()

	owners := map[string]config.Binding{}

	for binding, keys := range bindings {
		for _, key := range keys {
			name := key.String()
			reparsed, err := config.ParseKey(name)
			require.NoError(t, err)
			assert.Equal(t, key, reparsed)
			assert.NotEqual(t, "ctrl+c", name)
			assert.False(t, isTextEntryScope(scope) && typesText(name), "%q types text in %s", name, scope)

			owner, taken := owners[name]
			assert.False(t, taken && owner != binding, "%q is bound twice in %s", name, scope)
			owners[name] = binding
		}
	}

	return keySet(owners)
}

func isTextEntryScope(scope config.Scope) bool {
	return slices.Contains(
		[]config.Scope{config.ScopeEditor, config.ScopeContent, config.ScopeSearch, config.ScopePicker},
		scope,
	)
}

func typesText(name string) bool {
	return name == "space" || utf8.RuneCountInString(name) == 1
}

func keySet(owners map[string]config.Binding) map[string]bool {
	set := make(map[string]bool, len(owners))
	for name := range owners {
		set[name] = true
	}

	return set
}
