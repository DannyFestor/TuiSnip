package folderpath_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
)

func TestPaths_Full(t *testing.T) {
	t.Parallel()

	sample := foldertree.New(t)

	tests := []struct {
		name   string
		folder domain.FolderID
		want   string
	}{
		{name: "spells the Root", folder: domain.FolderID{}, want: "Root"},
		{name: "starts a top-level Folder at the Root", folder: sample.Go.ID(), want: "Root / go"},
		{name: "spells every ancestor of a nested Folder", folder: sample.Tests.ID(), want: "Root / go / testing"},
		{
			name:   "marks a Folder missing from the tree",
			folder: testkit.NewSequentialIDs().NewFolderID(),
			want:   "Root / …",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, folderpath.New(sample.Tree).Full(tt.folder))
		})
	}
}

func TestPaths_Short(t *testing.T) {
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

			assert.Equal(t, tt.want, folderpath.New(sample.Tree).Short(tt.folder))
		})
	}
}

func TestPaths_zeroValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Root", folderpath.Paths{}.Full(domain.FolderID{}))
}
