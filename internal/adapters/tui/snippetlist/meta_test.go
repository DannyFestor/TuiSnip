package snippetlist_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetlist"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		specs []testkit.FragmentSpec
		want  string
	}{
		{name: "names the Language of a single Fragment", specs: []testkit.FragmentSpec{{Language: "Go"}}, want: "Go"},
		{
			name:  "counts the Fragments after the first one's Language",
			specs: []testkit.FragmentSpec{{Language: "Go"}, {Language: "YAML"}, {Title: "README"}},
			want:  "Go +2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snippet := testkit.Snippet(t, testkit.SnippetSpec{Fragments: tt.specs})

			assert.Equal(t, tt.want, snippetlist.Language(snippet))
		})
	}
}
