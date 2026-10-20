package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrashedSnippetsLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		count int
		want  []string
	}{
		{name: "an export without trashed Snippets prints nothing", count: 0, want: nil},
		{name: "counts one trashed Snippet", count: 1, want: []string{"Put 1 Snippet in the Trash"}},
		{name: "counts several trashed Snippets", count: 5, want: []string{"Put 5 Snippets in the Trash"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, trashedSnippetsLines(test.count))
		})
	}
}
