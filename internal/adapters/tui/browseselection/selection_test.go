package browseselection_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestSelection(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	folderID := ids.NewFolderID()
	tagID := ids.NewTagID()

	tests := []struct {
		name       string
		selection  browseselection.Selection
		wantFolder domain.FolderID
		inFolder   bool
		wantTag    domain.TagID
		withTag    bool
	}{
		{name: "the zero selection is the Root", selection: browseselection.Selection{}, inFolder: true},
		{name: "a Folder", selection: browseselection.InFolder(folderID), wantFolder: folderID, inFolder: true},
		{name: "a Tag", selection: browseselection.WithTag(tagID), wantTag: tagID, withTag: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotFolder, inFolder := tt.selection.Folder()
			gotTag, withTag := tt.selection.Tag()

			assert.Equal(t, tt.wantFolder, gotFolder)
			assert.Equal(t, tt.inFolder, inFolder)
			assert.Equal(t, tt.wantTag, gotTag)
			assert.Equal(t, tt.withTag, withTag)
		})
	}
}

func TestSelection_compares(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	tagID := ids.NewTagID()
	root := browseselection.InFolder(domain.FolderID{})
	tag := browseselection.WithTag(tagID)

	assert.Equal(t, browseselection.Selection{}, root)
	assert.Equal(t, browseselection.WithTag(tagID), tag)
	assert.NotEqual(t, browseselection.WithTag(ids.NewTagID()), tag)
}
