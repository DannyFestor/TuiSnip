package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/app/library"
)

func TestTrashLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		counts library.TrashCounts
		want   string
	}{
		{name: "an empty Trash prints nothing", counts: library.TrashCounts{Folders: 0, Snippets: 0}, want: ""},
		{
			name:   "leaves out a zero Folder count",
			counts: library.TrashCounts{Folders: 0, Snippets: 1},
			want:   "In the Trash: 1 Snippet\n",
		},
		{
			name:   "leaves out a zero Snippet count",
			counts: library.TrashCounts{Folders: 2, Snippets: 0},
			want:   "In the Trash: 2 Folders\n",
		},
		{
			name:   "counts both when both are held",
			counts: library.TrashCounts{Folders: 1, Snippets: 3},
			want:   "In the Trash: 1 Folder, 3 Snippets\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, trashLine(test.counts))
		})
	}
}
