package binding_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestRows(t *testing.T) {
	t.Parallel()

	t.Run("every row has keys in the default config", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys

		for _, row := range binding.Rows() {
			assert.NotEmpty(t, keys[row.Scope][row.Name], "%s.%s", row.Scope, row.Name)
		}
	})

	t.Run("declares each Binding once", func(t *testing.T) {
		t.Parallel()

		seen := make(map[binding.Row]bool)

		for _, row := range binding.Rows() {
			declared := binding.Row{Scope: row.Scope, Name: row.Name, Label: ""}
			assert.False(t, seen[declared], "%s.%s", row.Scope, row.Name)
			seen[declared] = true
		}
	})

	t.Run("every hint in the default config has a key and a label", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys

		for _, scope := range []binding.Scope{
			binding.ScopeFolders, binding.ScopeTags, binding.ScopeSnippetList, binding.ScopeSnippetPane,
			binding.ScopeEditor, binding.ScopeContent, binding.ScopeSearch, binding.ScopeConfirm,
		} {
			for _, hint := range keys.For(scope).ShortHelp() {
				assert.True(t, hint.Enabled(), "%s", scope)
				assert.NotEmpty(t, hint.Help().Desc, "%s", scope)
			}
		}
	})
}
