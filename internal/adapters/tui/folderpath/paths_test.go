package folderpath_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
)

func TestPaths_Folder(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	sample := foldertree.New(t)

	tests := []struct {
		name   string
		folder domain.FolderID
		want   string
	}{
		{name: "spells the Root", folder: domain.FolderID{}, want: "Root"},
		{name: "starts a top-level Folder at the Root", folder: sample.Go.ID(), want: "Root / go"},
		{name: "spells every ancestor of a nested Folder", folder: sample.Tests.ID(), want: "Root / go / testing"},
		{name: "marks a Folder missing from the tree", folder: ids.NewFolderID(), want: "Root / …"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, folderpath.New(sample.Tree).Folder(tt.folder))
		})
	}
}

func TestPaths_Of(t *testing.T) {
	t.Parallel()

	sample := foldertree.New(t)
	snippet := testkit.Snippet(t, testkit.SnippetSpec{FolderID: sample.Tests.ID()})

	assert.Equal(t, "Root / go / testing", folderpath.New(sample.Tree).Of(snippet))
}

func TestPaths_Compact(t *testing.T) {
	t.Parallel()

	sample := foldertree.New(t)

	tests := []struct {
		name   string
		folder domain.FolderID
		want   string
	}{
		{name: "spells the Root", folder: domain.FolderID{}, want: "Root"},
		{name: "leaves the Root out of a Folder's path", folder: sample.Tests.ID(), want: "go / testing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snippet := testkit.Snippet(t, testkit.SnippetSpec{FolderID: tt.folder})

			assert.Equal(t, tt.want, folderpath.New(sample.Tree).Compact(snippet))
		})
	}
}

func TestPaths_zeroValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Root", folderpath.Paths{}.Of(testkit.Snippet(t, testkit.SnippetSpec{})))
}
